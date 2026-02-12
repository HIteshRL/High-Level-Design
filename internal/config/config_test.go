package config

import "testing"

func TestValidatePanicsInProductionWithDefaultJWTSecret(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("JWT_SECRET", "CHANGE_ME_generate_with_openssl_rand_hex_32")
	t.Setenv("LLM_API_KEY", "test-key")

	cfg := Load()

	didPanic := false
	func() {
		defer func() {
			if recover() != nil {
				didPanic = true
			}
		}()
		cfg.Validate()
	}()

	if !didPanic {
		t.Fatal("expected Validate to panic in production with default JWT secret")
	}
}

func TestValidatePanicsInProductionWithoutLLMAPIKey(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("JWT_SECRET", "this_is_a_secure_32_characters_min_secret")
	t.Setenv("LLM_API_KEY", "")

	cfg := Load()

	didPanic := false
	func() {
		defer func() {
			if recover() != nil {
				didPanic = true
			}
		}()
		cfg.Validate()
	}()

	if !didPanic {
		t.Fatal("expected Validate to panic in production when LLM_API_KEY is missing")
	}
}

func TestValidateDoesNotPanicInDevelopment(t *testing.T) {
	t.Setenv("APP_ENV", "development")
	t.Setenv("JWT_SECRET", "CHANGE_ME_generate_with_openssl_rand_hex_32")
	t.Setenv("LLM_API_KEY", "")

	cfg := Load()

	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("did not expect panic in development, got: %v", r)
			}
		}()
		cfg.Validate()
	}()
}
