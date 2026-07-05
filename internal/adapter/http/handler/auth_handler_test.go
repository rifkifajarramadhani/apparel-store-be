package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	dto "github.com/rifkifajarramadhani/golang-clean-architecture/internal/adapter/http/dto/user"
	appauth "github.com/rifkifajarramadhani/golang-clean-architecture/internal/auth"
	"github.com/rifkifajarramadhani/golang-clean-architecture/internal/user"
)

type authServiceFake struct{}

func (authServiceFake) Register(_ context.Context, account *user.User) error {
	account.ID = 42
	return nil
}

func (authServiceFake) Login(context.Context, string, string) (*appauth.Tokens, error) {
	return nil, nil
}

func (authServiceFake) Refresh(context.Context, string) (*appauth.Tokens, error) {
	return nil, nil
}

func (authServiceFake) Logout(context.Context, string) error               { return nil }
func (authServiceFake) VerifyEmail(context.Context, string) error          { return nil }
func (authServiceFake) ResendVerification(context.Context, string) error   { return nil }
func (authServiceFake) SendVerificationForUser(context.Context, int) error { return nil }

func (authServiceFake) Me(context.Context, int) (*user.User, error) {
	return nil, nil
}

type verificationAuthFake struct {
	authServiceFake
	token string
	err   error
}

func (f *verificationAuthFake) VerifyEmail(_ context.Context, token string) error {
	f.token = token
	return f.err
}

func TestVerifyEmailLinkRedirectsToStorefront(t *testing.T) {
	tests := []struct {
		name       string
		requestURL string
		verifyErr  error
		location   string
		token      string
	}{
		{name: "success", requestURL: "/api/auth/verify-email?token=one-time-token", location: "https://shop.example.com/?verification=success", token: "one-time-token"},
		{name: "invalid token", requestURL: "/api/auth/verify-email?token=bad", verifyErr: errors.New("invalid token"), location: "https://shop.example.com/login?verification=invalid", token: "bad"},
		{name: "missing token", requestURL: "/api/auth/verify-email", location: "https://shop.example.com/login?verification=invalid"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			service := &verificationAuthFake{err: test.verifyErr}
			app := fiber.New()
			app.Get("/api/auth/verify-email", NewAuthHandler(service, slog.New(slog.NewTextHandler(io.Discard, nil)), "https://shop.example.com").VerifyEmailLink)
			response, err := app.Test(httptest.NewRequest("GET", test.requestURL, nil))
			if err != nil {
				t.Fatal(err)
			}

			defer func() { _ = response.Body.Close() }()
			if response.StatusCode != fiber.StatusSeeOther || response.Header.Get("Location") != test.location {
				t.Fatalf("status/location = %d %q", response.StatusCode, response.Header.Get("Location"))
			}

			if response.Header.Get("Cache-Control") != "no-store" || response.Header.Get("Referrer-Policy") != "no-referrer" {
				t.Fatalf("security headers = %+v", response.Header)
			}

			if service.token != test.token {
				t.Fatalf("token = %q, want %q", service.token, test.token)
			}
		})
	}
}

func TestRegisterResponseCompatibility(t *testing.T) {
	app := fiber.New()
	app.Post("/api/auth/register", NewAuthHandler(
		authServiceFake{},
		slog.New(slog.NewTextHandler(io.Discard, nil)),
	).Register)
	request := httptest.NewRequest("POST", "/api/auth/register", bytes.NewBufferString(
		`{"username":"rifki","email":"rifki@example.com","password":"long-password"}`,
	))
	request.Header.Set("Content-Type", "application/json")
	response, err := app.Test(request)
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = response.Body.Close() }()
	if response.StatusCode != fiber.StatusCreated {
		t.Fatalf("status = %d", response.StatusCode)
	}

	var body map[string]any
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}

	if body["id"] != float64(42) || body["username"] != "rifki" || body["email"] != "rifki@example.com" {
		t.Fatalf("response = %+v", body)
	}
}

func TestRegisterRejectsUnknownJSONFields(t *testing.T) {
	app := fiber.New()
	app.Post("/api/auth/register", NewAuthHandler(
		authServiceFake{},
		slog.New(slog.NewTextHandler(io.Discard, nil)),
	).Register)
	request := httptest.NewRequest("POST", "/api/auth/register", bytes.NewBufferString(
		`{"username":"rifki","email":"rifki@example.com","password":"long-password","role":"admin"}`,
	))
	request.Header.Set("Content-Type", "application/json")
	response, err := app.Test(request)
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = response.Body.Close() }()
	if response.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("status = %d", response.StatusCode)
	}
}

func TestProfileValuesSupportsPartialUpdates(t *testing.T) {
	username := "new-name"
	gotUsername, gotEmail := profileValues(
		dto.UpdateUserRequest{Username: &username},
		&user.User{Username: "old-name", Email: "user@example.com"},
	)
	if gotUsername != username || gotEmail != "user@example.com" {
		t.Fatalf("profile values = %q, %q", gotUsername, gotEmail)
	}
}
