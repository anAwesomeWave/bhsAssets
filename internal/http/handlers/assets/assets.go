package assets

import (
	"bhsAssets/internal/http/handlers/site"
	mauth "bhsAssets/internal/http/middleware/auth"
	"bhsAssets/internal/http/middleware/common"
	"bhsAssets/internal/storage"
	"bhsAssets/internal/storage/models"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"html/template"
	"log"
	"net/http"
	"strconv"
)

const ASSET_CREATION_QUERY = "asset_id"

func GetAssetsCreationPage(w http.ResponseWriter, r *http.Request) {
	user, ok := mauth.FromContext(r.Context())
	// maybe add info about assets
	if !ok {
		http.Error(w, "User not found in context", http.StatusUnauthorized)
		return
	}
	assetId := r.URL.Query().Get(ASSET_CREATION_QUERY)
	tmpl, err := template.ParseFiles("./templates/common/base.html", "./templates/assets/create.html")
	if err != nil {
		http.Error(w, "Error loading template", http.StatusInternalServerError)
		log.Println(err)
		return
	}
	if err := tmpl.Execute(w, map[string]interface{}{"isLogined": user != nil && user.Id > 0, "createdId": assetId}); err != nil {
		http.Error(w, "Error templating", http.StatusInternalServerError)
		log.Println(err)
		return
	}
}

func CreateAsset(strg storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		isApi, ok := common.IsApiFromContext(r.Context())
		if !ok {
			http.Error(w, "Failed to get context", http.StatusInternalServerError)
		}
		userId, err := mauth.IdFromContext(r.Context())

		if err != nil {
			http.Error(w, "unauthorized request", http.StatusUnauthorized)
		}

		var newAsset models.Assets

		if isApi {
			if err := json.NewDecoder(r.Body).Decode(&newAsset); err != nil {
				http.Error(w, "Invalid request", http.StatusBadRequest)
				return
			}
		} else {
			if err := r.ParseForm(); err != nil {
				http.Error(w, "Unable to parse form", http.StatusInternalServerError)
				return
			}
			if _, ok := r.Form["name"]; !ok {
				http.Error(w, "Bad form. Unable to find `name` field", http.StatusBadRequest)
				return
			}
			if _, ok := r.Form["description"]; !ok {
				http.Error(w, "Bad form. Unable to find `description` field", http.StatusBadRequest)
				return
			}
			if _, ok := r.Form["price"]; !ok {
				http.Error(w, "Bad form. Unable to find `price` field", http.StatusBadRequest)
				return
			}
			newAsset.Name = r.Form["name"][0]
			newAsset.Description = r.Form["description"][0]
			newAsset.Price, err = strconv.ParseFloat(r.Form["price"][0], 64)
			if err != nil {
				http.Error(w, "Bad form. `price` field is not convertable to float64", http.StatusBadRequest)
				return
			}
		}
		if userId < 0 || newAsset.Name == "" {
			http.Error(w, "Bad data. Check if asset name is not empty and you are authenticated", http.StatusBadRequest)
			return
		}

		newAsset.CreatorId = userId
		id, err := strg.CreateAsset(&newAsset)
		if err != nil {
			log.Println(err)
			http.Error(w, "Bad request, asset cannot be created", http.StatusBadRequest)
			return

		}
		http.Redirect(
			w,
			r,
			fmt.Sprintf("/assets/create?%s=%d", ASSET_CREATION_QUERY, id),
			http.StatusSeeOther,
		)
	}
}

func GetAsset(strg storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		userId, authErr := mauth.IdFromContext(r.Context())
		if authErr != nil && !errors.Is(authErr, mauth.UnauthorizedErr) {
			http.Error(w, "Auth token internal error", http.StatusInternalServerError)
			return
		}

		assetId, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
		if err != nil {
			site.NotFoundHandler(w, r)
			return
		}
		asset, err := strg.GetAssetById(assetId)
		if err != nil {
			site.NotFoundHandler(w, r)
			return
		}
		isApi, ok := common.IsApiFromContext(r.Context())
		if !ok {
			http.Error(w, "Failed to get context", http.StatusInternalServerError)
		}
		if isApi {
			json.NewEncoder(w).Encode(asset)
		} else {
			tmpl, err := template.ParseFiles("./templates/common/base.html", "./templates/assets/get.html")
			if err != nil {
				http.Error(w, "Error loading template", http.StatusInternalServerError)
				log.Println(err)
				return
			}

			creator, err := strg.GetUserById(asset.CreatorId)
			if err != nil {
				log.Println(w, fmt.Sprintf("creator with id %v not found for asset %v: %v", asset.Id, asset, err))
				http.Error(w, "Creator Not Found", http.StatusInternalServerError)
				return
			}

			if err := tmpl.Execute(w, map[string]interface{}{
				"asset":     asset,
				"creator":   creator,
				"isCreator": authErr == nil && creator.Id == userId,
				"isLogined": authErr == nil && userId > 0,
			}); err != nil {
				http.Error(w, "Error templating", http.StatusInternalServerError)
				log.Println(err)
				return
			}

		}
	}
}

func DeleteAsset(strg storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		userId, authErr := mauth.IdFromContext(r.Context())
		if authErr != nil && !errors.Is(authErr, mauth.UnauthorizedErr) {
			http.Error(w, "Auth token internal error", http.StatusInternalServerError)
			return
		}

		assetId, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
		if err != nil {
			site.NotFoundHandler(w, r)
			return
		}
		asset, err := strg.GetAssetById(assetId)
		if err != nil {
			site.NotFoundHandler(w, r)
			return
		}

		if asset.CreatorId != userId {
			http.Error(w, "forbidden. You are not the creator of the asset", http.StatusForbidden)
			return
		}
		deletionErr := strg.DeleteAsset(assetId)

		isApi, ok := common.IsApiFromContext(r.Context())
		if !ok {
			http.Error(w, "Failed to get context", http.StatusInternalServerError)
		}
		if isApi {
			if deletionErr != nil {
				if errors.Is(deletionErr, storage.ErrNotFound) {
					w.WriteHeader(http.StatusNotFound)
					json.NewEncoder(w).Encode(map[string]string{"status": "404"})
					return
				}
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]string{"status": "500"})
				return
			}
			json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
		} else {
			tmpl, err := template.ParseFiles("./templates/common/base.html", "./templates/assets/get.html")
			if err != nil {
				http.Error(w, "Error loading template", http.StatusInternalServerError)
				log.Println(err)
				return
			}

			creator, err := strg.GetUserById(asset.CreatorId)
			if err != nil {
				log.Println(w, fmt.Sprintf("creator with id %v not found for asset %v: %v", asset.Id, asset, err))
				http.Error(w, "Creator Not Found", http.StatusInternalServerError)
				return
			}

			if err := tmpl.Execute(w, map[string]interface{}{
				"asset":     asset,
				"creator":   creator,
				"isCreator": authErr == nil && creator.Id == userId,
				"isLogined": authErr == nil && userId > 0,
			}); err != nil {
				http.Error(w, "Error templating", http.StatusInternalServerError)
				log.Println(err)
				return
			}

		}
	}
}

func GetAllAssets(strg storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const fn = "assets.GetAllAssets"
		log.Println("STARTING ALL ASSETS GETTER")
		userId, authErr := mauth.IdFromContext(r.Context())
		if authErr != nil && !errors.Is(authErr, mauth.UnauthorizedErr) {
			http.Error(w, "Auth token internal error", http.StatusInternalServerError)
			return
		}
		isApi, ok := common.IsApiFromContext(r.Context())
		if !ok {
			http.Error(w, "Failed to get context", http.StatusInternalServerError)
		}
		filters := make(map[string]string, 3)
		if nm := r.URL.Query().Get("name"); nm != "" {
			filters["name"] = nm
		}
		if mp := r.URL.Query().Get("min_price"); mp != "" {
			filters["min_price"] = mp
		}
		if maxp := r.URL.Query().Get("max_price"); maxp != "" {
			filters["max_price"] = maxp
		}
		assets, err := strg.GetAllAssetsFiltered(filters)
		if err != nil {
			log.Printf("%s: error getting list of assets: %v\n", fn, err)
			http.Error(w, "error getting list of assets", http.StatusInternalServerError)
			return
		}
		if isApi {
			json.NewEncoder(w).Encode(assets)
		} else {
			tmpl, err := template.ParseFiles("./templates/common/base.html", "./templates/assets/getAll.html")
			if err != nil {
				http.Error(w, "Error loading template", http.StatusInternalServerError)
				log.Println(err)
				return
			}
			if err := tmpl.Execute(w, map[string]interface{}{
				"assets":    assets,
				"isLogined": authErr == nil && userId > 0,
			}); err != nil {
				http.Error(w, "Error templating", http.StatusInternalServerError)
				log.Println(err)
				return
			}

		}
	}
}

func BuyAsset(strg storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := mauth.FromContext(r.Context())
		if !ok {
			http.Error(w, "User not found in context", http.StatusUnauthorized)
			return
		}

		assetId, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
		if err != nil {
			site.NotFoundHandler(w, r)
			return
		}

		asset, err := strg.GetAssetById(assetId)
		if err != nil {
			site.NotFoundHandler(w, r)
			return
		}

		creator, err := strg.GetUserById(asset.CreatorId)
		if err != nil {
			http.Error(
				w,
				"We are sorry, but... Creator for this asset was not found. You cannot purchase this asset",
				http.StatusInternalServerError,
			)
			return
		}

		if user.Id == creator.Id {
			http.Error(
				w,
				"You cannot buy this asset since you are the creator of this asset",
				http.StatusForbidden,
			)
			return
		}
		if user.Balance < asset.Price {
			http.Error(
				w,
				"You don't have enough money to buy this asset.",
				http.StatusConflict,
			)
			return
		}

		isApi, ok := common.IsApiFromContext(r.Context())
		if !ok {
			http.Error(w, "Failed to get context", http.StatusInternalServerError)
		}

		buyAssetErr := strg.BuyAsset(user, asset, creator)
		status := http.StatusCreated
		data := "ok"
		if buyAssetErr != nil {
			log.Println(err)
			status = http.StatusBadRequest
			data = "cannot process request"
		}
		if isApi {
			w.WriteHeader(status)
			json.NewEncoder(w).Encode(map[string]string{"status": data})
		} else {
			if buyAssetErr != nil {
				http.Error(w, data, status)
				return
			}
			http.Redirect(w, r, "/users/me", http.StatusSeeOther)
		}
	}
}
