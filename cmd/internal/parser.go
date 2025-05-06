package internal

import (
	"bufio"
	"encoding/json"
	"os"
	"strconv"
	"strings"
)

func LoadConfig(filename string) (*Config, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var cfg Config
	decoder := json.NewDecoder(f)
	if err := decoder.Decode(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func LoadEvents(filename string) ([]Event, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	var events []Event
	for scanner.Scan() {
		line := scanner.Text()
		timeStr := line[1:13]
		parts := strings.Fields(line[14:])

		eventID, _ := strconv.Atoi(parts[0])
		competitor, _ := strconv.Atoi(parts[1])
		extra := parts[2:]

		events = append(events, Event{
			RawLine:     line,
			Time:        timeStr,
			EventID:     eventID,
			Competitor:  competitor,
			ExtraParams: extra,
		})
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return events, nil
}
