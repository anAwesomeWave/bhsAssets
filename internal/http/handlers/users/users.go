package users

import (
	"bhsAssets/internal/http/middleware/auth"
	"bhsAssets/internal/http/middleware/common"
	"bhsAssets/internal/storage"
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"strconv"
)

func GetUserData(strg storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const fn = "internal.http.users.GetUserData"

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

		boughtAssets, err := strg.GetBoughtAssets(user.Id)
		if err != nil {
			log.Printf("%s: failed to get bought assets for user with id %d: %v\n", fn, user.Id, err)
			http.Error(w, "Failed to get bought assets ", http.StatusInternalServerError)
		}
		createdAssets, err := strg.GetAllAssetsFiltered(map[string]string{"creator_id": strconv.FormatInt(user.Id, 10)})
		if err != nil {
			log.Printf("%s: failed to get created assets for user with id %d: %v\n", fn, user.Id, err)
			http.Error(w, "Failed to get created assets ", http.StatusInternalServerError)
		}
		if isApi {
			json.NewEncoder(w).Encode(user)
		} else {
			tmpl, err := template.ParseFiles("./templates/common/base.html", "./templates/users/me.html")
			if err != nil {
				http.Error(w, "Error loading template", http.StatusInternalServerError)
				log.Println(err)
				return
			}
			if err := tmpl.Execute(w, map[string]interface{}{
				"user":          user,
				"isLogined":     user.Id > 0,
				"createdAssets": createdAssets,
				"boughtAssets":  boughtAssets,
			}); err != nil {
				http.Error(w, "Error templating", http.StatusInternalServerError)
				log.Println(err)
				return
			}

		}
	}
}

func UpdateBalanceInfo(strg storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const fn = "internal.http.auth.UpdateBalanceInfo"
		userId, err := auth.IdFromContext(r.Context())
		if err != nil {
			http.Error(w, "Unauthorized request", http.StatusUnauthorized)
		}
		isApi, ok := common.IsApiFromContext(r.Context())
		if !ok {
			http.Error(w, "Failed to get context", http.StatusInternalServerError)
		}
		balance := &struct {
			Balance float64
		}{0}
		if isApi {
			if err := json.NewDecoder(r.Body).Decode(balance); err != nil {
				http.Error(w, "Invalid request", http.StatusBadRequest)
				return
			}
		} else {
			if err := r.ParseForm(); err != nil {
				http.Error(w, "Unable to parse form", http.StatusInternalServerError)
				return
			}
			if _, ok := r.Form["balance"]; !ok {
				http.Error(w, "Bad form. Unable to find `balance` field", http.StatusBadRequest)
				return
			}
			balance.Balance, err = strconv.ParseFloat(r.Form["balance"][0], 64)
			if err != nil {
				log.Printf("%s: unable to convert balance to float %v: %v", fn, r.Form["balance"][0], err)
				http.Error(w, "Bad form. Unable to convert `balance` field", http.StatusBadRequest)
				return
			}
		}
		if err := strg.UpdateUserBalance(balance.Balance, userId); err != nil {
			log.Printf("UpdateBalanceInfo: storage.UpdateUserBalance: %v", err)
			http.Error(w, "Internal Error", http.StatusInternalServerError)
			return
		}
		if isApi {
			json.NewEncoder(w).Encode(&balance)
			return
		} else {
			http.Redirect(
				w,
				r,
				"/users/me",
				http.StatusSeeOther,
			)
		}
	}
}

func UpdateBalancePage(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.FromContext(r.Context())
	if !ok {
		http.Error(w, "Cannot get user from context", http.StatusInternalServerError)
	}
	tmpl, err := template.ParseFiles("./templates/common/base.html", "./templates/users/balance.html")
	if err != nil {
		http.Error(w, "Error loading template", http.StatusInternalServerError)
		log.Println(err)
		return
	}
	if err := tmpl.Execute(w, map[string]interface{}{
		"user":      user,
		"isLogined": user.Id > 0,
	}); err != nil {
		http.Error(w, "Error templating", http.StatusInternalServerError)
		log.Println(err)
		return
	}
}
