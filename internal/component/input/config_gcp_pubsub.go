package input

import (
	"cloud.google.com/go/pubsub"
)

// GCPPubSubConfig contains configuration values for the input type.
type GCPPubSubConfig struct {
	ProjectID              string `json:"project" yaml:"project"`
	SubscriptionID         string `json:"subscription" yaml:"subscription"`
	MaxOutstandingMessages int    `json:"max_outstanding_messages" yaml:"max_outstanding_messages"`
	MaxOutstandingBytes    int    `json:"max_outstanding_bytes" yaml:"max_outstanding_bytes"`
	Sync                   bool   `json:"sync" yaml:"sync"`
}

// NewGCPPubSubConfig creates a new Config with default values.
func NewGCPPubSubConfig() GCPPubSubConfig {
	return GCPPubSubConfig{
		ProjectID:              "",
		SubscriptionID:         "",
		MaxOutstandingMessages: pubsub.DefaultReceiveSettings.MaxOutstandingMessages,
		MaxOutstandingBytes:    pubsub.DefaultReceiveSettings.MaxOutstandingBytes,
		Sync:                   false,
	}
}
