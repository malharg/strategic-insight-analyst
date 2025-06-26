package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/malharg/strategic-insight-analyst/backend/ai"
	"github.com/malharg/strategic-insight-analyst/backend/auth"
	"github.com/malharg/strategic-insight-analyst/backend/database"
)

// Make sure the struct definition has the correct json tags.
type ChatRequest struct {
	DocumentID string `json:"documentId"`
	Query      string `json:"query"`
}

func ChatHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Get user ID from the authentication middleware.
	userID := r.Context().Value(auth.UserIDKey).(string)

	// 2. Decode the incoming JSON request from the frontend.
	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 3. Add a debug log to confirm we received the data correctly.
	log.Printf("DEBUG: Chat request received for docID: [%s], query: [%s]", req.DocumentID, req.Query)

	// 4. Check if the document ID is empty. If so, it's a client error.
	if req.DocumentID == "" {
		log.Println("ERROR: Received chat request with empty document ID.")
		http.Error(w, "Document ID is required.", http.StatusBadRequest)
		return
	}

	// 5. Generate the insight using our AI service.
	aiResponse, err := ai.GenerateInsight(r.Context(), req.DocumentID, req.Query)
	if err != nil {
		log.Printf("Error generating insight: %v", err)
		http.Error(w, "Failed to generate AI insight.", http.StatusInternalServerError)
		return
	}

	// 6. Respond to the frontend first. This makes the UI feel faster.
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"response": aiResponse})

	// 7. After responding, save the interaction to chat history in the background.
	// This is a "fire-and-forget" operation. If it fails, it doesn't break the user experience.
	go func() {
		tx, err := database.DB.Begin()
		if err != nil {
			log.Printf("Failed to begin transaction for chat history: %v", err)
			return
		}
		defer tx.Rollback() // Rollback on error

		// Save user message
		_, err = tx.Exec("INSERT INTO chat_history (id, document_id, user_id, message_type, message_content) VALUES (?, ?, ?, ?, ?)",
			uuid.New().String(), req.DocumentID, userID, "user", req.Query)
		if err != nil {
			log.Printf("Failed to save user message to chat history: %v", err)
			return
		}

		// Save AI response
		_, err = tx.Exec("INSERT INTO chat_history (id, document_id, user_id, message_type, message_content) VALUES (?, ?, ?, ?, ?)",
			uuid.New().String(), req.DocumentID, userID, "ai", aiResponse)
		if err != nil {
			log.Printf("Failed to save AI response to chat history: %v", err)
			return
		}

		if err := tx.Commit(); err != nil {
			log.Printf("Failed to commit chat history transaction: %v", err)
		} else {
			log.Printf("Successfully saved chat history for docID: %s", req.DocumentID)
		}
	}()
}
