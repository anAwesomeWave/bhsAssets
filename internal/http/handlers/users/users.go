package users

import (
	"bhsAssets/internal/http/middleware/auth"
	"bhsAssets/internal/storage"
	"encoding/json"
	"log"
	"net/http"
)

func GetUserData(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.FromContext(r.Context())
	// maybe add info about assets
	if !ok {
		http.Error(w, "User not found in context", http.StatusUnauthorized)
		return
	}

	json.NewEncoder(w).Encode(user)
}
func UpdateBalanceInfo(strg storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		//user, ok := auth.FromContext(r.Context())
		// maybe add info about assets
		user_id, err := auth.IdFromContext(r.Context())
		if err != nil {
			http.Error(w, "unathorized request", http.StatusUnauthorized)
		}

		balance := struct {
			Balance float64
		}{0}
		if err := json.NewDecoder(r.Body).Decode(&balance); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}
		if err := strg.UpdateUserBalance(balance.Balance, user_id); err != nil {
			log.Printf("UpdateBalanceInfo: storage.UpdateUserBalance: %v", err)
			http.Error(w, "Internal Error", http.StatusInternalServerError)
		}

		log.Println(balance)
		json.NewEncoder(w).Encode(&balance)
	}
}
