package service

import "testing"

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
