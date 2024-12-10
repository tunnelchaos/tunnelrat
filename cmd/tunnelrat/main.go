package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"

	"github.com/go-ldap/ldap/v3"
	"github.com/tunnelchaos/hopger/pkg/helpers"
)

const (
	// LDAP server address
	ldapServer = "ldap://guru3.eventphone.de"
)

func main() {
	// Start listening on port 7070
	listener, err := net.Listen("tcp", ":7070")
	if err != nil {
		log.Fatalf("Error starting server: %v\n", err)
	}
	defer listener.Close()

	fmt.Println("Server is listening on port 7070...")

	for {
		// Accept a new connection
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Error accepting connection: %v\n", err)
			continue
		}
		fmt.Println("New connection established")

		// Handle the connection in a separate goroutine
		go handleConnection(conn)
	}
}

func fillLine(line string, length int) string {
	for len(line) < length {
		line += " "
	}
	return line
}

func formatResponse(entries []*ldap.Entry) string {
	response := "Search results:\n"
	response += helpers.CreateMaxLine("-") + "\n"
	indent := len("Location: ")
	for _, entry := range entries {
		response += fillLine("Name:", indent) + entry.GetAttributeValue("cn") + "\n"
		response += fillLine("Number:", indent) + entry.GetAttributeValue("sn") + "\n"
		response += fillLine("Location:", indent) + entry.GetAttributeValue("l") + "\n"
		response += helpers.CreateMaxLine("-") + "\n"
	}
	return response
}

func searchLDAP(l *ldap.Conn, base, search string) string {
	fmt.Println("Searching", base, search)
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
	fmt.Println("Entries found", len(sr.Entries))
	if len(sr.Entries) == 0 {
		return "No entries found"
	}
	response := formatResponse(sr.Entries)
	response = fmt.Sprintf("i%s\tfake\t(NULL)\t0", response)
	return response
}

func generateResponse(event string, selector string, search string) string {
	fmt.Println("Generating response", event, selector, search)
	baseDN := fmt.Sprintf("ou=%s,dc=eventphone,dc=de", event)
	l, err := ldap.DialURL(ldapServer)
	if err != nil {
		log.Printf("Error writing to connection: %v\n", err)
		return ""
	}
	defer l.Close()
	fmt.Println("Connected to LDAP server")
	selector = strings.ToLower(selector)
	switch selector {
	case "number":
		return searchLDAP(l, baseDN, fmt.Sprintf("(cn=%s)", search))
	case "user":
		return searchLDAP(l, baseDN, fmt.Sprintf("(sn=*%s*)", search))
	case "phonebook":
		return searchLDAP(l, baseDN, "(sn=*)")
	}
	return ""
}

func handleConnection(conn net.Conn) {
	defer conn.Close() // Ensure the connection is closed when done

	// Read data from the connection
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		line := scanner.Text()
		split := strings.Split(line, "\t")
		selectors := strings.Split(split[0], "/")
		fmt.Println(line)
		fmt.Println("Split", split)
		fmt.Println("Selectors", selectors)
		if len(selectors) == 3 {
			fmt.Println("Correct selectors")
			event := selectors[1]
			selector := selectors[2]
			search := "all"
			if len(split) > 1 {
				search = split[1]
			}
			response := generateResponse(event, selector, search)
			_, err := conn.Write([]byte(response))
			if err != nil {
				log.Printf("Error writing to connection: %v\n", err)
				break
			}
		}
		break
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Error reading from connection: %v\n", err)
	}
}
