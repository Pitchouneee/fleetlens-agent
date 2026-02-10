package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Pitchouneee/fleetlens-agent/apiclient"
	"github.com/Pitchouneee/fleetlens-agent/collector"
)

func main() {
	apiURL := flag.String("api", "", "FleetLens API base URL (e.g. http://localhost:8080)")
	interval := flag.Duration("interval", 24*time.Hour, "Interval between data collection cycles (e.g. 30m, 1h, 24h)")
	flag.Parse()

	if *apiURL == "" {
		*apiURL = os.Getenv("FLEETLENS_API_URL")
	}

	if *apiURL == "" {
		log.Fatal("API URL is required: use -api flag or set FLEETLENS_API_URL environment variable")
	}

	client := apiclient.NewClient(*apiURL)

	log.Printf("FleetLens agent started (interval: %s)", *interval)

	// Run immediately on startup, then on each tick
	collectAndSend(client)

	ticker := time.NewTicker(*interval)
	defer ticker.Stop()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case <-ticker.C:
			collectAndSend(client)
		case s := <-sig:
			log.Printf("Received %s, shutting down", s)
			return
		}
	}
}

func collectAndSend(client *apiclient.Client) {
	info := collector.Collect()

	data, _ := json.MarshalIndent(info, "", "  ")
	log.Printf("Collected system info:\n%s", data)

	if err := client.SendSystemInfo(info); err != nil {
		log.Printf("ERROR: failed to send system info: %v", err)
		return
	}

	log.Println("System info sent successfully")
}
