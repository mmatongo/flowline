package outline

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/mmatongo/flowline/pkg/config"
	"github.com/mmatongo/flowline/pkg/logger"
)

type Collection struct {
	Name string `json:"name"`
	ID   string `json:"id"`
	URL  string `json:"url"`
}

func GetCollections(a *logger.App) (string, error) {
	cfg := config.NewConfig()
	url := fmt.Sprintf("%s/collections.list", cfg.BaseURL)

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		a.Logger.Errorf("error making document request, %s", err)
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+cfg.APIKey)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := cfg.Client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response body")
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("error status code %v", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return "", fmt.Errorf("error unmarshalling json %W", err)
	}

	data, ok := result["data"].([]interface{})
	if !ok {
		return "", fmt.Errorf("'data' field not found in response")
	}

	var collections []Collection
	for _, item := range data {
		if collection, ok := item.(map[string]interface{}); ok {
			name, _ := collection["name"].(string)
			id, _ := collection["id"].(string)
			url, _ := collection["url"].(string)
			collections = append(collections, Collection{
				Name: name,
				ID:   id,
				URL:  cfg.BaseURL + url,
			})
		}
	}

	output, err := json.MarshalIndent(collections, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal collections to JSON: %w", err)
	}

	return string(output), nil
}
