package database

import (
	"finalproject/helpers"
	"finalproject/models"
	"testing"
)

func TestBootstrapUserUpdatesReconcilesPasswordRoleAndStatus(t *testing.T) {
	existing := models.User{
		Password:  helpers.HashPass("old-reviewer-password"),
		Role:      models.RoleUser,
		Status:    models.UserStatusBanned,
		BanReason: "assessment test",
	}

	updates := bootstrapUserUpdates(existing, "new-reviewer-password", models.RoleAdmin, true)

	password, ok := updates["password"].(string)
	if !ok || !helpers.ComparePass([]byte(password), []byte("new-reviewer-password")) {
		t.Fatal("expected the bootstrap password hash to be reconciled")
	}
	if updates["role"] != models.RoleAdmin {
		t.Fatalf("expected admin role update, got %#v", updates["role"])
	}
	if updates["status"] != models.UserStatusActive {
		t.Fatalf("expected active status update, got %#v", updates["status"])
	}
	if updates["banned_at"] != nil || updates["ban_reason"] != "" {
		t.Fatal("expected bootstrap reconciliation to clear the ban state")
	}
}

func TestBootstrapUserUpdatesLeavesMatchingActiveUserUnchanged(t *testing.T) {
	password := "matching-reviewer-password"
	existing := models.User{
		Password: helpers.HashPass(password),
		Role:     models.RoleUser,
		Status:   models.UserStatusActive,
	}

	updates := bootstrapUserUpdates(existing, password, models.RoleUser, false)
	if len(updates) != 0 {
		t.Fatalf("expected no updates for a matching active user, got %#v", updates)
	}
}
