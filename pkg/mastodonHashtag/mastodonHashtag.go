package mastodonHashtag

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"git.mills.io/prologic/go-gopher"
	"github.com/tunnelchaos/go-packages/config"
	"github.com/tunnelchaos/go-packages/gopherhelpers"
	"github.com/tunnelchaos/go-packages/helpers"
	"golang.org/x/net/html"
)

const (
	// MastodonAPIURL is the URL of the Mastodon API
	mastodonAPIURL = "https://chaos.social/api/v2/search?q=%s&type=statuses&limit=42"
)

func extractText(n *html.Node, linkList []string, linkCounter int) (string, []string, int) {
	if n.Type == html.TextNode {
		// Get the text from text nodes
		return n.Data, linkList, linkCounter
	}
	if n.Type == html.ElementNode && (n.Data == "script" || n.Data == "style") {
		// Skip <script> and <style> content
		return "", linkList, linkCounter
	}
	if n.Data == "a" {
		// Check if the class contains "mention"
		for _, attr := range n.Attr {
			if attr.Key == "class" && strings.Contains(attr.Val, "mention") {
				// Skip processing for mention links
				var buf bytes.Buffer
				for c := n.FirstChild; c != nil; c = c.NextSibling {
					childText, updatedList, updatedCounter := extractText(c, linkList, linkCounter)
					buf.WriteString(childText)
					linkList, linkCounter = updatedList, updatedCounter
				}
				return buf.String(), linkList, linkCounter
			}
		}

		// Extract href attribute
		href := ""
		for _, attr := range n.Attr {
			if attr.Key == "href" {
				href = attr.Val
				break
			}
		}

		if href != "" {
			linkCounter++
			linkList = append(linkList, href)
			return fmt.Sprintf("[%d]", linkCounter), linkList, linkCounter
		}
	}

	var buf bytes.Buffer
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		childText, updatedList, updatedCounter := extractText(c, linkList, linkCounter)
		buf.WriteString(childText)
		linkList, linkCounter = updatedList, updatedCounter
	}

	// Add spacing for certain elements to preserve readability
	if n.Type == html.ElementNode {
		switch n.Data {
		case "p", "br":
			buf.WriteString("\n")
		case "h1", "h2", "h3", "h4", "h5", "h6":
			headerText := strings.TrimSpace(buf.String())
			buf.Reset()
			buf.WriteString(headerText)
			buf.WriteString("\n" + strings.Repeat("=", len(headerText)) + "\n")
		case "li":
			buf.WriteString("- ")
		}
	}

	return buf.String(), linkList, linkCounter
}

// ConvertHTMLToText converts HTML content to plain text
func convertHTMLToText(htmlContent string) (string, []string, error) {
	// Parse the HTML
	var linkList []string
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return "", linkList, err
	}

	// Extract text
	text, finallist, _ := extractText(doc, linkList, -1)

	// Clean up extra whitespace
	text = strings.TrimSpace(text)
	text = strings.ReplaceAll(text, "\n\n\n", "\n\n") // Collapse excessive newlines
	return text, finallist, nil
}

func (s Statuses) toGopher() string {
	ident := len("Created at: ")
	result := gopherhelpers.CreateGopherInfo(gopherhelpers.FormatForGopherMap(ident, "Nickname:", s.Account.Acct))
	result += gopherhelpers.CreateGopherInfo(gopherhelpers.FormatForGopherMap(ident, "Name:", s.Account.DisplayName))
	result += gopherhelpers.CreateGopherInfo(gopherhelpers.FormatForGopherMap(ident, "Created at:", s.CreatedAt.Format(time.DateTime)))
	result += gopherhelpers.CreateGopherURL("Status", s.URL, "tunnelchaos.net", 7070)
	content, urllist, err := convertHTMLToText(s.Content)
	if err != nil {
		log.Println("Error converting HTML to text", err)
		content = s.Content
	}
	result += gopherhelpers.FormatInfoForGophermap(ident, "Content:", content)
	for i, link := range urllist {
		result += gopherhelpers.CreateGopherURL(fmt.Sprintf("[%d]", i), link, "tunnelchaos.net", 7070)
	}

	result += gopherhelpers.CreateGopherInfo(gopherhelpers.CreateMaxLine("-"))
	return result
}

func SearchMastodonHashtag(hashtag string, s config.Secrets) string {
	log.Println("Searching Mastodon for hashtag", hashtag)
	netClinet := helpers.CreateHttpClient()
	url := fmt.Sprintf(mastodonAPIURL, hashtag)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Println("Error creating request", err)
		return gopherhelpers.CreateGopherInfo("Error creating Mastodon API request")
	}
	req.Header.Set("User-Agent", "Tunnelrat Gopher Client")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	var accesstoken string
	var ok bool
	if accesstoken, ok = s["mastodontoken"]; !ok {
		log.Println("No Mastodon access token found")
		return gopherhelpers.CreateGopherInfo("No Mastodon access token found")
	}
	authtoken := fmt.Sprintf("Bearer %s", accesstoken)
	req.Header.Set("Authorization", authtoken)
	resp, err := netClinet.Do(req)
	if err != nil {
		log.Println("Error getting", url, err)
		return gopherhelpers.CreateGopherInfo("Error getting Mastodon API")
	}
	defer resp.Body.Close()
	var result searchResult
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		log.Println("Error decoding", err)
		return gopherhelpers.CreateGopherInfo("Error decoding Mastodon API")
	}
	response := gopherhelpers.CreateGopherInfo("Toots with Hashtags:")
	response += gopherhelpers.CreateGopherInfo(gopherhelpers.CreateMaxLine("-"))
	for _, status := range result.Statuses {
		response += status.toGopher()
	}
	return response

}

func Handler(w gopher.ResponseWriter, r *gopher.Request, s config.Secrets) {
	split := strings.Split(r.Selector, "\t")
	selectors := strings.Split(split[0], "/")
	log.Println("Mastodon Split", split)
	log.Println("Mastodon Selectors", selectors)
	if len(selectors) < 2 {
		w.Write([]byte("Invalid request: Missing event or selector"))
		return
	}
	hashtag := selectors[2]
	w.Write([]byte(SearchMastodonHashtag(hashtag, s)))
}
