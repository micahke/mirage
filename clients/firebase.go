package clients

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"os"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/auth"
	"google.golang.org/api/option"
)

type FirebaseClient interface {
	CreateUser(ctx context.Context, email string, password string) (*auth.UserRecord, error)
	GetUser(ctx context.Context, uid string) (*auth.UserRecord, error)
	UpdateUser(ctx context.Context, uid string, user *auth.UserToUpdate) (*auth.UserRecord, error)
	DeleteUser(ctx context.Context, uid string) error
	GetUserByEmail(ctx context.Context, email string) (*auth.UserRecord, error)
	SetDisplayName(ctx context.Context, uid string, displayName string) error
	VerifyIdToken(ctx context.Context, idToken string) (*auth.Token, error)
}

type Client struct {
	app  *firebase.App
	auth *auth.Client
}

func NewFirebaseClientFromENV(ctx context.Context) *Client {
	app, err := firebase.NewApp(ctx, nil)
	if err != nil {
		log.Fatalf("error initializing app: %v", err)
	}

	auth, err := app.Auth(ctx)
	if err != nil {
		log.Fatalf("error getting auth client: %v", err)
	}

	return &Client{
		app:  app,
		auth: auth,
	}
}

func NewFirebaseClientFromServiceAccount(ctx context.Context, path string) *Client {
	opt := option.WithCredentialsFile(path)
	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		log.Fatalf("error initializing app: %v", err)
	}

	auth, err := app.Auth(ctx)
	if err != nil {
		log.Fatalf("error getting auth client: %v", err)
	}

	return &Client{
		app:  app,
		auth: auth,
	}
}

func NewFirebaseClientFromBase64String(ctx context.Context, base64Str string) *Client {
	decoded, err := base64.StdEncoding.DecodeString(base64Str)
	if err != nil {
		log.Fatalf("failed to decode base64 Firebase credentials: %v", err)
	}

	fmt.Println("Decoded Firebase credentials successfully")
	fmt.Println(string(decoded))

	// Create a temp file and write the decoded JSON to it
	tmpFile, err := os.CreateTemp("", "firebase-credentials-*.json")
	if err != nil {
		log.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write(decoded); err != nil {
		tmpFile.Close()
		log.Fatalf("failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	fmt.Println("Temp file created successfully:", tmpFile.Name())

	// Use the existing function with the temp file path
	return NewFirebaseClientFromServiceAccount(ctx, tmpFile.Name())
}

func (c *Client) CreateUser(ctx context.Context, email string, password string) (*auth.UserRecord, error) {
	user := &auth.UserToCreate{}
	user.Email(email).Password(password)
	return c.auth.CreateUser(ctx, user)
}

func (c *Client) GetUser(ctx context.Context, uid string) (*auth.UserRecord, error) {
	return c.auth.GetUser(ctx, uid)
}

func (c *Client) UpdateUser(ctx context.Context, uid string, user *auth.UserToUpdate) (*auth.UserRecord, error) {
	return c.auth.UpdateUser(ctx, uid, user)
}

func (c *Client) DeleteUser(ctx context.Context, uid string) error {
	return c.auth.DeleteUser(ctx, uid)
}

func (c *Client) GetUserByEmail(ctx context.Context, email string) (*auth.UserRecord, error) {
	return c.auth.GetUserByEmail(ctx, email)
}

func (c *Client) VerifyIdToken(ctx context.Context, idToken string) (*auth.Token, error) {
	user, err := c.auth.VerifyIDToken(ctx, idToken)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (c *Client) SetDisplayName(ctx context.Context, uid string, displayName string) error {
	userToUpdate := &auth.UserToUpdate{}
	userToUpdate.DisplayName(displayName)
	_, err := c.auth.UpdateUser(ctx, uid, userToUpdate)
	return err
}
