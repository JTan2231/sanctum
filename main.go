package main

import (
	"log"
	"net/http"
	"sanctum/database"
	"sanctum/handlers"
	"sanctum/middleware"
)

func main() {
	err := database.InitDB()
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/auth", middleware.LoggingMiddleware(handlers.AuthHandler))
	http.HandleFunc("/generate-deck", middleware.LoggingMiddleware(middleware.AuthMiddleware(handlers.GenerateDeckHandler)))
	http.HandleFunc("/prompt-suggestion", middleware.LoggingMiddleware(middleware.AuthMiddleware(handlers.PromptSuggestionHandler)))

	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
