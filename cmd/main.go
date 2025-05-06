package main

import (
	"log"
	"os"

	"github.com/SafroO-Ov/Project_YADRO/cmd/internal"
)

func main() {
	config, err := internal.LoadConfig("config.json")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	events, err := internal.LoadEvents("events")
	if err != nil {
		log.Fatalf("Failed to load events: %v", err)
	}

	logFile, err := os.Create("output.log")
	if err != nil {
		log.Fatalf("Failed to create output log: %v", err)
	}
	defer logFile.Close()

	logger := log.New(logFile, "", 0)
	processor := internal.NewProcessor(config, logger)
	processor.Process(events)

	err = processor.WriteResults("result.txt")
	if err != nil {
		log.Fatalf("Failed to write results: %v", err)
	}
}
