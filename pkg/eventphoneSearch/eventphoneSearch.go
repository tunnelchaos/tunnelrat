package eventphoneSearch

import (
	"fmt"
	"log"
	"strings"

	"git.mills.io/prologic/go-gopher"
	"github.com/go-ldap/ldap/v3"
	"github.com/tunnelchaos/go-packages/gopherhelpers"
)

const (
	// LDAP server address
	ldapServer = "ldap://guru3.eventphone.de"
)

func entrytoGopher(entry *ldap.Entry, indent int) string {
	result := gopherhelpers.CreateGopherInfo(gopherhelpers.FormatForGopherMap(indent, "Name:", entry.GetAttributeValue("cn")))
	result += gopherhelpers.CreateGopherInfo(gopherhelpers.FormatForGopherMap(indent, "Number:", entry.GetAttributeValue("sn")))
	result += gopherhelpers.CreateGopherInfo(gopherhelpers.FormatForGopherMap(indent, "Location:", entry.GetAttributeValue("l")))
	result += gopherhelpers.CreateGopherInfo(gopherhelpers.CreateMaxLine("-"))
	return result
}

func formatResponse(entries []*ldap.Entry) string {
	response := gopherhelpers.CreateGopherInfo("Search results:")
	response += gopherhelpers.CreateGopherInfo(gopherhelpers.CreateMaxLine("-"))
	indent := len("Location: ")
	for _, entry := range entries {
		response += entrytoGopher(entry, indent)
	}
	return response
}

func searchLDAP(l *ldap.Conn, base, search string) string {
	log.Println("Searching", base, search)
	searchReq := ldap.NewSearchRequest(
		base, // The base dn to search
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		search,                    // The filter to apply
		[]string{"cn", "sn", "l"}, // A list attributes to retrieve
		nil,
	)
	sr, err := l.Search(searchReq)
	if err != nil {
		log.Printf("Error writing to connection: %v\n", err)
		return ""
	}
	log.Println("Entries found", len(sr.Entries))
	if len(sr.Entries) == 0 {
		return gopherhelpers.CreateGopherInfo("No entries found")
	}
	response := formatResponse(sr.Entries)
	return response
}

func generateResponse(event string, selector string, search string) string {
	log.Println("Generating response", event, selector, search)
	baseDN := fmt.Sprintf("ou=%s,dc=eventphone,dc=de", event)
	l, err := ldap.DialURL(ldapServer)
	if err != nil {
		log.Printf("Error writing to connection: %v\n", err)
		return ""
	}
	defer l.Close()
	log.Println("Connected to LDAP server")
	selector = strings.ToLower(selector)
	switch selector {
	case "number":
		return searchLDAP(l, baseDN, fmt.Sprintf("(cn=%s)", search))
	case "user":
		return searchLDAP(l, baseDN, fmt.Sprintf("(sn=*%s*)", search))
	case "phonebook":
		return searchLDAP(l, baseDN, "(sn=*)")
	}
	return "Unknown selector"
}

func Handler(w gopher.ResponseWriter, r *gopher.Request) {
	split := strings.Split(r.Selector, "\t")
	selectors := strings.Split(split[0], "/")
	if len(selectors) < 3 {
		w.Write([]byte("Invalid request: Missing event or selector"))
		return
	}
	search := ""
	if len(split) > 1 {
		search = split[1]
	}
	event := selectors[2]
	selector := selectors[3]
	log.Println("Searching Eventphone", event, selector, search)
	response := generateResponse(event, selector, search)
	w.Write([]byte(response))
}
