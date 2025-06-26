package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/malharg/strategic-insight-analyst/backend/auth"
	"github.com/malharg/strategic-insight-analyst/backend/config"
	"github.com/malharg/strategic-insight-analyst/backend/database"
	"github.com/malharg/strategic-insight-analyst/backend/handlers"
	"github.com/rs/cors"
	"github.com/unidoc/unipdf/v3/common/license"
)

func main() {
	config.LoadConfig()
	err := license.SetMeteredKey(config.AppConfig.UnidocLicenseKey)
	if err != nil {
		log.Fatalf("FATAL: Failed to set UniDoc license key: %v", err)
	}
	log.Println("UniDoc license key set successfully.")
	database.InitDB("./sia.db")
	auth.InitFirebaseAuth()

	// Main router
	mux := http.NewServeMux()

	// --- Public Routes ---
	mux.HandleFunc("/api/health", healthCheckHandler)

	//  public route for testing Supabase upload
	mux.HandleFunc("/api/test-supabase-upload", handlers.TestSupabaseUploadHandler)

	// public route for minimal direct Supabase upload (no SDK)
	mux.HandleFunc("/api/minimal-supabase-upload", handlers.MinimalSupabaseUploadHandler)

	// --- Protected Routes ---
	// Handler for the secure ping test
	securePingHandler := http.HandlerFunc(securePingHandlerFunc)
	mux.Handle("/api/secure-ping", auth.AuthMiddleware(securePingHandler))

	// ADDED: Handler for document uploads. It's also protected by the auth middleware.
	uploadHandler := http.HandlerFunc(handlers.UploadDocumentHandler)
	mux.Handle("/api/documents/upload", auth.AuthMiddleware(uploadHandler))

	// chat handler for handling user chats
	chatHandler := http.HandlerFunc(handlers.ChatHandler)
	mux.Handle("/api/chat", auth.AuthMiddleware(chatHandler))

	//doc handler route
	listDocsHandler := http.HandlerFunc(handlers.ListDocumentsHandler)
	mux.Handle("/api/documents", auth.AuthMiddleware(listDocsHandler))

	//  new route for deleting documents
	deleteDocHandler := http.HandlerFunc(handlers.DeleteDocumentHandler)
	mux.Handle("/api/documents/delete", auth.AuthMiddleware(deleteDocHandler))

	// Configure CORS
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000"}, // Your frontend URL
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
		AllowCredentials: true,
		Debug:            true, // during development
	})

	// Wrap the main router with the CORS middleware
	handler := c.Handler(mux)

	port := ":8080"
	log.Printf("Backend server starting on port %s\n", port)

	// Use the CORS-wrapped handler
	server := &http.Server{
		Addr:         port,
		Handler:      handler,
		ReadTimeout:  15 * time.Second, // Increased timeout slightly for uploads
		WriteTimeout: 15 * time.Second,
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
	// Retrieve the user ID from the context (set by the middleware)
	userID := r.Context().Value(auth.UserIDKey).(string)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Pong! You are authenticated.",
		"userID":  userID,
	})
}
