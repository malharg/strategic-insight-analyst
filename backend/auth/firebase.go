package auth

import (
	"context"
	"log"
	"os"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"google.golang.org/api/option"
)

var AuthClient *auth.Client

func InitFirebaseAuth() {
	var opt option.ClientOption

	// Render provides the content of a "Secret File" via an environment variable.
	// We check if that variable exists.
	firebaseCredentials := os.Getenv("FIREBASE_CREDENTIALS")

	if firebaseCredentials != "" {
		// If it exists (we're on Render), initialize from the JSON content.
		log.Println("Initializing Firebase Auth from FIREBASE_CREDENTIALS environment variable...")
		opt = option.WithCredentialsJSON([]byte(firebaseCredentials))
	} else {
		// Otherwise (we're running locally), initialize from the file.
		log.Println("Initializing Firebase Auth from serviceAccountKey.json file...")
		opt = option.WithCredentialsFile("serviceAccountKey.json")
	}
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		log.Fatalf("error initializing app: %v\n", err)
	}

	AuthClient, err = app.Auth(context.Background())
	if err != nil {
		log.Fatalf("error getting Auth client: %v\n", err)
	}

	log.Println("Firebase Auth client initialized successfully.")
}
