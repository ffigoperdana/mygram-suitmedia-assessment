package services

import (
	"encoding/json"
	"finalproject/config"
	"finalproject/models"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	webpush "github.com/SherClockHolmes/webpush-go"
	"gorm.io/gorm"
)

type pushPayload struct {
	Title string `json:"title"`
	Body  string `json:"body"`
	URL   string `json:"url"`
	Tag   string `json:"tag"`
}

func NotifyNewPhoto(db *gorm.DB, photo models.Photo) {
	if db == nil || photo.ID == 0 {
		return
	}

	var userIDs []uint
	if err := db.Model(&models.PushSubscription{}).
		Distinct("user_id").
		Where("user_id <> ?", photo.UserID).
		Pluck("user_id", &userIDs).Error; err != nil {
		log.Printf("push notification user lookup failed: %v", err)
		return
	}

	if len(userIDs) == 0 {
		return
	}

	sendPushToUsers(db, userIDs, pushPayload{
		Title: "New post on MyGram",
		Body:  fmt.Sprintf("%s was posted.", fallbackPushTitle(photo.Title, "A new photo")),
		URL:   "/feed",
		Tag:   "mygram-new-posts",
	})
}

func NotifyNewComment(db *gorm.DB, photo models.Photo, comment models.Comment) {
	if db == nil || photo.ID == 0 || photo.UserID == 0 || photo.UserID == comment.UserID {
		return
	}

	sendPushToUsers(db, []uint{photo.UserID}, pushPayload{
		Title: "New comment on your post",
		Body:  truncatePushBody(comment.Message, 120),
		URL:   fmt.Sprintf("/photos/%d", photo.ID),
		Tag:   fmt.Sprintf("mygram-photo-%d-comments", photo.ID),
	})
}

func sendPushToUsers(db *gorm.DB, userIDs []uint, payload pushPayload) {
	cfg := config.Load()
	if !cfg.PushNotificationsConfigured() {
		return
	}

	var subscriptions []models.PushSubscription
	if err := db.Where("user_id IN ?", userIDs).Find(&subscriptions).Error; err != nil {
		log.Printf("push notification subscription lookup failed: %v", err)
		return
	}

	if len(subscriptions) == 0 {
		return
	}

	body, err := json.Marshal(payload)
	if err != nil {
		log.Printf("push notification payload encode failed: %v", err)
		return
	}

	for _, subscription := range subscriptions {
		sendPush(db, cfg, subscription, body)
	}
}

func sendPush(db *gorm.DB, cfg config.Config, subscription models.PushSubscription, body []byte) {
	response, err := webpush.SendNotification(body, &webpush.Subscription{
		Endpoint: subscription.Endpoint,
		Keys: webpush.Keys{
			P256dh: subscription.P256DH,
			Auth:   subscription.Auth,
		},
	}, &webpush.Options{
		Subscriber:      cfg.VAPIDSubject,
		VAPIDPublicKey:  cfg.VAPIDPublicKey,
		VAPIDPrivateKey: cfg.VAPIDPrivateKey,
		TTL:             86400,
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	})
	if err != nil {
		log.Printf("push notification send failed: %v", err)
		return
	}
	defer drainAndClose(response.Body)

	if response.StatusCode == http.StatusGone || response.StatusCode == http.StatusNotFound {
		if err := db.Delete(&models.PushSubscription{}, subscription.ID).Error; err != nil {
			log.Printf("push subscription cleanup failed: %v", err)
		}
		return
	}

	if response.StatusCode >= http.StatusBadRequest {
		log.Printf("push notification rejected status=%d endpoint=%s", response.StatusCode, subscription.Endpoint)
	}
}

func fallbackPushTitle(value string, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}

	return value
}

func truncatePushBody(value string, max int) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "Someone commented on your post."
	}

	if len(value) <= max {
		return value
	}

	return value[:max] + "..."
}

func drainAndClose(body io.ReadCloser) {
	if body == nil {
		return
	}

	_, _ = io.Copy(io.Discard, body)
	_ = body.Close()
}
