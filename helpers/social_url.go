package helpers

import (
	"errors"
	"net/url"
	"regexp"
	"strings"
)

var (
	socialHandlePattern        = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,99}$`)
	youtubeChannelValuePattern = regexp.MustCompile(`^[A-Za-z0-9_-]{3,120}$`)
)

var blockedSocialPathSegments = map[string]struct{}{
	"about":       {},
	"discover":    {},
	"events":      {},
	"explore":     {},
	"groups":      {},
	"hashtag":     {},
	"hashtags":    {},
	"help":        {},
	"home":        {},
	"intent":      {},
	"live":        {},
	"login":       {},
	"marketplace": {},
	"p":           {},
	"pages":       {},
	"post":        {},
	"posts":       {},
	"privacy":     {},
	"reel":        {},
	"reels":       {},
	"search":      {},
	"settings":    {},
	"share":       {},
	"shorts":      {},
	"signup":      {},
	"stories":     {},
	"story":       {},
	"support":     {},
	"tag":         {},
	"tags":        {},
	"terms":       {},
	"videos":      {},
	"watch":       {},
}

var unsupportedSocialProfileURLError = errors.New("social media URL must be a supported direct profile or channel URL")

// ValidateSocialProfileURL only allows direct profile/channel URLs for supported social platforms.
// Search, post, tracking, and generic website URLs are rejected intentionally.
func ValidateSocialProfileURL(rawURL string) error {
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

	if parsed.Port() != "" {
		return errors.New("social media URL must not include a custom port")
	}

	if parsed.RawQuery != "" || parsed.Fragment != "" {
		return errors.New("social media URL must be a direct profile or channel link without search, sharing, or tracking parameters")
	}

	host := normalizeSocialHost(parsed.Hostname())
	segments, err := socialPathSegments(parsed.Path)
	if err != nil || len(segments) == 0 || containsBlockedSocialSegment(segments) {
		return unsupportedSocialProfileURLError
	}

	if isSupportedSocialProfile(host, segments) {
		return nil
	}

	return unsupportedSocialProfileURLError
}

func normalizeSocialHost(host string) string {
	host = strings.ToLower(strings.TrimSpace(host))
	host = strings.TrimPrefix(host, "www.")
	host = strings.TrimPrefix(host, "m.")
	return host
}

func socialPathSegments(path string) ([]string, error) {
	path = strings.Trim(path, "/")
	if path == "" {
		return nil, nil
	}

	segments := strings.Split(path, "/")
	for index, segment := range segments {
		decoded, err := url.PathUnescape(segment)
		if err != nil {
			return nil, err
		}

		decoded = strings.TrimSpace(decoded)
		if decoded == "" || decoded == "." || decoded == ".." {
			return nil, errors.New("invalid social media url path")
		}

		segments[index] = decoded
	}

	return segments, nil
}

func containsBlockedSocialSegment(segments []string) bool {
	for _, segment := range segments {
		normalized := strings.ToLower(strings.TrimPrefix(segment, "@"))
		if _, blocked := blockedSocialPathSegments[normalized]; blocked {
			return true
		}
	}

	return false
}

func isSupportedSocialProfile(host string, segments []string) bool {
	switch host {
	case "github.com":
		return isSingleHandle(segments)
	case "linkedin.com":
		return len(segments) == 2 &&
			(segments[0] == "in" || segments[0] == "company") &&
			isPlainHandle(segments[1])
	case "instagram.com":
		return isSingleHandle(segments)
	case "tiktok.com":
		return isSinglePrefixedHandle(segments, "@")
	case "youtube.com":
		return isYouTubeProfile(segments)
	case "x.com", "twitter.com":
		return isSingleHandle(segments)
	case "facebook.com", "fb.com":
		return isSingleHandle(segments)
	case "threads.net":
		return isSinglePrefixedHandle(segments, "@")
	case "twitch.tv":
		return isSingleHandle(segments)
	default:
		return false
	}
}

func isSingleHandle(segments []string) bool {
	return len(segments) == 1 && isPlainHandle(segments[0])
}

func isSinglePrefixedHandle(segments []string, prefix string) bool {
	return len(segments) == 1 &&
		strings.HasPrefix(segments[0], prefix) &&
		isPlainHandle(strings.TrimPrefix(segments[0], prefix))
}

func isPlainHandle(handle string) bool {
	return socialHandlePattern.MatchString(handle)
}

func isYouTubeProfile(segments []string) bool {
	if isSinglePrefixedHandle(segments, "@") {
		return true
	}

	if len(segments) != 2 {
		return false
	}

	switch strings.ToLower(segments[0]) {
	case "channel", "c", "user":
		return youtubeChannelValuePattern.MatchString(segments[1])
	default:
		return false
	}
}
