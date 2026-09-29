package auth

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestRegisterAndLogin(t *testing.T) {
	service := NewService(filepath.Join(t.TempDir(), "users.json"))

	password := "strong-pass-123"
	user, err := service.Register("  Student_1 ", password)
	if err != nil {
		t.Fatalf("Register returned error: %v", err)
	}
	if user.Username != "student_1" {
		t.Fatalf("unexpected username: %q", user.Username)
	}

	data, err := os.ReadFile(filepath.Join(filepath.Dir(service.path), "users.json"))
	if err != nil {
		t.Fatalf("failed to read stored accounts: %v", err)
	}
	if bytes.Contains(data, []byte(password)) {
		t.Fatal("password must not be stored in plain text")
	}

	user, err = service.Login("STUDENT_1", password)
	if err != nil {
		t.Fatalf("Login returned error: %v", err)
	}
	if user.Username != "student_1" {
		t.Fatalf("unexpected logged-in user: %q", user.Username)
	}

	if _, err := service.Login("student_1", "wrong-password"); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected invalid credentials, got %v", err)
	}
	if _, err := service.Register("student_1", "another-password"); !errors.Is(err, ErrUsernameTaken) {
		t.Fatalf("expected duplicate username error, got %v", err)
	}
}

func TestRegisterRejectsInvalidInput(t *testing.T) {
	service := NewService(filepath.Join(t.TempDir(), "users.json"))

	if _, err := service.Register("ab", "strong-pass-123"); !errors.Is(err, ErrInvalidUsername) {
		t.Fatalf("expected invalid username error, got %v", err)
	}
	if _, err := service.Register("student", "short"); !errors.Is(err, ErrInvalidPassword) {
		t.Fatalf("expected invalid password error, got %v", err)
	}
}

func TestRegisterRejectsCorruptedAccountDatabase(t *testing.T) {
	for _, invalidData := range []string{"null", ""} {
		t.Run(invalidData, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "users.json")
			if err := os.WriteFile(path, []byte(invalidData), 0o600); err != nil {
				t.Fatalf("failed to write invalid account database: %v", err)
			}

			service := NewService(path)
			if _, err := service.Register("student", "strong-pass-123"); err == nil {
				t.Fatal("expected invalid account database error")
			}
		})
	}
}
