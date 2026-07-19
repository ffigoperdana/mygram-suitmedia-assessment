package helpers

import "testing"

func TestValidateSocialProfileURL(t *testing.T) {
	t.Parallel()

	validURLs := []string{
		"https://github.com/example",
		"https://linkedin.com/in/example",
		"https://linkedin.com/company/example-inc",
		"https://instagram.com/example.user",
		"https://www.tiktok.com/@example_user",
		"https://youtube.com/@example",
		"https://youtube.com/channel/UC1234567890abcdef",
		"https://x.com/example",
		"https://twitter.com/example",
		"https://facebook.com/example.page",
		"https://threads.net/@example",
		"https://twitch.tv/example_channel",
	}

	for _, rawURL := range validURLs {
		if err := ValidateSocialProfileURL(rawURL); err != nil {
			t.Fatalf("expected %q to be valid: %v", rawURL, err)
		}
	}

	invalidURLs := []string{
		"not-a-url",
		"javascript:alert(1)",
		"https://example.com",
		"https://www.tiktok.com/search?q=robby%20pantjoro",
		"https://instagram.com/p/example",
		"https://x.com/example/status/123",
		"https://youtube.com/watch",
		"https://github.com/example?tab=repositories",
		"https://github.com:8443/example",
	}

	for _, rawURL := range invalidURLs {
		if err := ValidateSocialProfileURL(rawURL); err == nil {
			t.Fatalf("expected %q to be invalid", rawURL)
		}
	}
}
