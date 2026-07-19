package helpers

import (
	"errors"
	"net/url"
	"strings"
)

func ValidateHTTPURL(rawURL string) error {
	parsed, err := url.ParseRequestURI(strings.TrimSpace(rawURL))
	if err != nil {
		return errors.New("invalid url format")
	}

	scheme := strings.ToLower(parsed.Scheme)
	if scheme != "http" && scheme != "https" {
		return errors.New("url must start with http:// or https://")
	}

	if parsed.Host == "" {
		return errors.New("url must include a host")
	}

	return nil
}
