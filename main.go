package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/pquerna/otp/totp"
	"golang.design/x/clipboard"
)

type TOTPEntry struct {
	Name   string `json:"name"`
	Secret string `json:"secret"`
}

type TOTPData struct {
	Entries []TOTPEntry `json:"entries"`
}

const dataFile = "totp.json"

func main() {
	if len(os.Args) < 2 {
		fmt.Print(`Authinator CLI Help Guide

Usage: authinator [command] [arguments...]

Commands:
  create [name] [secret]   Create a new TOTP entry with the given name and secret.
                           Example: authinator create my_account JBSWY3DPEHPK3PXP

  list                     List all stored TOTP entries with their current codes and time remaining.
                           Example: authinator list

  [name]                   Get the current TOTP code for the entry with the specified name.
                           Also shows the time remaining until the next code.
                           Example: authinator my_account

  remove [name]            Remove the TOTP entry with the specified name.
                           Example: authinator remove my_account

  serve                    Start an HTTP server on port 8055 to manage TOTP entries via REST API.
                           Example: authinator serve

  help                     Display this help guide.

Detailed Guide:

1. Creating a New TOTP Entry:
   - You can create a new TOTP entry by providing a name and a secret key.
   - The secret key is typically provided by the service you are setting up 2FA for.
   - If you don’t have a secret key, you can usually generate a QR code and scan it using the CLI.

   Example: 
   authinator create github JBSWY3DPEHPK3PXP

   This will create a TOTP entry named 'github' using the secret key provided.

2. Listing All TOTP Entries:
   - Use the 'list' command to view all stored TOTP entries.
   - The list will display each entry's current code and the time remaining until the code expires.

   Example:
   authinator list

3. Retrieving a TOTP Code:
   - Simply run the command with the entry name to get the current TOTP code.
   - The output will include the code and the time remaining until it changes.
   - The code will also be copied to your clipboard automatically.

   Example:
   authinator github

4. Removing a TOTP Entry:
   - Use the 'remove' command to delete a TOTP entry by its name.

   Example:
   authinator remove github

5. Serving the Authinator via HTTP:
   - The 'serve' command starts an HTTP server on port 8055.
   - You can interact with your TOTP entries via REST API calls.
   - The following endpoints are available:
     - GET /totps: List all TOTP entries.
     - GET /totps/{name}: Get the current TOTP code for the specified entry.
     - POST /totps: Create a new TOTP entry by sending a JSON payload.
     - DELETE /totps/{name}: Delete a TOTP entry.

   Example:
   authinator serve

   Then you can use curl or any HTTP client to interact with the service:
   
   - List all entries:
     curl -X GET http://localhost:8055/totps
   
   - Create a new entry:
     curl -X POST -H "Content-Type: application/json" -d '{"name":"example","secret":"SECRETKEY"}' http://localhost:8055/totps

   - Get the TOTP code for an entry:
     curl -X GET http://localhost:8055/totps/example

   - Delete an entry:
     curl -X DELETE http://localhost:8055/totps/example
`)
		return
	}

	command := os.Args[1]

	switch command {
	case "create":
		if len(os.Args) == 4 {
			createEntry(os.Args[2], os.Args[3])
		} else {
			createEntryInteractive()
		}
	case "list":
		listEntries()
	case "remove":
		if len(os.Args) == 3 {
			removeEntry(os.Args[2])
		} else {
			fmt.Println("Usage: authinator remove [name]")
		}
	case "serve":
		startServer()
	default:
		if len(os.Args) == 2 {
			getCode(os.Args[1])
		} else {
			fmt.Println("Usage: authinator [command] [arguments...]")
		}
	}
}

// HTTP Handlers
func startServer() {
	http.HandleFunc("/totps", handleTOTPRequests)
	http.HandleFunc("/totps/", handleTOTPRequestsByID)

	fmt.Println("Serving on http://0.0.0.0:8055")
	log.Fatal(http.ListenAndServe("0.0.0.0:8055", nil))
}

func handleTOTPRequests(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		listEntriesHTTP(w, r)
	case "POST":
		createEntryHTTP(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleTOTPRequestsByID(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(r.URL.Path, "/totps/")

	switch r.Method {
	case "GET":
		getCodeHTTP(w, r, name)
	case "DELETE":
		removeEntryHTTP(w, r, name)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// HTTP-specific functions

func listEntriesHTTP(w http.ResponseWriter, r *http.Request) {
	data := loadData()

	json.NewEncoder(w).Encode(data.Entries)
}

func createEntryHTTP(w http.ResponseWriter, r *http.Request) {
	var entry TOTPEntry
	err := json.NewDecoder(r.Body).Decode(&entry)
	if err != nil || entry.Name == "" || entry.Secret == "" {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	createEntry(entry.Name, entry.Secret)
	fmt.Fprintf(w, "TOTP entry '%s' created successfully.\n", entry.Name)
}

func getCodeHTTP(w http.ResponseWriter, r *http.Request, name string) {
	data := loadData()

	for _, entry := range data.Entries {
		if entry.Name == name {
			// Generate the current TOTP code
			code, err := totp.GenerateCode(entry.Secret, time.Now())
			if err != nil {
				http.Error(w, "Error generating TOTP code", http.StatusInternalServerError)
				return
			}

			// Calculate time remaining in the current period
			remaining := 30 - (time.Now().Unix() % 30)

			response := map[string]interface{}{
				"code":       code,
				"expires_in": remaining,
			}
			json.NewEncoder(w).Encode(response)
			return
		}
	}

	http.Error(w, "No entry found with that name.", http.StatusNotFound)
}

func removeEntryHTTP(w http.ResponseWriter, r *http.Request, name string) {
	data := loadData()

	// Find the entry and remove it
	found := false
	newEntries := []TOTPEntry{}
	for _, entry := range data.Entries {
		if entry.Name != name {
			newEntries = append(newEntries, entry)
		} else {
			found = true
		}
	}

	if !found {
		http.Error(w, "No entry found with that name.", http.StatusNotFound)
		return
	}

	// Save the updated entries back to the JSON file
	data.Entries = newEntries
	saveData(data)

	fmt.Fprintf(w, "Entry '%s' has been removed.\n", name)
}

func createEntry(name, secret string) {
	data := loadData()

	for _, entry := range data.Entries {
		if entry.Name == name {
			fmt.Println("Entry with this name already exists.")
			return
		}
	}

	data.Entries = append(data.Entries, TOTPEntry{Name: name, Secret: secret})
	saveData(data)
	fmt.Println("Entry created successfully!")
}

func createEntryInteractive() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter name: ")
	name, _ := reader.ReadString('\n')
	name = strings.TrimSpace(name)

	fmt.Print("Enter TOTP secret: ")
	secret, _ := reader.ReadString('\n')
	secret = strings.TrimSpace(secret)

	createEntry(name, secret)
}

func listEntries() {
	data := loadData()

	if len(data.Entries) == 0 {
		fmt.Println("No entries found.")
		return
	}

	fmt.Println("Stored TOTP entries:")
	for _, entry := range data.Entries {
		// Generate the current TOTP code for each entry
		code, err := totp.GenerateCode(entry.Secret, time.Now())
		if err != nil {
			log.Printf("Error generating TOTP code for %s: %v", entry.Name, err)
			continue
		}

		// Calculate time remaining in the current period
		remaining := 30 - (time.Now().Unix() % 30)

		// Display the entry name, code, and time remaining
		fmt.Printf(" - %s: %s (expires in %d seconds)\n", entry.Name, code, remaining)
	}
}

func removeEntry(name string) {
	data := loadData()

	// Find the entry and remove it
	found := false
	newEntries := []TOTPEntry{}
	for _, entry := range data.Entries {
		if entry.Name != name {
			newEntries = append(newEntries, entry)
		} else {
			found = true
		}
	}

	if !found {
		fmt.Printf("No entry found with the name: %s\n", name)
		return
	}

	// Save the updated entries back to the JSON file
	data.Entries = newEntries
	saveData(data)

	fmt.Printf("Entry '%s' has been removed.\n", name)
}

func getCode(name string) {
	data := loadData()

	for _, entry := range data.Entries {
		if entry.Name == name {
			// Generate the current TOTP code
			currentTime := time.Now()
			code, err := totp.GenerateCode(entry.Secret, currentTime)
			if err != nil {
				log.Fatalf("Error generating current TOTP code: %v", err)
			}

			// Calculate time remaining in the current period
			remaining := 30 - (currentTime.Unix() % 30)
			fmt.Printf("Your current TOTP code is: %s (Time remaining: %d seconds)\n", code, remaining)

			// Generate the next TOTP code
			nextTime := currentTime.Add(time.Duration(remaining) * time.Second)
			nextCode, err := totp.GenerateCode(entry.Secret, nextTime)
			if err != nil {
				log.Fatalf("Error generating next TOTP code: %v", err)
			}
			fmt.Printf("After this, your next TOTP code will be: %s\n", nextCode)

			// Copy the current code to clipboard
			if err := clipboard.Write(clipboard.FmtText, []byte(code)); err != nil {
				log.Printf("Failed to copy code to clipboard: %v", err)
			} else {
				fmt.Println("Current code copied to clipboard.")
			}
			return
		}
	}

	fmt.Println("No entry found with that name.")
}

func loadData() TOTPData {
	data := TOTPData{}
	if _, err := os.Stat(dataFile); err == nil {
		file, err := os.Open(dataFile)
		if err != nil {
			log.Fatalf("Error reading data file: %v", err)
		}
		defer file.Close()

		content, err := io.ReadAll(file)
		if err != nil {
			log.Fatalf("Error reading file content: %v", err)
		}

		if err := json.Unmarshal(content, &data); err != nil {
			log.Fatalf("Error parsing data file: %v", err)
		}
	}
	return data
}

func saveData(data TOTPData) {
	file, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		log.Fatalf("Error saving data: %v", err)
	}
	err = os.WriteFile(dataFile, file, 0644)
	if err != nil {
		log.Fatalf("Error writing data file: %v", err)
	}
}
