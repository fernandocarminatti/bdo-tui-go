package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type FamilyInfo struct {
	PAPD string `json:"PAPD"`
	Guild string `json:"Guild"`
}

type Profile struct {
	FamilyInfo FamilyInfo `json:"FamilyInfo"`
}

func parseUserData(userData Profile) (int, bool) {
	papd := strings.TrimSpace(userData.FamilyInfo.PAPD)
	if papd == "Privado" || papd == "" {
		return 0, false
	}
	num, err := strconv.Atoi(papd)
	if err != nil {
		return 0, false
	}
	return num, true
}

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("Usage: %s <folder>", os.Args[0])
	}
	folder := os.Args[1]

	var values []int
	var usersWithPrivateData int
	absFolder, err := filepath.Abs(folder)
	guildName := filepath.Base(absFolder)

	if err != nil {
		log.Fatal("Could not detemine guild name from absolute path")
	}

	err = filepath.Walk(folder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || filepath.Ext(path) != ".json" {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		var profile Profile
		if err := json.Unmarshal(data, &profile); err != nil {
			return err
		}

		gearscore, validValue := parseUserData(profile)
		if validValue {
			values = append(values, gearscore)
		} else {
			usersWithPrivateData++
		}

		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	if len(values) == 0 {
		log.Println("No gearscore values found")
		return
	}

	var sum int
	count := 0

	for _, val := range values {
		sum += val
		count++
	}

	if count == 0 {
		fmt.Println("No valid gearscore values found")
		return
	}

	totalGuildMembers := count + usersWithPrivateData
	average := sum / count
	fmt.Printf("Parsing for guild: %s\n", guildName)
	fmt.Printf("Found %d members\n", totalGuildMembers)
	fmt.Printf("Users with Private Data: %d\n", usersWithPrivateData)
	fmt.Printf("Average gearscore: %d\n", average)
}
