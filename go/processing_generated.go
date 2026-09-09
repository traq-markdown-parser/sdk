// Code generated from Rust preset exports. DO NOT EDIT.
package markdown

type ProcessorPreset string

const (
	ProcessorPresetTraQV1 ProcessorPreset = "traq.v1"
)

type ProcessorOptions struct {
	Origin string `json:"origin"`
}
type ProcessOutput struct {
	Attachments      []string      `json:"attachments"`
	Citations        []string      `json:"citations"`
	Embedding        EmbeddingPlan `json:"embedding"`
	NotificationText string        `json:"notificationText"`
	PlainText        string        `json:"plainText"`
	References       References    `json:"references"`
}
type EmbeddedInfo struct {
	ID   string `json:"id"`
	Raw  string `json:"raw"`
	Type string `json:"type"`
}
type EmbeddingCandidate struct {
	End   uint32 `json:"end"`
	Kind  string `json:"kind"`
	Name  string `json:"name"`
	Raw   string `json:"raw"`
	Start uint32 `json:"start"`
}
type EmbeddingPlan struct {
	Candidates     []EmbeddingCandidate `json:"candidates"`
	UnembeddedText string               `json:"unembeddedText"`
}
type LookupKind = string
type References struct {
	ChannelLinks  []string       `json:"channelLinks"`
	Embeddings    []EmbeddedInfo `json:"embeddings"`
	GroupMentions []string       `json:"groupMentions"`
	Mentions      []string       `json:"mentions"`
}
