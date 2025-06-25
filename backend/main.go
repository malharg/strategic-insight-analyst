// backend/main.go (Modified)
package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/malharg/strategic-insight-analyst/backend/auth"
	"github.com/malharg/strategic-insight-analyst/backend/database"
)

func main() {
	database.InitDB("./sia.db")
	auth.InitFirebaseAuth() // Initialize Firebase Admin

	// Public endpoint
	http.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]bool{"ok": true})
	})

	// Create a new ServeMux for protected routes
	mux := http.NewServeMux()
	mux.HandleFunc("/api/secure-ping", securePingHandler)

	// Wrap the mux with the AuthMiddleware
	protectedRoutes := auth.AuthMiddleware(mux)

	// Register the protected routes handler
	http.Handle("/api/secure-ping", protectedRoutes)

	port := ":8080"
	log.Printf("Backend server starting on port %s\n", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatal(err)
	}
}

// This is our new protected handler
func securePingHandler(w http.ResponseWriter, r *http.Request) {
	// Retrieve the user ID from the context
	userID := r.Context().Value(auth.UserIDKey).(string)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Pong! You are authenticated.",
		"userID":  userID,
	})
}
