package ocm

// Strategy defines an interface for context management strategies.
// Strategies allow customization of how contexts are processed, for example
// by limiting the number of messages (sliding window) or by limiting the
// total token count.
type Strategy interface {
	// Apply processes the context and returns a (potentially modified) copy.
	// The original context is not mutated.
	Apply(ctx *Context) *Context
}

// SlidingWindowStrategy keeps only the most recent N messages in the context.
// If a system message exists at the beginning, it is always preserved.
type SlidingWindowStrategy struct {
	MaxMessages int
}

// NewSlidingWindowStrategy creates a new SlidingWindowStrategy with the given
// maximum number of messages.
func NewSlidingWindowStrategy(maxMessages int) *SlidingWindowStrategy {
	return &SlidingWindowStrategy{MaxMessages: maxMessages}
}

// Apply keeps only the last MaxMessages messages, always preserving a leading
// system message if present.
func (s *SlidingWindowStrategy) Apply(ctx *Context) *Context {
	if ctx == nil || len(ctx.Messages) <= s.MaxMessages {
		return ctx
	}

	result := &Context{
		ID:        ctx.ID,
		Metadata:  ctx.Metadata,
		CreatedAt: ctx.CreatedAt,
		UpdatedAt: ctx.UpdatedAt,
	}

	msgs := ctx.Messages
	hasSystemPrefix := len(msgs) > 0 && msgs[0].Role == RoleSystem

	if hasSystemPrefix {
		// Keep system message + last (MaxMessages-1) messages
		keep := s.MaxMessages - 1
		if keep < 0 {
			keep = 0
		}
		start := len(msgs) - keep
		if start < 1 {
			start = 1
		}
		result.Messages = make([]Message, 0, 1+len(msgs[start:]))
		result.Messages = append(result.Messages, msgs[0])
		result.Messages = append(result.Messages, msgs[start:]...)
	} else {
		start := len(msgs) - s.MaxMessages
		result.Messages = make([]Message, len(msgs[start:]))
		copy(result.Messages, msgs[start:])
	}

	return result
}

// TokenCountFunc is a function that counts the number of tokens in a string.
type TokenCountFunc func(s string) int

// TokenLimitStrategy limits the context based on a maximum token count.
// It uses a configurable token counting function to estimate token usage.
type TokenLimitStrategy struct {
	MaxTokens      int
	TokenCountFunc TokenCountFunc
}

// NewTokenLimitStrategy creates a new TokenLimitStrategy.
// If tokenCountFunc is nil, a simple word-based estimation is used.
func NewTokenLimitStrategy(maxTokens int, tokenCountFunc TokenCountFunc) *TokenLimitStrategy {
	if tokenCountFunc == nil {
		tokenCountFunc = defaultTokenCount
	}
	return &TokenLimitStrategy{
		MaxTokens:      maxTokens,
		TokenCountFunc: tokenCountFunc,
	}
}

// defaultTokenCount provides a simple token count estimation based on
// character length divided by 4, which is a rough approximation.
func defaultTokenCount(s string) int {
	if len(s) == 0 {
		return 0
	}
	return (len(s) + 3) / 4
}

// Apply keeps messages from the end of the context until the token limit
// is reached. A leading system message is always preserved.
func (s *TokenLimitStrategy) Apply(ctx *Context) *Context {
	if ctx == nil || len(ctx.Messages) == 0 {
		return ctx
	}

	result := &Context{
		ID:        ctx.ID,
		Metadata:  ctx.Metadata,
		CreatedAt: ctx.CreatedAt,
		UpdatedAt: ctx.UpdatedAt,
	}

	msgs := ctx.Messages
	hasSystemPrefix := msgs[0].Role == RoleSystem

	totalTokens := 0
	var kept []Message

	// If there's a system message, always include it and count its tokens
	startIdx := 0
	if hasSystemPrefix {
		totalTokens += s.TokenCountFunc(msgs[0].Content)
		startIdx = 1
	}

	// Iterate from the end, adding messages until we hit the token limit
	for i := len(msgs) - 1; i >= startIdx; i-- {
		msgTokens := s.TokenCountFunc(msgs[i].Content)
		if totalTokens+msgTokens > s.MaxTokens {
			break
		}
		totalTokens += msgTokens
		kept = append([]Message{msgs[i]}, kept...)
	}

	if hasSystemPrefix {
		result.Messages = make([]Message, 0, 1+len(kept))
		result.Messages = append(result.Messages, msgs[0])
		result.Messages = append(result.Messages, kept...)
	} else {
		result.Messages = kept
	}

	return result
}
