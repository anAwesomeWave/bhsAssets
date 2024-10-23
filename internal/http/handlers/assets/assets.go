package assets

import (
	"bhsAssets/internal/http/middleware/auth"
	"bhsAssets/internal/http/middleware/common"
	"bhsAssets/internal/storage"
	"bhsAssets/internal/storage/models"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
)

const ASSET_CREATION_QUERY = "asset_id"

func GetAssetsCreationPage(w http.ResponseWriter, r *http.Request) {
	assetId := r.URL.Query().Get(ASSET_CREATION_QUERY)
	tmpl, err := template.ParseFiles("./templates/assets/create.html")
	if err != nil {
		http.Error(w, "Error loading template", http.StatusInternalServerError)
		log.Println(err)
		return
	}
	if err := tmpl.Execute(w, map[string]interface{}{"createdId": assetId}); err != nil {
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
		userId, err := auth.IdFromContext(r.Context())

		if err != nil {
			http.Error(w, "unauthorized request", http.StatusUnauthorized)
		}

		var newAsset models.Assets

		newAsset.CreatorId = userId

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
