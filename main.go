package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
)

func main() {
	if len(os.Args) == 1 {
		fmt.Println("Provide a race file path for deduping.")
		os.Exit(-1)
	}

	lines := readLines(os.Args[1])
	races := parseRaces(lines)
	fmt.Printf("Initially, there were %v races\n", len(races))
	keyedRaces := dedupeRaces(races)
	fmt.Printf("After deduping, there are %d races.\n", len(keyedRaces))
	for _, race := range keyedRaces {
		for _, line := range race {
			fmt.Println(line)
		}
	}
}

func readLines(path string) []string {
	// read the file contents in, line by line
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	content := make([]string, 0)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		content = append(content, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	return content
}

func parseRaces(lines []string) [][]string {
	foundRace := false
	var currRace []string
	races := make([][]string, 0)
	// process each line to find data race sections
	for _, line := range lines {
		if foundRace {
			currRace = append(currRace, line)
			if strings.Contains(line, "==================") {
				foundRace = false
				races = append(races, currRace)
			}
		} else if strings.Contains(line, "WARNING: DATA RACE") {
			foundRace = true
			currRace = make([]string, 0)
			currRace = append(currRace, "==================", "WARNING: DATA RACE")
		}
	}
	return races
}

func dedupeRaces(races [][]string) map[string][]string {
	keyedRaces := make(map[string][]string)
	re := regexp.MustCompile("(/[/\\w\\.-]+:[0-9]+)(\\s|$)")

	for _, race := range races {
		frs := false // found read section?
		fws := false // found write section?
		readKey := ""
		writeKey := ""
		for _, line := range race {
			if frs || fws {
				if match := re.FindString(line); len(match) > 0 {
					if frs {
						readKey = match
						frs = false
					} else {
						writeKey = match
						fws = false
					}
				}
			} else if strings.Contains(line, "ead by goroutine") {
				frs = true
				fws = false
			} else if strings.Contains(line, "rite by goroutine") {
				frs = false
				fws = true
			}
		}
		if len(readKey) > 0 && len(writeKey) > 0 {
			key := readKey + "|" + writeKey
			if _, exists := keyedRaces[key]; !exists {
				keyedRaces[key] = race
			}
			continue
		} else {
			fmt.Printf("Could not find read and write key for:\n %v\n", race)
		}
	}
	return keyedRaces
}
