package outline

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/mmatongo/flowline/pkg/config"
	"github.com/mmatongo/flowline/pkg/logger"
	"github.com/mmatongo/flowline/rate"
)

type DocumentPayload struct {
	Title        string `json:"title"`
	Text         string `json:"text"`
	CollectionId string `json:"collectionId"`
	Publish      bool   `json:"publish"`
}

func createDocument(title, text, collectionID string, a *logger.App) (map[string]interface{}, error) {
	cfg := config.NewConfig()
	rate.LimitRequest(a)
	url := fmt.Sprintf("%s/documents.create", cfg.BaseURL)

	payload := DocumentPayload{
		Title:        title,
		Text:         text,
		CollectionId: collectionID,
		Publish:      true,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		a.Logger.Errorf("error marshalling document payload, %s", err)
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		a.Logger.Errorf("error making document request, %s", err)
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+cfg.APIKey)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := cfg.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		a.Logger.Errorf("failed to create document, %s", resp.Status)
		return nil, fmt.Errorf("failed to create document: %s", resp.Status)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	a.Print("successfully created document: ", title)
	return result["data"].(map[string]interface{}), nil
}
