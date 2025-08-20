package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"time"
	"net/http"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type FamilyInfo struct {
	Name          string
	CreationDate  string
	Guild         string
	PAPD          string
	Energy        string
	Contribution  string
}

type LifeSkill struct {
	Name   string
	LevelName  string
	LevelInt string
	Mastery string
}

type Character struct {
	Name   string
	Class  string
	Level  string
}

type FullProfile struct {
	FamilyInfo FamilyInfo
	LifeSkills []LifeSkill
	Characters []Character
}

func getDocument(url string) (*goquery.Document, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0")

	client := &http.Client{Timeout: 10 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("server returned status: %s", res.Status)
	}

	return goquery.NewDocumentFromReader(res.Body)
}

func fetchProfileData(profileURL string) (*FullProfile, error) {
	doc, err := getDocument(profileURL)
	if err != nil {
		return nil, err
	}

	var profile FullProfile

	familyInfoBox := doc.Find("div.profile_detail")
	profile.FamilyInfo.Name = strings.TrimSpace(familyInfoBox.Find("div.nick_wrap > p.nick").Text())
	profile.FamilyInfo.CreationDate = strings.TrimSpace(familyInfoBox.Find("ul.line_list > li:nth-child(1) > span.desc").Text())
	profile.FamilyInfo.Guild = strings.TrimSpace(familyInfoBox.Find("ul.line_list > li > span.guild > a").Text())
	profile.FamilyInfo.PAPD = strings.TrimSpace(familyInfoBox.Find("ul.line_list > li:nth-child(3) > span.desc").Text())
	profile.FamilyInfo.Energy = strings.TrimSpace(familyInfoBox.Find("ul.line_list > li:nth-child(4) > span.desc").Text())
	profile.FamilyInfo.Contribution = strings.TrimSpace(familyInfoBox.Find("ul.line_list > li:nth-child(5) > span.desc").Text())

	lifeskillsBox := doc.Find("ul.character_data_box")
	lifeskillsBox.Find("li").Each(func(i int, s *goquery.Selection) {
		name := strings.TrimSpace(s.Find("span.spec_name").Text())
		rawLevel := strings.TrimSpace(s.Find("span.spec_level").Text())
		normalizedLevel := strings.Replace(rawLevel, "Nv.", " Nv.", 1)
		normalizedSplit := strings.SplitN(normalizedLevel, " ", 2)
		levelName, levelInt := "", ""
		if len(normalizedSplit) == 2 {
			levelName = normalizedSplit[0]
			levelInt = normalizedSplit[1]
		}
		mastery := strings.TrimSpace(s.Find("span.spec_stat").Text())

		profile.LifeSkills = append(profile.LifeSkills, LifeSkill{
			Name:      name,
			LevelName: levelName,
			LevelInt:  levelInt,
			Mastery:   mastery,
		})
	})

	charactersBox := doc.Find("ul.character_list")
	charactersBox.Find("li").Each(func(i int, s *goquery.Selection) {
		isMain := strings.Contains(s.Find("p.character_name span.selected_label").Text(), "Personagem Principal")
		name := strings.TrimSpace(
			s.Find("p.character_name").Contents().FilterFunction(func(i int, sel *goquery.Selection) bool {
				return goquery.NodeName(sel) == "#text"
			}).Text(),
		)

		class := strings.TrimSpace(s.Find("span.character_symbol em:nth-child(2)").Text())
		level := strings.TrimSpace(s.Find("span.character_info span:nth-child(2)").Text())
		if isMain {
			level += " Personagem Principal"
		}

		profile.Characters = append(profile.Characters, Character{
			Name:  name,
			Class: class,
			Level: level,
		})
	})

	return &profile, nil
}

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("Usage: %s <input.csv>", os.Args[0])
	}
	csvPath := os.Args[1]

	csvFile, err := os.Open(csvPath)
	if err != nil {
		log.Fatal(err)
	}
	defer csvFile.Close()

	reader := csv.NewReader(csvFile)
	records, err := reader.ReadAll()
	if err != nil {
		log.Fatal(err)
	}

	for i, record := range records {
		if i == 0 && record[0] == "Nickname" {
			continue // skip header
		}
		if len(record) < 2 {
			log.Printf("Skipping malformed row: %v", record)
			continue
		}

		nickname := record[0]
		url := record[1]

		profile, err := fetchProfileData(url)
		if err != nil {
			log.Printf("[%s] error: %v", nickname, err)
			continue
		}

		outFile := fmt.Sprintf("%s.json", nickname)
		data, _ := json.MarshalIndent(profile, "", "  ")
		if err := os.WriteFile(outFile, data, 0644); err != nil {
			log.Printf("[%s] write error: %v", nickname, err)
			continue
		}

		log.Printf("[%s] written to %s", nickname, outFile)
	}
}
