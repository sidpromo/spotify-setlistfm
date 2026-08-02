package auth

import (
	"testing"
	"time"
)

func TestJWTService_GenerateAndValidate(t *testing.T) {
	svc := NewJWTService("test-secret-key-32chars!!", 15*time.Minute, 7*24*time.Hour)

	token, err := svc.GenerateAccessToken("user-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token == "" {
		t.Fatal("expected non-empty token")
	}

	claims, err := svc.ValidateToken(token)
	if err != nil {
		t.Fatalf("validate failed: %v", err)
	}
	if claims.UserID != "user-123" {
		t.Errorf("expected userID 'user-123', got %q", claims.UserID)
	}
}

func TestJWTService_GenerateRefreshToken(t *testing.T) {
	svc := NewJWTService("test-secret", 15*time.Minute, 7*24*time.Hour)

	token, err := svc.GenerateRefreshToken("user-456")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	claims, err := svc.ValidateToken(token)
	if err != nil {
		t.Fatalf("validate failed: %v", err)
	}
	if claims.UserID != "user-456" {
		t.Errorf("expected 'user-456', got %q", claims.UserID)
	}
}

func TestJWTService_ExpiredToken(t *testing.T) {
	svc := NewJWTService("test-secret", -1*time.Second, 7*24*time.Hour)

	token, _ := svc.GenerateAccessToken("user-789")
	_, err := svc.ValidateToken(token)
	if err != ErrInvalidToken {
		t.Fatalf("expected ErrInvalidToken for expired token, got %v", err)
	}
}

func TestJWTService_InvalidSignature(t *testing.T) {
	svc1 := NewJWTService("secret-1", 15*time.Minute, 7*24*time.Hour)
	svc2 := NewJWTService("secret-2", 15*time.Minute, 7*24*time.Hour)

	token, _ := svc1.GenerateAccessToken("user-123")
	_, err := svc2.ValidateToken(token)
	if err != ErrInvalidToken {
		t.Fatalf("expected ErrInvalidToken for wrong secret, got %v", err)
	}
}

func TestJWTService_GarbageToken(t *testing.T) {
	svc := NewJWTService("test-secret", 15*time.Minute, 7*24*time.Hour)

	_, err := svc.ValidateToken("not.a.valid.jwt")
	if err != ErrInvalidToken {
		t.Fatalf("expected ErrInvalidToken, got %v", err)
	}
}

func TestJWTService_EmptyToken(t *testing.T) {
	svc := NewJWTService("test-secret", 15*time.Minute, 7*24*time.Hour)

	_, err := svc.ValidateToken("")
	if err != ErrInvalidToken {
		t.Fatalf("expected ErrInvalidToken, got %v", err)
	}
}
