package outline

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/mmatongo/flowline/pkg/config"
	"github.com/mmatongo/flowline/pkg/logger"
	"github.com/mmatongo/flowline/rate"
	"github.com/mmatongo/flowline/utils"
)

type AttachmentPayload struct {
	Name        string `json:"name"`
	ContentType string `json:"contentType"`
	Size        int64  `json:"size"`
	Preset      string `json:"preset"`
}

func createAttachment(filePath string, a *logger.App) (map[string]interface{}, error) {
	if strings.ToLower(filepath.Ext(filePath)) == ".html" {
		a.Print("skipping HTML file: ", filePath)
		return nil, nil
	}

	rate.LimitRequest(a)

	cfg := config.NewConfig()
	url := fmt.Sprintf("%s/attachments.create", cfg.BaseURL)

	fileName := filepath.Base(filePath)
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		a.Logger.Errorf("failed to get file info: %v", err)
		return nil, err
	}

	fileSize := fileInfo.Size()
	mimeType := utils.GetMimeType(filePath)

	if mimeType == "" {
		a.Logger.Errorf("cannot determine mimetype of %v, skipping...", filePath)
		return nil, nil
	}

	payload := AttachmentPayload{
		Name:        fileName,
		ContentType: mimeType,
		Size:        fileSize,
		Preset:      "documentAttachment",
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		a.Logger.Errorf("failed to marshal payload: %v", err)
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		a.Logger.Errorf("failed to create request: %v", err)
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+cfg.APIKey)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := cfg.Client.Do(req)
	if err != nil {
		a.Logger.Errorf("failed to send request: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to create attachment: %s - %s", resp.Status, string(bodyBytes))
	}

	var result map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	data, ok := result["data"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid response format")
	}

	uploadURL, ok := data["uploadUrl"].(string)
	if !ok {
		return nil, fmt.Errorf("no upload URL provided in the response")
	}

	if strings.HasPrefix(uploadURL, "/") {
		uploadURL = cfg.BaseURL + strings.TrimPrefix(uploadURL, "/api")
	}

	form, ok := data["form"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid form data in response")
	}

	rate.LimitRequest(a)

	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	for key, value := range form {
		if err := writer.WriteField(key, fmt.Sprintf("%v", value)); err != nil {
			return nil, fmt.Errorf("failed to write form field: %w", err)
		}
	}

	part, err := writer.CreateFormFile("file", fileName)
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %w", err)
	}

	if _, err := io.Copy(part, file); err != nil {
		return nil, fmt.Errorf("failed to copy file content: %w", err)
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close multipart writer: %w", err)
	}

	uploadReq, err := http.NewRequest("POST", uploadURL, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create upload request: %w", err)
	}

	uploadReq.Header.Set("Authorization", "Bearer "+cfg.APIKey)
	uploadReq.Header.Set("Content-Type", writer.FormDataContentType())

	uploadResp, err := cfg.Client.Do(uploadReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send upload request: %w", err)
	}
	defer uploadResp.Body.Close()

	uploadBodyBytes, _ := io.ReadAll(uploadResp.Body)

	if uploadResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to upload attachment: %s - %s", uploadResp.Status, string(uploadBodyBytes))
	}

	return map[string]interface{}{
		"id":   data["attachment"].(map[string]interface{})["id"],
		"key":  form["key"],
		"name": fileName,
	}, nil
}

func getAttachmentURL(attachmentInfo map[string]interface{}) string {
	cfg := config.NewConfig()
	if key, ok := attachmentInfo["key"].(string); ok {
		return fmt.Sprintf("%s/files.get?key=%s", cfg.BaseURL, key)
	}
	return ""
}

func uploadAndReplaceAttachments(htmlContent, basePath string, a *logger.App) (string, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		return "", err
	}

	processElement := func(s *goquery.Selection, attr string) {
		src, exists := s.Attr(attr)
		if exists && strings.HasPrefix(src, "attachments/") {
			cleanSrc := utils.CleanPath(src)
			srcPath := filepath.Join(basePath, cleanSrc)

			if _, err := os.Stat(srcPath); err == nil {
				attachment, err := createAttachment(srcPath, a)
				if err != nil {
					a.Logger.Printf("failed to upload attachment %s. error: %v", srcPath, err)
					return
				}

				if attachment != nil {
					attachmentURL := getAttachmentURL(attachment)
					if attachmentURL != "" {
						s.SetAttr(attr, attachmentURL)
					} else {
						a.Logger.Printf("failed to get URL for attachment %s. keeping original reference.", srcPath)
					}
				}
			} else {
				a.Logger.Printf("attachment file not found: %s", srcPath)
			}
		}
	}

	doc.Find("img").Each(func(i int, s *goquery.Selection) {
		if s.HasClass("emoticon") {
			s.Remove()
		} else {
			processElement(s, "src")
		}
	})

	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		processElement(s, "href")
	})

	return doc.Html()
}
