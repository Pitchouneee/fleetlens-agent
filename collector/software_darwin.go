//go:build darwin

package collector

import (
	"encoding/json"
	"log"
	"os/exec"
)

type spApplicationsData struct {
	SPApplicationsDataType []spApp `json:"SPApplicationsDataType"`
}

type spApp struct {
	Name      string `json:"_name"`
	Version   string `json:"version"`
	ObtainedFrom string `json:"obtained_from"`
	LastModified string `json:"lastModified"`
}

func collectSoftware() []Software {
	cmd := exec.Command("system_profiler", "SPApplicationsDataType", "-json")
	output, err := cmd.Output()
	if err != nil {
		log.Printf("WARNING: failed to collect installed software: %v", err)
		return nil
	}

	var data spApplicationsData
	if err := json.Unmarshal(output, &data); err != nil {
		log.Printf("WARNING: failed to parse software list: %v", err)
		return nil
	}

	software := make([]Software, 0, len(data.SPApplicationsDataType))
	for _, app := range data.SPApplicationsDataType {
		if app.Name == "" {
			continue
		}
		software = append(software, Software{
			Name:        app.Name,
			Version:     app.Version,
			Publisher:   app.ObtainedFrom,
			InstalledAt: app.LastModified,
		})
	}

	return software
}
