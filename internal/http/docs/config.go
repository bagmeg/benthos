package docs

import (
	"github.com/benthosdev/benthos/v4/internal/http/docs/auth"
	"github.com/benthosdev/benthos/v4/internal/metadata"
	"github.com/benthosdev/benthos/v4/internal/tls"
)

// Config is a configuration struct for an HTTP client.
type Config struct {
	URL             string                       `json:"url" yaml:"url"`
	Verb            string                       `json:"verb" yaml:"verb"`
	Headers         map[string]string            `json:"headers" yaml:"headers"`
	Metadata        metadata.IncludeFilterConfig `json:"metadata" yaml:"metadata"`
	ExtractMetadata metadata.IncludeFilterConfig `json:"extract_headers" yaml:"extract_headers"`
	RateLimit       string                       `json:"rate_limit" yaml:"rate_limit"`
	Timeout         string                       `json:"timeout" yaml:"timeout"`
	Retry           string                       `json:"retry_period" yaml:"retry_period"`
	MaxBackoff      string                       `json:"max_retry_backoff" yaml:"max_retry_backoff"`
	NumRetries      int                          `json:"retries" yaml:"retries"`
	BackoffOn       []int                        `json:"backoff_on" yaml:"backoff_on"`
	DropOn          []int                        `json:"drop_on" yaml:"drop_on"`
	SuccessfulOn    []int                        `json:"successful_on" yaml:"successful_on"`
	TLS             tls.Config                   `json:"tls" yaml:"tls"`
	ProxyURL        string                       `json:"proxy_url" yaml:"proxy_url"`
	auth.Config     `json:",inline" yaml:",inline"`
	OAuth2          auth.OAuth2Config `json:"oauth2" yaml:"oauth2"`
}

// NewConfig creates a new Config with default values.
func NewConfig() Config {
	return Config{
		URL:             "",
		Verb:            "POST",
		Headers:         map[string]string{},
		ExtractMetadata: metadata.NewIncludeFilterConfig(),
		RateLimit:       "",
		Timeout:         "5s",
		Retry:           "1s",
		MaxBackoff:      "300s",
		NumRetries:      3,
		BackoffOn:       []int{429},
		DropOn:          []int{},
		SuccessfulOn:    []int{},
		TLS:             tls.NewConfig(),
		Config:          auth.NewConfig(),
		OAuth2:          auth.NewOAuth2Config(),
	}
}
