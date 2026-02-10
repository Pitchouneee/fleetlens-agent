package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"

	"github.com/Pitchouneee/fleetlens-agent/apiclient"
	"github.com/Pitchouneee/fleetlens-agent/collector"
)

func main() {
	apiURL := flag.String("api", "", "FleetLens API base URL (e.g. http://localhost:8080)")
	flag.Parse()

	if *apiURL == "" {
		*apiURL = os.Getenv("FLEETLENS_API_URL")
	}

	if *apiURL == "" {
		log.Fatal("API URL is required: use -api flag or set FLEETLENS_API_URL environment variable")
	}

	info := collector.Collect()

	data, _ := json.MarshalIndent(info, "", "  ")
	log.Printf("Collected system info:\n%s", data)

	client := apiclient.NewClient(*apiURL)
	if err := client.SendSystemInfo(info); err != nil {
		log.Fatalf("Failed to send system info: %v", err)
	}

	log.Println("System info sent successfully")
}
