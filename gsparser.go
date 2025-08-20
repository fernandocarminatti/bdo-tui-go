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
}

type Profile struct {
	FamilyInfo FamilyInfo `json:"FamilyInfo"`
}

func parseGearscore(papd string) (int, bool) {
	papd = strings.TrimSpace(papd)
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
	if len(os.Args) < 1 {
		log.Fatalf("Usage: %s <folder>")
	}
	folder := os.Args[1]

	var values []int
	var usersWithPrivateData int

	err := filepath.Walk(folder, func(path string, info os.FileInfo, err error) error {
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

		gearscore, validValue := parseGearscore(profile.FamilyInfo.PAPD)
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

	average := sum / count
	fmt.Printf("Read from %d users\n", count)
	fmt.Printf("Average gearscore: %d\n", average)
	fmt.Printf("Users with Private Data: %d\n", usersWithPrivateData)
}
