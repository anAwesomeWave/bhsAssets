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
		site.ServeError(w, false, http.StatusUnauthorized, "User not found in context", false)
		return
	}
	assetId := r.URL.Query().Get(ASSET_CREATION_QUERY)
	tmpl, err := template.ParseFiles("./templates/common/base.html", "./templates/assets/create.html")
	if err != nil {
		site.ServeError(
			w,
			false,
			http.StatusInternalServerError,
			"Error loading template",
			user != nil && user.Id > 0,
		)
		log.Println(err)
		return
	}
	if err := tmpl.Execute(w, map[string]interface{}{"isLogined": user != nil && user.Id > 0, "createdId": assetId}); err != nil {
		site.ServeError(
			w,
			false,
			http.StatusInternalServerError,
			"Error serving (executing) template",
			user != nil && user.Id > 0,
		)
		log.Println(err)
		return
	}
}

func CreateAsset(strg storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		isApi, ok := common.IsApiFromContext(r.Context())
		if !ok {
			http.Error(w, "Failed to get context", http.StatusInternalServerError)
			return
		}
		userId, err := mauth.IdFromContext(r.Context())

		if err != nil {
			site.ServeError(
				w,
				isApi,
				http.StatusUnauthorized,
				"unauthorized request",
				false,
			)
			return
		}

		var newAsset models.Assets

		if isApi {
			if err := json.NewDecoder(r.Body).Decode(&newAsset); err != nil {
				site.ServeError(
					w,
					isApi,
					http.StatusBadRequest,
					"Invalid request",
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
					"Unable to parse form",
					userId > 0,
				)
				return
			}
			if _, ok := r.Form["name"]; !ok {
				site.ServeError(
					w,
					isApi,
					http.StatusBadRequest,
					"Bad form. Unable to find `name` field",
					userId > 0,
				)
				return
			}
			if _, ok := r.Form["description"]; !ok {
				site.ServeError(
					w,
					isApi,
					http.StatusBadRequest,
					"Bad form. Unable to find `description` field",
					userId > 0,
				)
				return
			}
			if _, ok := r.Form["price"]; !ok {
				site.ServeError(
					w,
					isApi,
					http.StatusBadRequest,
					"Bad form. Unable to find `price` field",
					userId > 0,
				)
				return
			}
			newAsset.Name = r.Form["name"][0]
			newAsset.Description = r.Form["description"][0]
			newAsset.Price, err = strconv.ParseFloat(r.Form["price"][0], 64)
			if err != nil {
				site.ServeError(
					w,
					isApi,
					http.StatusBadRequest,
					"Bad form. `price` field is not convertable to float64",
					userId > 0,
				)
				return
			}
		}
		if userId < 0 || newAsset.Name == "" {
			site.ServeError(
				w,
				isApi,
				http.StatusBadRequest,
				"Bad data. Check if asset name is not empty and you are authenticated.",
				userId > 0,
			)
			return
		}

		newAsset.CreatorId = userId
		id, err := strg.CreateAsset(&newAsset)
		if err != nil {
			log.Println(err)
			site.ServeError(
				w,
				isApi,
				http.StatusBadRequest,
				"Bad request, asset cannot be created.",
				userId > 0,
			)
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
		isApi, ok := common.IsApiFromContext(r.Context())
		if !ok {
			http.Error(w, "Failed to get context", http.StatusInternalServerError)
			return
		}

		userId, authErr := mauth.IdFromContext(r.Context())
		if authErr != nil && !errors.Is(authErr, mauth.UnauthorizedErr) {
			site.ServeError(
				w,
				isApi,
				http.StatusInternalServerError,
				"Auth token internal error",
				false,
			)
			return
		}

		assetId, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
		if err != nil {
			site.ServeError(
				w,
				isApi,
				http.StatusNotFound,
				"Asset was not found. Asset Id is incorrect. (non-int).",
				userId > 0,
			)
			return
		}
		asset, err := strg.GetAssetById(assetId)
		if err != nil {
			site.ServeError(
				w,
				isApi,
				http.StatusNotFound,
				"Asset was not found",
				userId > 0,
			)
			return
		}

		if isApi {
			json.NewEncoder(w).Encode(asset)
		} else {
			tmpl, err := template.ParseFiles("./templates/common/base.html", "./templates/assets/get.html")
			if err != nil {
				site.ServeError(
					w,
					isApi,
					http.StatusInternalServerError,
					"Error loading template",
					userId > 0,
				)
				log.Println(err)
				return
			}

			creator, err := strg.GetUserById(asset.CreatorId)
			if err != nil {
				log.Println(w, fmt.Sprintf("creator with id %v not found for asset %v: %v", asset.Id, asset, err))
				site.ServeError(
					w,
					isApi,
					http.StatusInternalServerError,
					"Asset Creator Was Not Found.",
					userId > 0,
				)
				return
			}

			if err := tmpl.Execute(w, map[string]interface{}{
				"asset":     asset,
				"creator":   creator,
				"isCreator": authErr == nil && creator.Id == userId,
				"isLogined": authErr == nil && userId > 0,
			}); err != nil {
				site.ServeError(
					w,
					isApi,
					http.StatusInternalServerError,
					"Error serving (executing) template.",
					userId > 0,
				)
				log.Println(err)
				return
			}

		}
	}
}

func DeleteAsset(strg storage.Storage) http.HandlerFunc {
	const fn = "assets.DeleteAsset"

	return func(w http.ResponseWriter, r *http.Request) {
		isApi, ok := common.IsApiFromContext(r.Context())
		if !ok {
			http.Error(w, "Failed to get context", http.StatusInternalServerError)
			return
		}
		userId, authErr := mauth.IdFromContext(r.Context())
		if authErr != nil && !errors.Is(authErr, mauth.UnauthorizedErr) {
			site.ServeError(
				w,
				isApi,
				http.StatusInternalServerError,
				"Auth token internal error.",
				false,
			)
			return
		}

		assetId, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
		if err != nil {
			site.ServeError(
				w,
				isApi,
				http.StatusNotFound,
				fmt.Sprintf("Asset with id %s was not found.", chi.URLParam(r, "id")),
				userId > 0,
			)
			return
		}
		asset, err := strg.GetAssetById(assetId)
		if err != nil {
			site.ServeError(
				w,
				isApi,
				http.StatusNotFound,
				fmt.Sprintf("Asset with id %d was not found.", assetId),
				userId > 0,
			)
			return
		}

		if asset.CreatorId != userId {
			site.ServeError(
				w,
				isApi,
				http.StatusForbidden,
				"forbidden. You are not the creator of the asset",
				userId > 0,
			)
			return
		}
		if deletionErr := strg.DeleteAsset(assetId); deletionErr != nil {
			if errors.Is(deletionErr, storage.ErrNotFound) {
				site.ServeError(
					w,
					isApi,
					http.StatusNotFound,
					fmt.Sprintf("Asset with id %d was not found.", assetId),
					userId > 0,
				)
				return
			}
			site.ServeError(
				w,
				isApi,
				http.StatusInternalServerError,
				"Internal server error occur, while deleting asset.",
				userId > 0,
			)
			log.Printf("%s: error with deleting asset %d: %v\n", fn, assetId, deletionErr)
			return
		}
		if isApi {
			json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
		} else {
			http.Redirect(w, r, "/users/me", http.StatusSeeOther)
		}
	}
}

func GetAllAssets(strg storage.Storage) http.HandlerFunc {
	const fn = "assets.GetAllAssets"

	return func(w http.ResponseWriter, r *http.Request) {
		isApi, ok := common.IsApiFromContext(r.Context())
		if !ok {
			http.Error(w, "Failed to get context", http.StatusInternalServerError)
			return
		}

		userId, authErr := mauth.IdFromContext(r.Context())
		if authErr != nil && !errors.Is(authErr, mauth.UnauthorizedErr) {
			site.ServeError(
				w,
				isApi,
				http.StatusInternalServerError,
				"Auth token internal error",
				false,
			)
			return
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
			site.ServeError(
				w,
				isApi,
				http.StatusInternalServerError,
				"error getting list of assets",
				userId > 0,
			)
			return
		}
		if isApi {
			json.NewEncoder(w).Encode(assets)
		} else {
			tmpl, err := template.ParseFiles("./templates/common/base.html", "./templates/assets/getAll.html")
			if err != nil {
				site.ServeError(
					w,
					isApi,
					http.StatusInternalServerError,
					"Error loading template",
					userId > 0,
				)
				log.Println(err)
				return
			}
			if err := tmpl.Execute(w, map[string]interface{}{
				"assets":    assets,
				"isLogined": authErr == nil && userId > 0,
			}); err != nil {
				site.ServeError(
					w,
					isApi,
					http.StatusInternalServerError,
					"Error serving (executing) templating",
					userId > 0,
				)
				log.Printf("%s: %v\n", fn, err)
				return
			}

		}
	}
}

func BuyAsset(strg storage.Storage) http.HandlerFunc {
	const fn = "assets.BuyAsset"
	return func(w http.ResponseWriter, r *http.Request) {
		isApi, ok := common.IsApiFromContext(r.Context())
		if !ok {
			http.Error(w, "Failed to get context", http.StatusInternalServerError)
			return
		}
		user, ok := mauth.FromContext(r.Context())
		if !ok {
			site.ServeError(
				w,
				isApi,
				http.StatusUnauthorized,
				"User not found in context",
				false,
			)
			return
		}

		assetId, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
		if err != nil {
			site.ServeError(
				w,
				isApi,
				http.StatusNotFound,
				fmt.Sprintf("Asset with id %s was not found.", chi.URLParam(r, "id")),
				user.Id > 0,
			)
			return
		}

		asset, err := strg.GetAssetById(assetId)
		if err != nil {
			site.ServeError(
				w,
				isApi,
				http.StatusNotFound,
				fmt.Sprintf("Asset with id %d was not found.", assetId),
				user.Id > 0,
			)
			return
		}

		creator, err := strg.GetUserById(asset.CreatorId)
		if err != nil {
			site.ServeError(
				w,
				isApi,
				http.StatusInternalServerError,
				"We are sorry, but... Creator for this asset was not found. You cannot purchase this asset",
				user.Id > 0,
			)
			return
		}

		if user.Id == creator.Id {
			site.ServeError(
				w,
				isApi,
				http.StatusForbidden,
				"You cannot buy this asset since you are the creator of this asset",
				user.Id > 0,
			)
			return
		}
		if user.Balance < asset.Price {
			site.ServeError(
				w,
				isApi,
				http.StatusConflict,
				"You don't have enough money to buy this asset.",
				user.Id > 0,
			)
			return
		}

		if buyAssetErr := strg.BuyAsset(user, asset, creator); buyAssetErr != nil {
			site.ServeError(
				w,
				isApi,
				http.StatusBadRequest,
				"Cannot process request. Error occured",
				user.Id > 0,
			)
			log.Printf("%s: error process transaction assetId-%d, userId-%d, creatorId-%d, : %v", fn, assetId, user.Id, creator.Id, buyAssetErr)
		}
		if isApi {
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
		} else {
			http.Redirect(w, r, "/users/me", http.StatusSeeOther)
		}
	}
}
