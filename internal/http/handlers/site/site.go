package site

import (
	mauth "bhsAssets/internal/http/middleware/auth"
	"bhsAssets/internal/http/middleware/common"
	"encoding/json"
	"html/template"
	"log"
	"net/http"
)

func NotFoundHandler(w http.ResponseWriter, r *http.Request) {
	isApi, ok := common.IsApiFromContext(r.Context())
	if !ok {
		http.Error(w, "Failed to get context", http.StatusInternalServerError)
	}
	if isApi {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "404"})
		return
	} else {
		tmpl, err := template.ParseFiles("./templates/common/base.html", "./templates/common/404_not_found.html")

		if err != nil {
			http.Error(w, "Error loading template", http.StatusInternalServerError)
			log.Println(err)
			return
		}
		userId, authErr := mauth.IdFromContext(r.Context())

		if err := tmpl.Execute(w, map[string]interface{}{"isLogined": authErr == nil && userId > 0}); err != nil {
			http.Error(w, "Error templating", http.StatusInternalServerError)
			log.Println(err)
			return
		}
	}
}
