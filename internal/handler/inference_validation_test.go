package handler

import (
	"strings"
	"testing"
)

func TestValidatePrompt_AllowsAtLimitUnicodeChars(t *testing.T) {
	prompt := strings.Repeat("🙂", 32000)
	if err := validatePrompt(prompt); err != nil {
		t.Fatalf("expected prompt at limit to be accepted, got error: %v", err)
	}
}

func TestValidatePrompt_RejectsOverLimitUnicodeChars(t *testing.T) {
	prompt := strings.Repeat("🙂", 32001)
	if err := validatePrompt(prompt); err == nil {
		t.Fatal("expected prompt over limit to be rejected")
	}
}
