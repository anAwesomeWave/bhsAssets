package users

import (
	"bhsAssets/internal/http/handlers/site"
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

		isApi, ok := common.IsApiFromContext(r.Context())
		if !ok {
			http.Error(w, "Failed to get context", http.StatusInternalServerError)
			return
		}
		user, ok := auth.FromContext(r.Context())
		if !ok {
			site.ServeError(
				w,
				isApi,
				http.StatusUnauthorized,
				"user not found",
				false,
			)
			return
		}

		boughtAssets, err := strg.GetBoughtAssets(user.Id)
		if err != nil {
			log.Printf("%s: failed to get bought assets for user with id %d: %v\n", fn, user.Id, err)
			site.ServeError(
				w,
				isApi,
				http.StatusInternalServerError,
				"Failed to get info about bought assets.",
				user.Id > 0,
			)
			return
		}
		createdAssets, err := strg.GetAllAssetsFiltered(map[string]string{"creator_id": strconv.FormatInt(user.Id, 10)})
		if err != nil {
			log.Printf("%s: failed to get created assets for user with id %d: %v\n", fn, user.Id, err)
			site.ServeError(
				w,
				isApi,
				http.StatusInternalServerError,
				"Failed to get info about created assets.",
				user.Id > 0,
			)
			return
		}
		if isApi {
			json.NewEncoder(w).Encode(map[string]any{"user": user, "bought": boughtAssets, "created": createdAssets})
		} else {
			tmpl, err := template.ParseFiles("./templates/common/base.html", "./templates/users/me.html")
			if err != nil {
				site.ServeError(
					w,
					isApi,
					http.StatusInternalServerError,
					"Error loading template.",
					user.Id > 0,
				)
				log.Println(err)
				return
			}
			if err := tmpl.Execute(w, map[string]interface{}{
				"user":          user,
				"isLogined":     user.Id > 0,
				"createdAssets": createdAssets,
				"boughtAssets":  boughtAssets,
			}); err != nil {
				site.ServeError(
					w,
					isApi,
					http.StatusInternalServerError,
					"Error serving (executing) template.",
					user.Id > 0,
				)
				log.Println(err)
				return
			}

		}
	}
}

func UpdateBalanceInfo(strg storage.Storage) http.HandlerFunc {
	const fn = "internal.http.auth.UpdateBalanceInfo"

	return func(w http.ResponseWriter, r *http.Request) {
		isApi, ok := common.IsApiFromContext(r.Context())
		if !ok {
			http.Error(w, "Failed to get context", http.StatusInternalServerError)
			return
		}

		userId, err := auth.IdFromContext(r.Context())
		if err != nil {
			http.Error(w, "Unauthorized request", http.StatusUnauthorized)
			site.ServeError(
				w,
				isApi,
				http.StatusUnauthorized,
				"Unauthorized request.",
				false,
			)
			return
		}
		balance := &struct {
			Balance float64
		}{0}
		if isApi {
			if err := json.NewDecoder(r.Body).Decode(balance); err != nil {
				site.ServeError(
					w,
					isApi,
					http.StatusBadRequest,
					"Invalid request.",
					userId > 0,
				)
				return
			}
		} else {
			if err := r.ParseForm(); err != nil {
				site.ServeError(
					w,
					isApi,
					http.StatusInternalServerError,
					"Unable to parse form.",
					userId > 0,
				)
				return
			}
			if _, ok := r.Form["balance"]; !ok {
				site.ServeError(
					w,
					isApi,
					http.StatusBadRequest,
					"Bad form. Unable to find `balance` field.",
					userId > 0,
				)
				return
			}
			balance.Balance, err = strconv.ParseFloat(r.Form["balance"][0], 64)
			if err != nil {
				log.Printf("%s: unable to convert balance to float %v: %v", fn, r.Form["balance"][0], err)
				site.ServeError(
					w,
					isApi,
					http.StatusBadRequest,
					"Bad form. Unable to convert `balance` field.",
					userId > 0,
				)
				return
			}
		}
		if err := strg.UpdateUserBalance(balance.Balance, userId); err != nil {
			log.Printf("UpdateBalanceInfo: storage.UpdateUserBalance: %v", err)
			site.ServeError(
				w,
				isApi,
				http.StatusInternalServerError,
				"Internal Error.",
				userId > 0,
			)
			return
		}
		if isApi {
			json.NewEncoder(w).Encode(&balance)
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
		site.ServeError(
			w,
			false,
			http.StatusInternalServerError,
			"Cannot get user from context.",
			false,
		)
		return
	}
	tmpl, err := template.ParseFiles("./templates/common/base.html", "./templates/users/balance.html")
	if err != nil {
		site.ServeError(
			w,
			false,
			http.StatusInternalServerError,
			"Error loading template.",
			user.Id > 0,
		)
		log.Println(err)
		return
	}
	if err := tmpl.Execute(w, map[string]interface{}{
		"user":      user,
		"isLogined": user.Id > 0,
	}); err != nil {
		site.ServeError(
			w,
			false,
			http.StatusInternalServerError,
			"Error serving (executing) template.",
			user.Id > 0,
		)
		log.Println(err)
		return
	}
}
