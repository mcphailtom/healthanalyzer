package llm

import "context"

// Message represents a single message in a conversation with an LLM.
type Message struct {
	Role    string // "system", "user", "assistant"
	Content string
}

// CompletionClient generates text completions from an LLM.
type CompletionClient interface {
	Complete(ctx context.Context, messages []Message) (string, error)
}

// EmbeddingClient generates vector embeddings from text.
type EmbeddingClient interface {
	Embed(ctx context.Context, text string) ([]float32, error)
}
