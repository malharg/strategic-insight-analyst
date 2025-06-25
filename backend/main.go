// backend/main.go
package main

import (
	"fmt"
	"log"
	"net/http"

	// Import the new database package
	"github.com/malharg/strategic-insight-analyst/backend/database"
)

func main() {
	// Initialize the database
	database.InitDB("./sia.db") // This will create a sia.db file in the backend folder

	// Define a simple handler function
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello from the Go Backend! Database is connected.")
	})

	// Start the server on port 8080
	port := ":8080"
	log.Printf("Backend server starting on port %s\n", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatal(err)
	}
}
