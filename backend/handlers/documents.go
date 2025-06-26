package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/malharg/strategic-insight-analyst/backend/auth"
	"github.com/malharg/strategic-insight-analyst/backend/config"
	"github.com/malharg/strategic-insight-analyst/backend/database"
	"github.com/malharg/strategic-insight-analyst/backend/processing"
	storage_go "github.com/supabase-community/storage-go"
)

const supabaseBucketName = "documents"

func UploadDocumentHandler(w http.ResponseWriter, r *http.Request) {
	// --- Step 1 & 2: Auth and File Parsing (No changes needed here) ---
	userID := r.Context().Value(auth.UserIDKey).(string)
	userRecord, err := auth.AuthClient.GetUser(r.Context(), userID)
	if err != nil {
		log.Printf("Failed to get user record from Firebase: %v", err)
		http.Error(w, "Could not verify user.", http.StatusInternalServerError)
		return
	}
	userSQL := "INSERT OR IGNORE INTO users (id, email) VALUES (?, ?)"
	if _, err = database.DB.Exec(userSQL, userID, userRecord.Email); err != nil {
		log.Printf("Failed to upsert user: %v", err)
		http.Error(w, "Failed to save user data.", http.StatusInternalServerError)
		return
	}

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "File is too large.", http.StatusBadRequest)
		return
	}
	file, header, err := r.FormFile("document")
	if err != nil {
		http.Error(w, "Invalid file key 'document'.", http.StatusBadRequest)
		return
	}
	defer file.Close()

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Could not read file.", http.StatusInternalServerError)
		return
	}

	// --- Step 3: Direct HTTP Upload to Supabase (Your working fix, no changes) ---
	docID := uuid.New().String()
	storagePath := fmt.Sprintf("%s/%s/%s", userID, docID, filepath.Base(header.Filename))
	contentType := http.DetectContentType(fileBytes)
	uploadURL := fmt.Sprintf("%s/storage/v1/object/%s/%s", config.AppConfig.SupabaseURL, supabaseBucketName, storagePath)

	req, err := http.NewRequest("POST", uploadURL, bytes.NewReader(fileBytes))
	if err != nil {
		http.Error(w, "Failed to create upload request.", http.StatusInternalServerError)
		return
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", config.AppConfig.SupabaseSvcKey))
	req.Header.Set("Content-Type", contentType)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		http.Error(w, "Failed to upload file to cloud storage.", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		log.Printf("Direct upload failed: Status=%s, Body=%s", resp.Status, string(respBody))
		http.Error(w, "Failed to upload file to cloud storage.", http.StatusInternalServerError)
		return
	}

	log.Printf("File %s uploaded successfully to path: %s", header.Filename, storagePath)

	// =========================================================================
	// START OF NEW PROCESSING & DATABASE LOGIC
	// =========================================================================

	// --- Step 4: Extract text content from the file ---
	textContent, err := processing.ExtractTextFromFile(fileBytes, header.Filename)
	if err != nil {
		log.Printf("ERROR during text extraction: %v", err)
		http.Error(w, "File uploaded, but failed to extract text content.", http.StatusInternalServerError)
		return
	}
	log.Printf("DEBUG: Extracted ~%d characters of text.", len(textContent))

	// --- Step 5: Chunk the extracted text ---
	textChunks := processing.ChunkText(textContent)
	log.Printf("DEBUG: Document split into %d chunks.", len(textChunks))

	if len(textChunks) == 0 {
		log.Println("WARN: No chunks were generated from the document. Nothing to save to chunks table.")
	}

	// --- Step 6: Save document metadata and all chunks in a single database transaction ---
	ctx := r.Context()
	tx, err := database.DB.BeginTx(ctx, nil)
	if err != nil {
		log.Printf("ERROR: Failed to begin database transaction: %v", err)
		http.Error(w, "Failed to start database transaction.", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback() // Ensures rollback on any error path

	// Insert the parent document record
	sqlDoc := "INSERT INTO documents (id, user_id, file_name, storage_path) VALUES (?, ?, ?, ?)"
	if _, err := tx.ExecContext(ctx, sqlDoc, docID, userID, header.Filename, storagePath); err != nil {
		log.Printf("ERROR: Failed to insert document metadata in transaction: %v", err)
		http.Error(w, "Failed to save document metadata.", http.StatusInternalServerError)
		return
	}
	log.Printf("DEBUG: Inserted document record with ID: %s", docID)

	// Prepare the statement for inserting chunks for efficiency
	stmt, err := tx.PrepareContext(ctx, "INSERT INTO document_chunks (id, document_id, chunk_index, content) VALUES (?, ?, ?, ?)")
	if err != nil {
		log.Printf("ERROR: Failed to prepare chunk insert statement: %v", err)
		http.Error(w, "Failed to prepare for saving document content.", http.StatusInternalServerError)
		return
	}
	defer stmt.Close()

	var chunksInserted int
	for i, chunk := range textChunks {
		chunkID := uuid.New().String()
		// Use the prepared statement to insert each chunk
		if _, err := stmt.ExecContext(ctx, chunkID, docID, i, chunk); err != nil {
			log.Printf("ERROR: Failed to insert chunk %d for docID %s: %v", i, docID, err)
			http.Error(w, "Failed to save document content chunks.", http.StatusInternalServerError)
			return // This will trigger the deferred tx.Rollback()
		}
		chunksInserted++
	}

	// If we reach here, all inserts were successful. Commit the transaction.
	if err := tx.Commit(); err != nil {
		log.Printf("ERROR: Failed to commit transaction: %v", err)
		http.Error(w, "Failed to finalize saving document.", http.StatusInternalServerError)
		return
	}

	log.Printf("SUCCESS: Committed transaction. Saved document %s and %d chunks to DB.", docID, chunksInserted)

	// =========================================================================
	// END OF NEW PROCESSING & DATABASE LOGIC
	// =========================================================================

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("File uploaded and processed successfully!"))
}

type DocumentInfo struct {
	ID         string    `json:"id"`
	FileName   string    `json:"fileName"`
	UploadedAt time.Time `json:"uploadedAt"`
}

func ListDocumentsHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(auth.UserIDKey).(string)
	rows, err := database.DB.Query("SELECT id, file_name, uploaded_at FROM documents WHERE user_id = ? ORDER BY uploaded_at DESC", userID)
	if err != nil {
		http.Error(w, "Failed to retrieve documents.", http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	documents := make([]DocumentInfo, 0)
	for rows.Next() {
		var doc DocumentInfo
		if err := rows.Scan(&doc.ID, &doc.FileName, &doc.UploadedAt); err != nil {
			http.Error(w, "Failed to process document list.", http.StatusInternalServerError)
			return
		}
		documents = append(documents, doc)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(documents)
}

// Add a test endpoint for Supabase connectivity
func TestSupabaseUploadHandler(w http.ResponseWriter, r *http.Request) {
	storageClient := storage_go.NewClient(config.AppConfig.SupabaseURL, config.AppConfig.SupabaseSvcKey, nil)
	testContent := []byte("test upload from TestSupabaseUploadHandler")
	testPath := "test/test-upload.txt"
	contentType := "text/plain"
	_, err := storageClient.UploadFile(supabaseBucketName, testPath, bytes.NewReader(testContent), storage_go.FileOptions{
		ContentType: &contentType,
	})
	if err != nil {
		log.Printf("[TestSupabaseUploadHandler] Failed: %#v", err)
		if se, ok := err.(*storage_go.StorageError); ok {
			log.Printf("[TestSupabaseUploadHandler] Supabase StorageError: Status=%d, Message=%s", se.Status, se.Message)
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Supabase StorageError: Status=%d, Message=%s", se.Status, se.Message)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Upload error: %v", err)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Test upload to Supabase succeeded."))
}

// Minimal upload to Supabase Storage using net/http (no SDK)
func MinimalSupabaseUploadHandler(w http.ResponseWriter, r *http.Request) {
	url := fmt.Sprintf("%s/storage/v1/object/%s/%s", config.AppConfig.SupabaseURL, supabaseBucketName, "test/minimal-upload.txt")
	content := []byte("minimal upload test")
	req, err := http.NewRequest("POST", url, bytes.NewReader(content))
	if err != nil {
		log.Printf("[MinimalSupabaseUploadHandler] Failed to create request: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Failed to create request: %v", err)
		return
	}
	// Set headers
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", config.AppConfig.SupabaseSvcKey))
	req.Header.Set("Content-Type", "text/plain")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("[MinimalSupabaseUploadHandler] Request error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Request error: %v", err)
		return
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	log.Printf("[MinimalSupabaseUploadHandler] Status: %s, Body: %s", resp.Status, string(respBody))
	w.WriteHeader(resp.StatusCode)
	fmt.Fprintf(w, "Status: %s\nBody: %s", resp.Status, string(respBody))
}

func DeleteDocumentHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Get user ID from context.
	userID := r.Context().Value(auth.UserIDKey).(string)

	// 2. Get the document ID from the request. We'll pass it as a query parameter.
	// e.g., /api/documents/delete?id=some-uuid
	docID := r.URL.Query().Get("id")
	if docID == "" {
		http.Error(w, "Document ID is required.", http.StatusBadRequest)
		return
	}

	log.Printf("Attempting to delete document ID: %s for user: %s", docID, userID)

	// 3. Find the document in the DB to get its storage_path and verify ownership.
	var storagePath string
	err := database.DB.QueryRow("SELECT storage_path FROM documents WHERE id = ? AND user_id = ?", docID, userID).Scan(&storagePath)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Document not found or you do not have permission to delete it.", http.StatusNotFound)
			return
		}
		log.Printf("Error finding document to delete: %v", err)
		http.Error(w, "Failed to find document.", http.StatusInternalServerError)
		return
	}

	// 4. Delete the file from Supabase storage (using your direct HTTP method)
	deleteURL := fmt.Sprintf("%s/storage/v1/object/%s/%s", config.AppConfig.SupabaseURL, supabaseBucketName, storagePath)
	req, err := http.NewRequest("DELETE", deleteURL, nil)
	if err != nil {
		http.Error(w, "Failed to create delete request.", http.StatusInternalServerError)
		return
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", config.AppConfig.SupabaseSvcKey))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		// Log the error but continue to try deleting from DB.
		log.Printf("Error deleting file from storage, but continuing to delete DB record. Error: %v", err)
	} else {
		defer resp.Body.Close()
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			respBody, _ := io.ReadAll(resp.Body)
			log.Printf("Storage delete failed with status %s: %s", resp.Status, string(respBody))
		} else {
			log.Printf("Successfully deleted file from storage: %s", storagePath)
		}
	}

	// 5. Delete the document record from our database.

	_, err = database.DB.Exec("DELETE FROM documents WHERE id = ?", docID)
	if err != nil {
		log.Printf("Failed to delete document from database: %v", err)
		http.Error(w, "Failed to delete document metadata.", http.StatusInternalServerError)
		return
	}

	log.Printf("Successfully deleted document record from DB: %s", docID)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Document deleted successfully."))
}
