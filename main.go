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

	// read the file contents in, line by line
	file, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	keyedRaces := make(map[string][]string)
	re := regexp.MustCompile("(/[/\\w\\.-]+:[0-9]+)(\\s|$)")
	foundRace := false
	var currRace []string
	// process each line to find data race sections
	frs := false // found read section?
	fws := false // found write section?
	readKey := ""
	writeKey := ""

	for scanner.Scan() {
		line := scanner.Text()
		if foundRace {
			currRace = append(currRace, line)
			// look for end of race section
			if strings.Contains(line, "==================") {
				foundRace = false
			} else if frs || fws { // we are matching either a read-by or write-by stack
				if match := re.FindString(line); len(match) > 0 {
					if frs { // read-by
						readKey = match
						frs = false
					} else { // write-by
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
			// reset flags
			foundRace = true
			frs = false
			fws = false
			readKey = ""
			writeKey = ""
			currRace = make([]string, 0)
			currRace = append(currRace, "==================", "WARNING: DATA RACE")
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	for _, race := range keyedRaces {
		for _, line := range race {
			fmt.Println(line)
		}
	}
}
