package helpers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"finalproject/config"
)

var (
	ErrCaptchaNotConfigured = errors.New("captcha is not configured")
	ErrCaptchaRequired      = errors.New("captcha token is required")
	ErrCaptchaFailed        = errors.New("captcha verification failed")
	ErrCaptchaUnavailable   = errors.New("captcha verification service is unavailable")
)

type captchaVerifyRequest struct {
	Secret   string `json:"secret"`
	Response string `json:"response"`
}

type captchaVerifyResponse struct {
	Success bool     `json:"success"`
	Error   string   `json:"error,omitempty"`
	Message string   `json:"message,omitempty"`
	Errors  []string `json:"errors,omitempty"`
}

func VerifyCaptchaToken(cfg config.Config, token string) error {
	if !cfg.CapEnabled {
		return nil
	}

	token = strings.TrimSpace(token)
	if token == "" {
		return ErrCaptchaRequired
	}

	if strings.TrimSpace(cfg.CapBaseURL) == "" || strings.TrimSpace(cfg.CapSiteKey) == "" || strings.TrimSpace(cfg.CapSecretKey) == "" {
		return ErrCaptchaNotConfigured
	}

	endpoint := fmt.Sprintf("%s/%s/siteverify", strings.TrimRight(cfg.CapBaseURL, "/"), url.PathEscape(cfg.CapSiteKey))
	body, err := json.Marshal(captchaVerifyRequest{
		Secret:   cfg.CapSecretKey,
		Response: token,
	})
	if err != nil {
		return fmt.Errorf("%w: %v", ErrCaptchaFailed, err)
	}

	request, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("%w: %v", ErrCaptchaNotConfigured, err)
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")

	client := http.Client{Timeout: 5 * time.Second}
	response, err := client.Do(request)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrCaptchaUnavailable, err)
	}
	defer response.Body.Close()

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		detail := captchaErrorDetail(response)
		switch response.StatusCode {
		case http.StatusBadRequest:
			if detail == "" {
				return fmt.Errorf("%w: status %d", ErrCaptchaFailed, response.StatusCode)
			}
			return fmt.Errorf("%w: %s", ErrCaptchaFailed, detail)
		case http.StatusUnauthorized, http.StatusForbidden:
			if detail == "" {
				detail = fmt.Sprintf("status %d", response.StatusCode)
			}
			return fmt.Errorf(
				"%w: captcha verification was rejected (%s). check CAP_SITE_KEY, CAP_SECRET_KEY, CAP_BASE_URL, or WAF allowlist",
				ErrCaptchaNotConfigured,
				detail,
			)
		default:
			if detail == "" {
				return fmt.Errorf("%w: status %d", ErrCaptchaUnavailable, response.StatusCode)
			}
			return fmt.Errorf("%w: status %d: %s", ErrCaptchaUnavailable, response.StatusCode, detail)
		}
	}

	result := captchaVerifyResponse{}
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		return fmt.Errorf("%w: invalid response", ErrCaptchaFailed)
	}

	if result.Success {
		return nil
	}

	reason := captchaFailureReason(result)
	if reason == "" {
		return ErrCaptchaFailed
	}

	return fmt.Errorf("%w: %s", ErrCaptchaFailed, reason)
}

func captchaErrorDetail(response *http.Response) string {
	body, err := io.ReadAll(io.LimitReader(response.Body, 1024))
	if err != nil || len(body) == 0 {
		return ""
	}

	result := captchaVerifyResponse{}
	if err := json.Unmarshal(body, &result); err == nil {
		if reason := captchaFailureReason(result); reason != "" {
			return reason
		}
	}

	return compactCaptchaDetail(string(body))
}

func captchaFailureReason(result captchaVerifyResponse) string {
	reason := strings.TrimSpace(result.Message)
	if reason == "" {
		reason = strings.TrimSpace(result.Error)
	}
	if reason == "" && len(result.Errors) > 0 {
		reason = strings.Join(result.Errors, ", ")
	}
	return compactCaptchaDetail(reason)
}

func compactCaptchaDetail(value string) string {
	value = strings.Join(strings.Fields(value), " ")
	if len(value) > 180 {
		return value[:180] + "..."
	}
	return value
}
