package clients

import (
	"context"
	"log"

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
