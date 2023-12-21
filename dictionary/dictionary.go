// dictionary.go
package dictionary

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
)

type Entry struct {
	Mot        string `json:"mot"`
	Definition string `json:"definition"`
}

type Dictionnaire struct {
	filePath string
	entries  []Entry
	mutex    sync.Mutex
}

func NewDictionnaire(filePath string) *Dictionnaire {
	return &Dictionnaire{
		filePath: filePath,
		entries:  nil,
	}
}

func (d *Dictionnaire) loadFromFile() error {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	fileData, err := ioutil.ReadFile(d.filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %v", err)
	}

	err = json.Unmarshal(fileData, &d.entries)
	if err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %v", err)
	}

	return nil
}

func (d *Dictionnaire) saveToFile() error {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	fileData, err := json.MarshalIndent(d.entries, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %v", err)
	}

	err = ioutil.WriteFile(d.filePath, fileData, 0644)
	if err != nil {
		return fmt.Errorf("failed to write file: %v", err)
	}

	return nil
}

func (d *Dictionnaire) AddEntryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var newEntry Entry
	err := json.NewDecoder(r.Body).Decode(&newEntry)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	d.mutex.Lock()
	defer d.mutex.Unlock()

	d.entries = append(d.entries, newEntry)

	err = d.saveToFile()
	if err != nil {
		http.Error(w, fmt.Sprintf("Internal server error: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}
func (d *Dictionnaire) RemoveEntryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Retrieve the 'mot' parameter from the URL query
	word := strings.TrimSpace(r.URL.Query().Get("mot"))

	// Check if 'mot' parameter is empty
	if word == "" {
		http.Error(w, "Bad Request: Missing 'mot' parameter in the URL query", http.StatusBadRequest)
		return
	}

	// Use a case-insensitive comparison for 'mot'
	word = strings.ToLower(word)

	d.mutex.Lock()
	defer d.mutex.Unlock()

	var updatedEntries []Entry
	found := false
	for _, entry := range d.entries {
		// Use a case-insensitive comparison for entry.Mot
		if strings.ToLower(entry.Mot) == word {
			found = true
		} else {
			updatedEntries = append(updatedEntries, entry)
		}
	}

	// If 'mot' not found in entries
	if !found {
		http.Error(w, fmt.Sprintf("Not Found: Mot '%s' not found in the dictionary", r.URL.Query().Get("mot")), http.StatusNotFound)
		return
	}

	// Update the entries without the removed entry
	d.entries = updatedEntries

	// Save the updated entries to the file
	err := d.saveToFile()
	if err != nil {
		http.Error(w, fmt.Sprintf("Internal Server Error: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (d *Dictionnaire) GetEntryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Retrieve the 'mot' parameter from the URL query
	word := strings.TrimSpace(r.URL.Query().Get("mot"))

	// Check if 'mot' parameter is empty
	if word == "" {
		http.Error(w, "Bad Request: Missing 'mot' parameter in the URL query", http.StatusBadRequest)
		return
	}

	// Use a case-insensitive comparison for 'mot'
	word = strings.ToLower(word)

	d.mutex.Lock()
	defer d.mutex.Unlock()

	// Debugging: Log the received 'mot' parameter
	fmt.Println("Received GET request with 'mot' parameter:", word)

	var definition string
	found := false
	for _, entry := range d.entries {
		// Debugging: Log the comparison details
		fmt.Printf("Comparing '%s' with '%s'\n", strings.ToLower(entry.Mot), word)

		// Use a case-insensitive comparison for entry.Mot
		if strings.ToLower(entry.Mot) == word {
			definition = entry.Definition
			found = true
			break
		}
	}

	// Debugging: Log whether the word was found
	fmt.Println("Word Found:", found)

	// If 'mot' not found in entries
	if !found {
		http.Error(w, fmt.Sprintf("Not Found: Mot '%s' not found in the dictionary", word), http.StatusNotFound)
		return
	}

	// Respond with the definition
	response, err := json.Marshal(map[string]string{"mot": word, "definition": definition})
	if err != nil {
		http.Error(w, "Failed to marshal data", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(response)
}

func (d *Dictionnaire) ListEntriesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	err := d.loadFromFile()
	if err != nil {
		http.Error(w, fmt.Sprintf("Internal server error: %v", err), http.StatusInternalServerError)
		return
	}

	response, err := json.Marshal(d.entries)
	if err != nil {
		http.Error(w, "Failed to marshal data", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(response)
}
