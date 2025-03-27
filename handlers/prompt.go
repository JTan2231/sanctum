package handlers

import (
	"net/http"
)

func PromptSuggestionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// TODO: Implement external API call
	w.Write([]byte("Prompt suggestion endpoint"))
}
