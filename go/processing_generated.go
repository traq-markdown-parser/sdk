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
	NotificationText string     `json:"notificationText"`
	References       References `json:"references"`
}
type References struct {
	ChannelLinks  []string `json:"channelLinks"`
	GroupMentions []string `json:"groupMentions"`
	Mentions      []string `json:"mentions"`
}
