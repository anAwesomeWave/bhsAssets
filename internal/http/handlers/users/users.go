package users

import (
	"bhsAssets/internal/http/middleware/auth"
	"encoding/json"
	"net/http"
)

func GetUserData(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.FromContext(r.Context())
	if !ok {
		http.Error(w, "User not found in context", http.StatusUnauthorized)
		return
	}

	json.NewEncoder(w).Encode(user)
}
