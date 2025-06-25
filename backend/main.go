// backend/main.go
package main

import (
    "fmt"
    "log"
    "net/http"
)

func main() {
    // Define a simple handler function
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprintln(w, "Hello from the Go Backend!")
    })

    // Start the server on port 8080
    port := ":8080"
    log.Printf("Backend server starting on port %s\n", port)
    if err := http.ListenAndServe(port, nil); err != nil {
        log.Fatal(err)
    }
}