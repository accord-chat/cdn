package routes

import (
	"encoding/json"
	"net/http"
	"strings"
)

func HandleUpload(w http.ResponseWriter, r *http.Request) {
	authorization := strings.Split(r.Header.Get("Authorization"), "Bearer ")[1]

	if authorization == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Missing required fields"})
		return
	}

}
