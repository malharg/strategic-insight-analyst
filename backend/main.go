// backend/main.go (Final Version for this step)
package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/malharg/strategic-insight-analyst/backend/auth"
	"github.com/malharg/strategic-insight-analyst/backend/database"
	"github.com/rs/cors" // Import the cors package
)

func main() {
	database.InitDB("./sia.db")
	auth.InitFirebaseAuth()

	// Main router
	mux := http.NewServeMux()

	// Public endpoint
	mux.HandleFunc("/api/health", healthCheckHandler)

	// Protected endpoint
	// We create a specific handler for secure routes that applies the auth middleware
	securePingHandler := http.HandlerFunc(securePingHandlerFunc)
	mux.Handle("/api/secure-ping", auth.AuthMiddleware(securePingHandler))

	// Configure CORS
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000"}, // Your frontend URL
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
		AllowCredentials: true,
		Debug:            true, // Enable debug logging
	})

	// Wrap the main router with the CORS middleware
	handler := c.Handler(mux)

	port := ":8080"
	log.Printf("Backend server starting on port %s\n", port)

	// Use the CORS-wrapped handler
	server := &http.Server{
		Addr:         port,
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"ok": true})
}

func securePingHandlerFunc(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(auth.UserIDKey).(string)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Pong! You are authenticated.",
		"userID":  userID,
	})
}
