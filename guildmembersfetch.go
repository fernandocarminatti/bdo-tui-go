package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"log"
	"encoding/csv"
	"path/filepath"
	
	"github.com/PuerkitoBio/goquery"
)

type GuildMember struct {
	Nickname string
	URL      string
}

func parseGuildMembers(html string) []GuildMember {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		log.Fatal(err)
	}

	uniqueMembers := make([]GuildMember, 0)
	seen := make(map[string]bool)

	doc.Find(`a[href*="profileTarget="]`).Each(func(i int, s *goquery.Selection) {
		nickname := strings.TrimSpace(s.Text())
		if nickname == "" {
			return
		}
		if !seen[nickname] {
			memberUrl, exists := s.Attr("href")
			if !exists {
				return
			}
			uniqueMembers = append(uniqueMembers, GuildMember{
				Nickname: nickname,
				URL:      memberUrl,
			})
			seen[nickname] = true
		}
	})

	return uniqueMembers
}

func persistToCsv(members []GuildMember, guildName string) error {
	if err := os.MkdirAll(guildName, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	fileName := fmt.Sprintf("%s_members.csv", guildName)
	filePath := filepath.Join(guildName, fileName)

	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	header := []string{"Nickname", "Ref"}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("failed to write header to csv: %w", err)
	}

	for _, member := range members {
		record := []string{member.Nickname, member.URL}
		if err := writer.Write(record); err != nil {
			return fmt.Errorf("failed to write record to csv: %w", err)
		}
	}
	return nil
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run guildmembersfetch.go <guild name>")
		os.Exit(1)
	}

	guildName := os.Args[1]
	
	baseURL := "https://www.sa.playblackdesert.com/Adventure/Guild/GuildProfile"
	params := url.Values{}
	params.Add("guildName", guildName)
	params.Add("region", "SA")
	fullURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	resp, err := http.Get(fullURL)
	if err != nil {
		fmt.Printf("Error fetching URL: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		fmt.Printf("HTTP Error: %d\n", resp.StatusCode)
		os.Exit(1)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response body: %v\n", err)
		os.Exit(1)
	}

	members := parseGuildMembers(string(body))
	
	if err := persistToCsv(members, guildName); err != nil {
		log.Fatalf("Error saving members to CSV: %v", err)
	}
	fmt.Println("Fetched guild members for: ", guildName)
	fmt.Printf("Found %d members\n", len(members))
}