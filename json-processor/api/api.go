package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

const apiEndpoint = "http://api:3000/api/v1"

func SendToApi(data []map[string]interface{}, dataType string) error {
	endpoints := map[string]string{
		"contract":                "/contracts",
		"councilor":               "/councilors",
		"frequency":               "/frequencies",
		"generalProductivity":     "/general-productivity",
		"proposition":             "/propositions",
		"propositionProductivity": "/proposition-productivity",
		"travelExpenses":          "/travel-expenses",
	}

	endpoint, exists := endpoints[dataType]
	if !exists {
		return fmt.Errorf("unknown data type: %s", dataType)
	}

	payload, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %v", err)
	}

	resp, err := http.Post(apiEndpoint+endpoint, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("failed to send data to API: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("API responded with status: %d", resp.StatusCode)
	}

	return nil
}
