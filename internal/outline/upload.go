package outline

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mmatongo/flowline/internal/confluence"
	"github.com/mmatongo/flowline/pkg/logger"
	"github.com/mmatongo/flowline/utils"
	"golang.org/x/net/html"
)

func PrepareAndProcess(inputPath, outputPath, collectionID string, verify bool, a *logger.App) error {
	if err := os.MkdirAll(outputPath, os.ModePerm); err != nil {
		a.Logger.Errorf("failed to create output directory: %v", err)
		return err
	}

	htmlContent, err := os.ReadFile(filepath.Join(inputPath, "index.html"))
	if err != nil {
		a.Logger.Errorf("failed to read index.html: %v", err)
		return err
	}

	doc, err := html.Parse(strings.NewReader(string(htmlContent)))
	if err != nil {
		a.Logger.Errorf("failed to parse HTML: %v", err)
		return err
	}

	pages := confluence.ProcessHTML(doc)
	return processPages(pages, inputPath, outputPath, collectionID, verify, a, "")
}

func processPages(pages []*confluence.Page, inputPath, outputPath, collectionID string, verify bool, a *logger.App, parentID string) error {
	for _, page := range pages {
		documentID, err := processAndUploadFile(page.Title, filepath.Join(inputPath, page.URL), outputPath, collectionID, verify, a, parentID)
		if err != nil {
			a.Logger.Errorf("error processing file %s: %v", page.URL, err)
			continue
		}

		if len(page.Children) > 0 {
			err = processPages(page.Children, inputPath, outputPath, collectionID, verify, a, documentID)
			if err != nil {
				a.Logger.Errorf("error processing children of %s: %v", page.Title, err)
			}
		}
	}
	return nil
}

func processAndUploadFile(title, inputPath, outputPath, collectionID string, verify bool, a *logger.App, parentID string) (string, error) {
	htmlContent, err := os.ReadFile(inputPath)
	if err != nil {
		return "", err
	}

	processedHTML, err := uploadAndReplaceAttachments(string(htmlContent), filepath.Dir(inputPath), a)
	if err != nil {
		return "", err
	}

	_, markdownContent, err := utils.ConvertHTMLToMarkdown(processedHTML, a)
	if err != nil {
		return "", err
	}

	if verify {
		a.Print("markdown content for: ", inputPath)
		fmt.Println(strings.Repeat("=", 50))
		fmt.Println(markdownContent)
		fmt.Println(strings.Repeat("=", 50))

		fmt.Print("do you want to proceed with uploading this document? (y/n): ")
		var userInput string
		fmt.Scanln(&userInput)
		if strings.ToLower(userInput) != "y" {
			a.Print("skipping this document.")
			return "", nil
		}
	}

	document, err := createDocument(title, markdownContent, collectionID, parentID, a)
	if err != nil {
		return "", err
	}

	documentID, ok := document["id"].(string)
	if !ok {
		return "", fmt.Errorf("invalid document Id")
	}

	a.Logger.Printf("successfully created document: %s with Id: %s", title, documentID)

	outputFilePath := filepath.Join(outputPath, strings.TrimSuffix(filepath.Base(inputPath), ".html")+".md")
	if err := os.WriteFile(outputFilePath, []byte(markdownContent), 0644); err != nil {
		return "", err
	}

	a.Print("processed and uploaded: ", inputPath)
	return documentID, nil
}
