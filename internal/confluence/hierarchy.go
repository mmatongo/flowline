package confluence

import (
	"strings"

	"golang.org/x/net/html"
)

type Page struct {
	Title    string
	URL      string
	Children []*Page
}

func ProcessHTML(n *html.Node) []*Page {
	var pages []*Page

	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "ul" {
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				if c.Type == html.ElementNode && c.Data == "li" {
					page := processLI(c)
					if page != nil && page.Title != "Home" {
						pages = append(pages, page)
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(n)

	return pages
}

func processLI(n *html.Node) *Page {
	var page *Page

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode && c.Data == "a" {
			page = extractPageInfo(c)
		} else if c.Type == html.ElementNode && c.Data == "ul" {
			if page == nil {
				page = &Page{}
			}
			for child := c.FirstChild; child != nil; child = child.NextSibling {
				if child.Type == html.ElementNode && child.Data == "li" {
					childPage := processLI(child)
					if childPage != nil {
						page.Children = append(page.Children, childPage)
					}
				}
			}
		}
	}

	return page
}

func extractPageInfo(n *html.Node) *Page {
	var url, title string
	for _, a := range n.Attr {
		if a.Key == "href" {
			url = a.Val
			break
		}
	}
	title = strings.TrimSpace(extractText(n))
	return &Page{Title: title, URL: url}
}

func extractText(n *html.Node) string {
	if n.Type == html.TextNode {
		return n.Data
	}
	var text string
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		text += extractText(c)
	}
	return text
}
