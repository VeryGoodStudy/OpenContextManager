package ocm

import (
	"testing"
)

func TestSlidingWindowStrategy(t *testing.T) {
	ctx := NewContext("sw-test")
	ctx.AddMessage(NewMessage(RoleSystem, "You are a helper."))
	ctx.AddMessage(NewMessage(RoleUser, "msg1"))
	ctx.AddMessage(NewMessage(RoleAssistant, "resp1"))
	ctx.AddMessage(NewMessage(RoleUser, "msg2"))
	ctx.AddMessage(NewMessage(RoleAssistant, "resp2"))
	ctx.AddMessage(NewMessage(RoleUser, "msg3"))

	s := NewSlidingWindowStrategy(3)
	result := s.Apply(ctx)

	// Should keep system message + last 2 messages
	if len(result.Messages) != 3 {
		t.Fatalf("Messages len = %d, want 3", len(result.Messages))
	}
	if result.Messages[0].Role != RoleSystem {
		t.Errorf("Messages[0].Role = %q, want system", result.Messages[0].Role)
	}
	if result.Messages[1].Content != "resp2" {
		t.Errorf("Messages[1].Content = %q, want resp2", result.Messages[1].Content)
	}
	if result.Messages[2].Content != "msg3" {
		t.Errorf("Messages[2].Content = %q, want msg3", result.Messages[2].Content)
	}
}

func TestSlidingWindowStrategy_NoSystem(t *testing.T) {
	ctx := NewContext("sw-nosys")
	ctx.AddMessage(NewMessage(RoleUser, "a"))
	ctx.AddMessage(NewMessage(RoleAssistant, "b"))
	ctx.AddMessage(NewMessage(RoleUser, "c"))
	ctx.AddMessage(NewMessage(RoleAssistant, "d"))

	s := NewSlidingWindowStrategy(2)
	result := s.Apply(ctx)

	if len(result.Messages) != 2 {
		t.Fatalf("Messages len = %d, want 2", len(result.Messages))
	}
	if result.Messages[0].Content != "c" {
		t.Errorf("Messages[0].Content = %q, want c", result.Messages[0].Content)
	}
	if result.Messages[1].Content != "d" {
		t.Errorf("Messages[1].Content = %q, want d", result.Messages[1].Content)
	}
}

func TestSlidingWindowStrategy_UnderLimit(t *testing.T) {
	ctx := NewContext("sw-under")
	ctx.AddMessage(NewMessage(RoleUser, "hello"))

	s := NewSlidingWindowStrategy(10)
	result := s.Apply(ctx)

	if len(result.Messages) != 1 {
		t.Fatalf("Messages len = %d, want 1", len(result.Messages))
	}
}

func TestSlidingWindowStrategy_Nil(t *testing.T) {
	s := NewSlidingWindowStrategy(5)
	result := s.Apply(nil)
	if result != nil {
		t.Error("Apply(nil) should return nil")
	}
}

func TestTokenLimitStrategy(t *testing.T) {
	ctx := NewContext("tl-test")
	ctx.AddMessage(NewMessage(RoleSystem, "sys"))                // ~1 token
	ctx.AddMessage(NewMessage(RoleUser, "short"))                // ~2 tokens
	ctx.AddMessage(NewMessage(RoleAssistant, "also short"))      // ~3 tokens
	ctx.AddMessage(NewMessage(RoleUser, "a longer message here")) // ~6 tokens

	// Custom token counter: 1 token per word
	wordCount := func(s string) int {
		if s == "" {
			return 0
		}
		count := 1
		for _, c := range s {
			if c == ' ' {
				count++
			}
		}
		return count
	}

	s := NewTokenLimitStrategy(5, wordCount)
	result := s.Apply(ctx)

	// System message (1 token) + last message (4 tokens) = 5 tokens
	if len(result.Messages) != 2 {
		t.Fatalf("Messages len = %d, want 2", len(result.Messages))
	}
	if result.Messages[0].Role != RoleSystem {
		t.Errorf("Messages[0].Role = %q, want system", result.Messages[0].Role)
	}
	if result.Messages[1].Content != "a longer message here" {
		t.Errorf("Messages[1].Content = %q, want 'a longer message here'", result.Messages[1].Content)
	}
}

func TestTokenLimitStrategy_DefaultCounter(t *testing.T) {
	s := NewTokenLimitStrategy(100, nil)
	ctx := NewContext("default-counter")
	ctx.AddMessage(NewMessage(RoleUser, "hello world"))

	result := s.Apply(ctx)
	if len(result.Messages) != 1 {
		t.Fatalf("Messages len = %d, want 1", len(result.Messages))
	}
}

func TestTokenLimitStrategy_Nil(t *testing.T) {
	s := NewTokenLimitStrategy(100, nil)
	result := s.Apply(nil)
	if result != nil {
		t.Error("Apply(nil) should return nil")
	}
}

func TestDefaultTokenCount(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{"", 0},
		{"a", 1},
		{"abcd", 1},
		{"abcde", 2},
		{"abcdefgh", 2},
	}
	for _, tt := range tests {
		got := defaultTokenCount(tt.input)
		if got != tt.want {
			t.Errorf("defaultTokenCount(%q) = %d, want %d", tt.input, got, tt.want)
		}
	}
}
