package users

import (
	"bhsAssets/internal/http/middleware/auth"
	"bhsAssets/internal/http/middleware/common"
	"bhsAssets/internal/storage"
	"encoding/json"
	"html/template"
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
	isApi, ok := common.IsApiFromContext(r.Context())
	if !ok {
		http.Error(w, "Failed to get context", http.StatusInternalServerError)
	}
	if isApi {
		json.NewEncoder(w).Encode(user)
	} else {
		tmpl, err := template.ParseFiles("./templates/users/me.html")
		if err != nil {
			http.Error(w, "Error loading template", http.StatusInternalServerError)
			log.Println(err)
			return
		}
		if err := tmpl.Execute(w, user); err != nil {
			http.Error(w, "Error templating", http.StatusInternalServerError)
			log.Println(err)
			return
		}

	}
}
func UpdateBalanceInfo(strg storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userId, err := auth.IdFromContext(r.Context())
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
		if err := strg.UpdateUserBalance(balance.Balance, userId); err != nil {
			log.Printf("UpdateBalanceInfo: storage.UpdateUserBalance: %v", err)
			http.Error(w, "Internal Error", http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(&balance)
	}
}
