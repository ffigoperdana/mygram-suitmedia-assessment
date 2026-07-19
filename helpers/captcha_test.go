package helpers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"finalproject/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVerifyCaptchaTokenDisabled(t *testing.T) {
	err := VerifyCaptchaToken(config.Config{CapEnabled: false}, "")
	require.NoError(t, err)
}

func TestVerifyCaptchaTokenRequiresToken(t *testing.T) {
	err := VerifyCaptchaToken(captchaTestConfig("https://cap.example.test"), " ")
	require.ErrorIs(t, err, ErrCaptchaRequired)
}

func TestVerifyCaptchaTokenSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/site-key/siteverify", r.URL.Path)
		assert.Equal(t, http.MethodPost, r.Method)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true}`))
	}))
	defer server.Close()

	err := VerifyCaptchaToken(captchaTestConfig(server.URL), "captcha-token")
	require.NoError(t, err)
}

func TestVerifyCaptchaTokenClassifiesForbiddenAsConfigurationIssue(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"success":false,"error":"FORBIDDEN"}`))
	}))
	defer server.Close()

	err := VerifyCaptchaToken(captchaTestConfig(server.URL), "captcha-token")
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrCaptchaNotConfigured), err)
	assert.Contains(t, err.Error(), "CAP_SECRET_KEY")
}

func TestVerifyCaptchaTokenClassifiesServerFailureAsUnavailable(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte(`{"success":false,"error":"temporarily down"}`))
	}))
	defer server.Close()

	err := VerifyCaptchaToken(captchaTestConfig(server.URL), "captcha-token")
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrCaptchaUnavailable), err)
	assert.Contains(t, err.Error(), "temporarily down")
}

func captchaTestConfig(baseURL string) config.Config {
	return config.Config{
		CapEnabled:   true,
		CapBaseURL:   baseURL,
		CapSiteKey:   "site-key",
		CapSecretKey: "secret-key",
	}
}
