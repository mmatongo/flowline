package markdown

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/mmatongo/flowline/internal/confluence"
	"github.com/mmatongo/flowline/pkg/logger"
	"github.com/mmatongo/flowline/utils"
	"golang.org/x/net/html"
)

func ExportToMarkdown(inputPath, outputPath string, verify bool, a *logger.App) error {
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

	processed := make(map[string]bool)
	pages := confluence.ProcessHTML(doc)
	return processMarkdownPages(pages, inputPath, outputPath, verify, a, "", processed)
}

func processMarkdownPages(pages []*confluence.Page, inputPath, outputPath string, verify bool, a *logger.App, currentPath string, processed map[string]bool) error {
	for _, page := range pages {
		if processed[page.URL] {
			continue
		}

		processed[page.URL] = true
		pagePath := filepath.Join(currentPath, sanitizeFilename(page.Title))
		fullOutputPath := filepath.Join(outputPath, pagePath)

		if err := os.MkdirAll(fullOutputPath, os.ModePerm); err != nil {
			a.Logger.Errorf("failed to create directory %s: %v", fullOutputPath, err)
			continue
		}

		err := processMarkdownFile(
			page.Title,
			filepath.Join(inputPath, page.URL),
			fullOutputPath,
			verify,
			a,
		)
		if err != nil {
			a.Logger.Errorf("error processing file %s: %v", page.URL, err)
			continue
		}

		if len(page.Children) > 0 {
			err = processMarkdownPages(page.Children, inputPath, outputPath, verify, a, pagePath, processed)
			if err != nil {
				a.Logger.Errorf("error processing children of %s: %v", page.Title, err)
			}
		}
	}
	return nil
}

func processMarkdownFile(_, inputPath, outputDir string, verify bool, a *logger.App) error {
	htmlContent, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %v", inputPath, err)
	}

	processedHTML, err := processAndCopyAttachments(string(htmlContent), filepath.Dir(inputPath), outputDir, a)
	if err != nil {
		return fmt.Errorf("failed to process attachments: %v", err)
	}

	_, markdownContent, err := utils.ConvertHTMLToMarkdown(processedHTML, a)
	if err != nil {
		return fmt.Errorf("failed to convert to markdown: %v", err)
	}

	if verify {
		a.Print("markdown content for: ", inputPath)
		fmt.Println(strings.Repeat("=", 50))
		fmt.Println(markdownContent)
		fmt.Println(strings.Repeat("=", 50))

		a.Print("do you want to proceed with saving this document? (y/n): ")
		var userInput string
		fmt.Scanln(&userInput)
		if strings.ToLower(userInput) != "y" {
			a.Print("skipping this document.")
			return nil
		}
	}

	dirName := filepath.Base(outputDir)

	outputPath := filepath.Join(outputDir, dirName+".md")
	if err := os.WriteFile(outputPath, []byte(markdownContent), 0644); err != nil {
		return fmt.Errorf("failed to write markdown file: %v", err)
	}

	a.Print("processed and saved: ", outputPath)
	return nil
}

func processAndCopyAttachments(htmlContent, sourcePath, outputDir string, a *logger.App) (string, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		return "", err
	}

	attachmentsDir := filepath.Join(outputDir, "attachments")

	processElement := func(s *goquery.Selection, attr string) error {
		src, exists := s.Attr(attr)
		if !exists || !strings.HasPrefix(src, "attachments/") {
			return nil
		}

		cleanSrc := utils.CleanPath(src)
		srcPath := filepath.Join(sourcePath, cleanSrc)

		if _, err := os.Stat(srcPath); err != nil {
			a.Logger.Printf("attachment file not found: %s", srcPath)
			return nil
		}

		if err := os.MkdirAll(attachmentsDir, os.ModePerm); err != nil {
			return fmt.Errorf("failed to create attachments directory: %v", err)
		}

		destPath := filepath.Join(attachmentsDir, filepath.Base(cleanSrc))
		if err := copyFile(srcPath, destPath); err != nil {
			return fmt.Errorf("failed to copy attachment: %v", err)
		}

		s.SetAttr(attr, fmt.Sprintf("attachments/%s", filepath.Base(cleanSrc)))
		return nil
	}

	doc.Find("img").Each(func(i int, s *goquery.Selection) {
		if s.HasClass("emoticon") {
			s.Remove()
		} else {
			if err := processElement(s, "src"); err != nil {
				a.Logger.Printf("error processing image: %v", err)
			}
		}
	})

	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		if err := processElement(s, "href"); err != nil {
			a.Logger.Printf("error processing link: %v", err)
		}
	})

	html, err := doc.Html()
	if err != nil {
		return "", fmt.Errorf("failed to generate HTML: %v", err)
	}

	return html, nil
}

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

func sanitizeFilename(filename string) string {
	invalid := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	result := filename

	for _, char := range invalid {
		result = strings.ReplaceAll(result, char, "_")
	}

	result = strings.Trim(result, " .")

	if result == "" {
		return "untitled"
	}

	return result
}
