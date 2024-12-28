package chaospost

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"

	"git.mills.io/prologic/go-gopher"
	"github.com/PuerkitoBio/goquery"
	"github.com/tunnelchaos/go-packages/config"
	"github.com/tunnelchaos/go-packages/gopherhelpers"
	"github.com/tunnelchaos/go-packages/helpers"
)

const (
	postURL = "https://office.c3post.de/"
)

type csrfToken struct {
	Token   string
	Expires time.Time
}

type csrfTokenRequest struct {
	Action string `json:"action"`
}

type csrfTokenResponse struct {
	Comment string    `json:"_comment"`
	Expires time.Time `json:"expires"`
	Ok      bool      `json:"ok"`
	Token   string    `json:"token"`
}

type sentMessageResponse struct {
	Code   string              `json:"code"`
	Errors map[string][]string `json:"errors"`
	Ok     bool                `json:"ok"`
}

type trackResponse struct {
	Code      string    `json:"code"`
	Delivered time.Time `json:"delivered"`
	History   []struct {
		Event     string    `json:"event"`
		State     string    `json:"state"`
		Timestamp time.Time `json:"timestamp"`
	} `json:"history"`
	State   string    `json:"state"`
	Updated time.Time `json:"updated"`
}

func (t *trackResponse) toGopher() string {
	indent := len("Delivered: ")
	response := gopherhelpers.FormatInfoForGophermap(indent, "Code: ", t.Code)
	if !t.Delivered.IsZero() {
		response += gopherhelpers.FormatInfoForGophermap(indent, "Delivered: ", t.Delivered.String())
	}
	response += gopherhelpers.FormatInfoForGophermap(indent, "State: ", t.State)
	response += gopherhelpers.FormatInfoForGophermap(indent, "Updated: ", t.Updated.Format(time.RFC3339))
	response += gopherhelpers.CreateGopherInfo("History:")
	for _, h := range t.History {
		response += gopherhelpers.CreateGopherInfo("State " + h.State + " happened at " + h.Timestamp.Format(time.RFC3339))
	}
	return response
}

var currentToken csrfToken

func getCSRFToken(s config.Secrets, client *http.Client) (string, error) {
	if currentToken.Token != "" && currentToken.Expires.After(time.Now()) {
		return currentToken.Token, nil
	}
	var accesstoken string
	var ok bool
	if accesstoken, ok = s["chaosposttoken"]; !ok {
		log.Println("No chaospost access token found")
		return "", errors.New("No chaospost access token found")
	}
	fmt.Println("Access token", accesstoken)
	request := csrfTokenRequest{
		Action: "write",
	}
	response := csrfTokenResponse{}
	jsonData, err := json.Marshal(request)
	if err != nil {
		log.Println("Error marshalling request", err)
		return "", errors.New("Error marshalling chaospost API request")
	}
	req, err := http.NewRequest("POST", postURL+"/api/v1/request_csrf_token", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Println("Error creating request", err)
		return "", errors.New("Error creating chaospost API request")
	}
	req.Header.Set("Authorization", accesstoken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Tunnelrat Gopher Client")
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Error getting", postURL, err)
		return "", errors.New("Error getting chaospost API")
	}
	defer resp.Body.Close()
	rawData, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Failed to read response body: %v", err)
		return "", errors.New("Error reading chaospost API")
	}
	decoder := json.NewDecoder(bytes.NewReader(rawData))
	err = decoder.Decode(&response)
	if err != nil {
		log.Println("Error decoding: ", err, string(rawData))
		return "", errors.New("Error decoding chaospost API")
	}
	if !response.Ok {
		return "", errors.New("Error getting CSRF token " + string(rawData))
	}
	currentToken.Token = response.Token
	currentToken.Expires = response.Expires
	return response.Token, nil

}

func findEvent(event string) (string, error) {
	client := helpers.CreateHttpClient()
	resp, err := client.Get(postURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", errors.New("Error getting event list")
	}
	doc, err := goquery.NewDocumentFromReader(resp.Body)

	result := ""
	// Find the select element with id "to_event_id" and its options
	doc.Find("#to_event_id option").Each(func(index int, option *goquery.Selection) {
		// Get the value and text of each option
		value, _ := option.Attr("value")
		text := option.Text()
		text = strings.ToLower(strings.TrimSpace(text))
		event = strings.ToLower(strings.TrimSpace(event))
		if text == event {
			result = value
		}
	})
	if result == "" {
		return "", errors.New("Event not found")
	}
	return result, nil
}

func buildNewSelector(selectors []string, arg string) string {
	newselector := ""
	for _, selector := range selectors {
		if selector == "" {
			continue
		}
		newselector += "/" + selector
	}
	newselector += "/" + arg
	return newselector
}

func generateResponse(response string) string {
	return gopherhelpers.CreateGopherInfo(response)
}

func writeResponse(w gopher.ResponseWriter, response string) {
	w.Write([]byte(generateResponse(response)))
}

func handleSend(w gopher.ResponseWriter, r *gopher.Request, s config.Secrets, event string, host string, port int) {
	selectors, split := gopherhelpers.SplitRequest(r.Selector)
	option := false
	arg := ""
	if len(split) > 0 {
		option = true
		arg = split[0]
	}
	response := ""
	switch len(selectors) {
	case 4:
		//First stage, we are getting the sender
		newselector := buildNewSelector(selectors, arg)
		response = gopherhelpers.CreateGopherEntry(gopher.INDEXSEARCH, "Sender recieved. Please enter recipient now", newselector, host, port)
	case 5:
		//Second stage, we are getting the recipient
		if !option {
			writeResponse(w, "Invalid request: Missing recipient")
			return
		}
		newselector := buildNewSelector(selectors, arg)
		response = gopherhelpers.CreateGopherEntry(gopher.INDEXSEARCH, "Recipient recieved. Please enter message now", newselector, host, port)
	case 6:
		//Third stage, we are getting the message
		if !option {
			writeResponse(w, "Invalid request: Missing message")
			return
		}
		client := helpers.CreateHttpClient()
		jar, err := cookiejar.New(nil)
		if err != nil {
			writeResponse(w, "Error creating cookie jar")
			log.Println("Error creating cookie jar", err)
			return
		}
		client.Jar = jar
		token, err := getCSRFToken(s, client)
		if err != nil {
			writeResponse(w, "Error getting CSRF token")
			log.Println("Error getting CSRF token", err)
			return
		}
		eventID, err := findEvent(event)
		if err != nil {
			writeResponse(w, "Error finding event")
			return
		}
		form := url.Values{}
		form.Add("csrf_token", token)
		form.Add("from_event_id", "-1")
		form.Add("to_event_id", eventID)
		form.Add("sender", selectors[4])
		form.Add("receiver", selectors[5])
		form.Add("message", arg)
		form.Add("postcard_id", "-1")
		form.Add("check", "1")
		req, err := http.NewRequest("POST", postURL+"?json", strings.NewReader(form.Encode()))
		if err != nil {
			writeResponse(w, "Error creating request "+err.Error())
			log.Println("Error creating request", err)
			return
		}
		req.Header.Set("X-CSRFToken", token)
		req.Header.Set("X-CSRF-Token", token)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("User-Agent", "Tunnelrat Gopher Client")
		resp, err := client.Do(req)
		if err != nil {
			writeResponse(w, "Error sending message")
			log.Println("Error sending message", err)
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			writeResponse(w, "Error sending message")
			log.Println("Error sending message", resp.StatusCode)
			return
		}
		rawData, err := io.ReadAll(resp.Body)
		if err != nil {
			writeResponse(w, "Error reading response")
			log.Println("Error reading response", err)
			return
		}
		decoder := json.NewDecoder(bytes.NewReader(rawData))
		var sentresponse sentMessageResponse
		err = decoder.Decode(&sentresponse)
		if err != nil {
			writeResponse(w, "Error decoding response")
			log.Println("Error decoding response", err, string(rawData))
			return
		}
		if !sentresponse.Ok {
			for key, value := range sentresponse.Errors {
				response += key + ": " + strings.Join(value, ", ") + ";"
			}
			response = generateResponse(response)
		} else {
			response = "Message sent. Code: " + sentresponse.Code
			response = generateResponse(response)
		}

	}
	fmt.Println(response)
	w.Write([]byte(response))

}

func handleTrack(w gopher.ResponseWriter, r *gopher.Request) {
	_, split := gopherhelpers.SplitRequest(r.Selector)
	if len(split) < 1 {
		writeResponse(w, "Invalid request: Missing tracking number")
	}
	trackingNumber := split[0]
	client := helpers.CreateHttpClient()
	resp, err := client.Get(postURL + "/track/" + trackingNumber + "?json")
	if err != nil {
		writeResponse(w, "Error getting tracking information")
		log.Println("Error getting tracking information", err)
		return
	}
	defer resp.Body.Close()
	rawData, err := io.ReadAll(resp.Body)
	if err != nil {
		writeResponse(w, "Error reading response")
		log.Println("Error reading response", err)
		return
	}
	decoder := json.NewDecoder(bytes.NewReader(rawData))
	var trackresponse trackResponse
	err = decoder.Decode(&trackresponse)
	if err != nil {
		writeResponse(w, "Error decoding response")
		log.Println("Error decoding response", err, string(rawData))
		return
	}
	response := trackresponse.toGopher()
	w.Write([]byte(response))
}

func Handler(w gopher.ResponseWriter, r *gopher.Request, s config.Secrets, host string, port int) {
	selectors, _ := gopherhelpers.SplitRequest(r.Selector)
	if len(selectors) < 4 {
		w.Write([]byte("Invalid request: Missing event or selector"))
		return
	}
	mode := strings.TrimSpace(strings.ToLower(selectors[2]))
	event := strings.TrimSpace(strings.ToLower(selectors[3]))
	switch mode {
	case "send":
		handleSend(w, r, s, event, host, port)
	case "track":
		handleTrack(w, r)
	default:
		w.Write([]byte("Invalid request: Unknown mode"))
	}
}
