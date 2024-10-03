package outline

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mmatongo/flowline/convert"
	"github.com/mmatongo/flowline/pkg/logger"
)

func PrepareAndProcess(inputPath, outputPath, collectionID string, verify bool, a *logger.App) error {
	// bleh
	if err := os.MkdirAll(outputPath, os.ModePerm); err != nil {
		a.Logger.Errorf("failed to create output directory: %v\n", err)
		return err
	}

	files, err := os.ReadDir(inputPath)
	if err != nil {
		a.Logger.Errorf("failed to read input directory: %v\n", err)
		return err
	}

	for _, file := range files {
		if filepath.Ext(file.Name()) == ".html" {
			input := filepath.Join(inputPath, file.Name())
			output := filepath.Join(outputPath, strings.TrimSuffix(file.Name(), ".html")+".md")

			if err := processAndUploadFile(input, output, collectionID, verify, a); err != nil {
				a.Logger.Errorf("error processing file %s: %v\n", file.Name(), err)
			}
		}
	}

	return nil
}

func processAndUploadFile(inputPath, outputPath, collectionID string, verify bool, a *logger.App) error {
	htmlContent, err := os.ReadFile(inputPath)
	if err != nil {
		return err
	}

	processedHTML, err := uploadAndReplaceAttachments(string(htmlContent), filepath.Dir(inputPath), a)
	if err != nil {
		return err
	}

	title, markdownContent, err := convert.ConvertHTMLToMarkdown(processedHTML, a)
	if err != nil {
		return err
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
			return nil
		}

	}

	document, err := createDocument(title, markdownContent, collectionID, a)
	if err != nil {
		return err
	}

	documentID, ok := document["id"].(string)
	if !ok {
		return fmt.Errorf("invalid document Id")
	}

	a.Logger.Printf("successfully created document: %s with Id: %s\n", title, documentID)

	if err := os.WriteFile(outputPath, []byte(markdownContent), 0644); err != nil {
		return err
	}

	a.Print("processed and uploaded: ", inputPath)
	return nil
}
