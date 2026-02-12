package service

import (
	"testing"

	"github.com/prakyathpnayak/roognis/internal/models"
)

func TestSemanticHashStableForSameInputs(t *testing.T) {
	k1 := SemanticHash("prompt", "gpt-4o-mini", "user-1", 0.7, 1024)
	k2 := SemanticHash("prompt", "gpt-4o-mini", "user-1", 0.7, 1024)

	if k1 != k2 {
		t.Fatalf("expected same hash for identical inputs, got %q vs %q", k1, k2)
	}
}

func TestSemanticHashChangesAcrossRelevantDimensions(t *testing.T) {
	base := SemanticHash("prompt", "gpt-4o-mini", "user-1", 0.7, 1024)

	cases := []string{
		SemanticHash("prompt-2", "gpt-4o-mini", "user-1", 0.7, 1024),
		SemanticHash("prompt", "gpt-4o", "user-1", 0.7, 1024),
		SemanticHash("prompt", "gpt-4o-mini", "user-2", 0.7, 1024),
		SemanticHash("prompt", "gpt-4o-mini", "user-1", 0.9, 1024),
		SemanticHash("prompt", "gpt-4o-mini", "user-1", 0.7, 2048),
	}

	for i, key := range cases {
		if key == base {
			t.Fatalf("case %d: expected distinct hash from base", i)
		}
	}
}

func TestSemanticContextHashStableForSameInputs(t *testing.T) {
	messages := []models.LLMMessage{
		{Role: "system", Content: "system prompt"},
		{Role: "user", Content: "hello"},
	}

	k1, err := SemanticContextHash(messages, "qwen2.5:0.5b", "user-1", 0.7, 1024)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	k2, err := SemanticContextHash(messages, "qwen2.5:0.5b", "user-1", 0.7, 1024)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if k1 != k2 {
		t.Fatalf("expected same hash for identical full context, got %q vs %q", k1, k2)
	}
}

func TestSemanticContextHashChangesWhenContextChanges(t *testing.T) {
	baseMessages := []models.LLMMessage{
		{Role: "system", Content: "system prompt"},
		{Role: "user", Content: "hello"},
	}

	base, err := SemanticContextHash(baseMessages, "qwen2.5:0.5b", "user-1", 0.7, 1024)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	variants := [][]models.LLMMessage{
		{
			{Role: "system", Content: "system prompt"},
			{Role: "assistant", Content: "hi there"},
			{Role: "user", Content: "hello"},
		},
		{
			{Role: "system", Content: "changed system"},
			{Role: "user", Content: "hello"},
		},
	}

	for i, msgs := range variants {
		key, hashErr := SemanticContextHash(msgs, "qwen2.5:0.5b", "user-1", 0.7, 1024)
		if hashErr != nil {
			t.Fatalf("case %d unexpected error: %v", i, hashErr)
		}
		if key == base {
			t.Fatalf("case %d expected key to differ from base context", i)
		}
	}
}
