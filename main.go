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
		fmt.Println("Provide the path for a file containing Go races.")
		os.Exit(-1)
	}

	lines := readLines(os.Args[1])
	keyedRaces := parseRaces(lines)

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

func parseRaces(lines []string) map[string][]string {
	keyedRaces := make(map[string][]string)
	re := regexp.MustCompile("(/[/\\w\\.-]+:[0-9]+)(\\s|$)")
	count := 0
	foundRace := false
	var currRace []string
	// process each line to find data race sections
	frs := false // found read section?
	fws := false // found write section?
	readKey := ""
	writeKey := ""
	for _, line := range lines {
		if foundRace {
			currRace = append(currRace, line)
			if strings.Contains(line, "==================") {
				foundRace = false
			} else if frs || fws {
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
			if len(readKey) > 0 && len(writeKey) > 0 {
				key := readKey + "|" + writeKey
				if _, exists := keyedRaces[key]; !exists {
					keyedRaces[key] = currRace
				}
				frs = false
				fws = false
				continue
			}
		} else if strings.Contains(line, "WARNING: DATA RACE") {
			foundRace = true
			count++
			frs = false // found read section?
			fws = false // found write section?
			readKey = ""
			writeKey = ""
			currRace = make([]string, 0)
			currRace = append(currRace, "==================", "WARNING: DATA RACE")
		}
	}
	return keyedRaces
}
