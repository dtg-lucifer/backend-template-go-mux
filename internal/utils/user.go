package utils

import "github.com/your-username/go-mux-backend-template/internal/db/repository"

// SafeUser returns a sanitized user payload for API responses.
func SafeUser(u repository.User) map[string]any {
	return map[string]any{
		"id":         u.ID,
		"first_name": u.FirstName,
		"last_name":  u.LastName,
		"email":      u.Email,
		"verified":   u.Verified,
		"created_at": u.CreatedAt,
	}
}
