package ai

// Message represents one turn in a chat conversation.
// Valid Role values are "system", "user", or "assistant".
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}
