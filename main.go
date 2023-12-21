// main.go
package main

import (
	"example/dictionnaire/dictionary"
	"fmt"
	"net/http"
)

func main() {
	filePath := "C:/Users/ziadf/OneDrive/Bureau/dictionnaire/entries.json"

	dict := dictionary.NewDictionnaire(filePath)

	// Configure HTTP routes
	http.HandleFunc("/add", dict.AddEntryHandler)
	http.HandleFunc("/get", dict.GetEntryHandler)
	http.HandleFunc("/remove", dict.RemoveEntryHandler)
	http.HandleFunc("/list", dict.ListEntriesHandler)

	// Start the HTTP server
	fmt.Println("Server is running on :8080")
	http.ListenAndServe(":8080", nil)
}
