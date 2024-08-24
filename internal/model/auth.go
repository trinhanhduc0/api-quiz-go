package models

import (
	"context"
	"log"
	"sync"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"google.golang.org/api/option"
)

var (
	firebaseConfigFile = "firebase-config.json"
	authClient         *auth.Client
	once               sync.Once
)

// GetAuth returns a singleton Firebase Auth client
func GetAuth() (*auth.Client, error) {
	var err error
	once.Do(func() {
		// Initialize Firebase
		ctx := context.Background()
		opt := option.WithCredentialsFile(firebaseConfigFile)

		app, initErr := firebase.NewApp(ctx, nil, opt)
		if initErr != nil {
			log.Fatalf("Error initializing Firebase app: %v", initErr)
			err = initErr
			return
		}

		authClient, err = app.Auth(ctx)
		if err != nil {
			log.Fatalf("Error initializing Firebase Auth client: %v", err)
		}
	})

	return authClient, err
}
