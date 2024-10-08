package convert

import (
	"regexp"
	"strings"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/JohannesKaufmann/html-to-markdown/plugin"
	"github.com/PuerkitoBio/goquery"
	"github.com/microcosm-cc/bluemonday"

	"github.com/mmatongo/flowline/pkg/logger"
)

func ConvertHTMLToMarkdown(htmlContent string, a *logger.App) (string, string, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		a.Logger.Errorf("error creating a reader from the html content, %v", err)
		return "", "", err
	}

	title := extractTitle(doc)

	// remove unnecessary elements
	doc.Find("head, script, style, #main-header, #footer").Remove()

	// extract main content
	contentElement := doc.Find("#main-content").First()
	if contentElement.Length() == 0 {
		contentElement = doc.Find("body").First()
	}

	preProcessTables(contentElement)

	// convert the extracted content to markdown
	html, err := contentElement.Html()
	if err != nil {
		return "", "", err
	}

	// sanitize the html
	p := bluemonday.UGCPolicy()
	sanitizedHTML := p.Sanitize(html)

	// convert sanitized HTML to Markdown
	converter := md.NewConverter("", true, nil)
	converter.Use(
		plugin.ConfluenceAttachments(),
		plugin.ConfluenceCodeBlock(),
		plugin.GitHubFlavored(),
		plugin.TaskListItems(),
		plugin.Table(), // Not sure about this plugin to be honest.
		plugin.YoutubeEmbed(),
		confluenceTable(),
	)

	markdown, err := converter.ConvertString(sanitizedHTML)
	if err != nil {
		a.Logger.Errorf("an error occured during the conversion, %v", err)
		return "", "", err
	}

	markdown = postProcessMarkdown(markdown)

	return title, markdown, nil
}

func extractTitle(doc *goquery.Document) string {
	title := "Untitled Document"
	if titleElement := doc.Find("title").First(); titleElement.Length() > 0 {
		fullTitle := strings.TrimSpace(titleElement.Text())
		titleParts := strings.Split(fullTitle, ":")
		if len(titleParts) > 1 {
			title = strings.TrimSpace(titleParts[len(titleParts)-1])
		} else {
			title = fullTitle
		}
	}
	return title
}

func preProcessTables(s *goquery.Selection) {
	s.Find("table").Each(func(i int, table *goquery.Selection) {
		if table.Find("th, td").Length() > 2 {
			table.AddClass("complex-table")
		}
	})
}

func confluenceTable() md.Plugin {
	return func(c *md.Converter) []md.Rule {
		return []md.Rule{
			{
				Filter: []string{"table"},
				Replacement: func(content string, selec *goquery.Selection, opt *md.Options) *string {
					var rows [][]string
					maxCols := 0

					selec.Find("tr").Each(func(i int, tr *goquery.Selection) {
						var row []string
						tr.Find("th, td").Each(func(j int, cell *goquery.Selection) {
							text := strings.TrimSpace(cell.Text())
							row = append(row, text)
						})
						if len(row) > 0 {
							rows = append(rows, row)
							if len(row) > maxCols {
								maxCols = len(row)
							}
						}
					})

					if len(rows) == 0 || maxCols == 0 {
						return nil
					}

					markdown := "\n\n"
					for i, row := range rows {
						for j := 0; j < maxCols; j++ {
							if j < len(row) {
								markdown += "| " + row[j] + " "
							} else {
								markdown += "|  "
							}
						}
						markdown += "|\n"

						if i == 0 {
							for j := 0; j < maxCols; j++ {
								markdown += "|---"
							}
							markdown += "|\n"
						}
					}
					markdown += "\n"

					return &markdown
				},
			},
		}
	}
}

func postProcessMarkdown(markdown string) string {
	// remove extra newlines between table rows
	markdown = regexp.MustCompile(`\n{3,}`).ReplaceAllString(markdown, "\n\n")

	// remove empty table rows
	markdown = regexp.MustCompile(`(?m)^\|([|\s]*\|)+\s*$`).ReplaceAllString(markdown, "")

	// ensure consistent spacing in table cells
	re := regexp.MustCompile(`(?m)^\|(.*)\|$`)
	markdown = re.ReplaceAllStringFunc(markdown, func(match string) string {
		cells := strings.Split(match, "|")
		for i, cell := range cells {
			cells[i] = " " + strings.TrimSpace(cell) + " "
		}
		return strings.Join(cells, "|")
	})

	return markdown
}
