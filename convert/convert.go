package convert

import (
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

	title := "Untitled Document"
	if titleElement := doc.Find("title").First(); titleElement.Length() > 0 {
		fullTitle := strings.TrimSpace(titleElement.Text())
		/*
			Confluence will append the space name to the title on export so
			split the title by ":" and take the last part
		*/
		titleParts := strings.Split(fullTitle, ":")
		if len(titleParts) > 1 {
			title = strings.TrimSpace(titleParts[len(titleParts)-1])
		} else {
			title = fullTitle
		}
	}

	// remove unnecessary elements
	doc.Find("head, script, style, #main-header, #footer").Remove()

	// extract main content
	contentElement := doc.Find("#main-content").First()
	if contentElement.Length() == 0 {
		contentElement = doc.Find("body").First()
	}

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
		plugin.TableCompat(),
	)

	markdown, err := converter.ConvertString(sanitizedHTML)
	if err != nil {
		a.Logger.Errorf("an error occured during the conversion, %v", err)
		return "", "", err
	}

	return title, markdown, nil
}
