package handlers

import (
	"net/http"
)

func GenerateDeckHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// TODO: Implement external API call
	w.Write([]byte("Generate deck endpoint"))
}
