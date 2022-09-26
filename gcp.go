package main

import (
	"context"
	"encoding/json"
	"fmt"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	secretmanagerpb "google.golang.org/genproto/googleapis/cloud/secretmanager/v1"
)

type Secret struct {
	Token          string `json:"token"`
	TokenSecret    string `json:"token_secret"`
	ConsumerKey    string `json:"consumer_key"`
	ConsumerSecret string `json:"consumer_secret"`
}

// Load keys/secrets for the twitter API from Google Cloud's Secret Manager
func readSecret(ctx context.Context) (*Secret, error) {
	// Create a client to read Google Cloud secrets
	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	// Build the secret access request
	accessRequest := &secretmanagerpb.AccessSecretVersionRequest{
		Name: fmt.Sprintf(
			"projects/%s/secrets/%s/versions/3",
			"1086379384073",
			"twitter-api",
		),
	}

	// Retrieve the secret
	result, err := client.AccessSecretVersion(ctx, accessRequest)
	if err != nil {
		return nil, err
	}

	// Decode the secret
	var secret *Secret
	if err := json.Unmarshal(result.Payload.Data, &secret); err != nil {
		return nil, err
	}
	return secret, nil
}
