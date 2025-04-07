package main

import (
	"log"
	"net/http"

	"sanctum/handlers"
	"sanctum/middleware"
)

func main() {
	http.HandleFunc("/grade", middleware.LoggingMiddleware(middleware.AuthMiddleware(handlers.GradeHandler)))
	http.HandleFunc("/add-card", middleware.LoggingMiddleware(middleware.AuthMiddleware(handlers.AddCardHandler)))
	http.Handle("/remove-card", middleware.LoggingMiddleware(middleware.AuthMiddleware(handlers.RemoveCardHandler)))

	http.HandleFunc("/auth", middleware.LoggingMiddleware(handlers.AuthHandler))
	http.HandleFunc("/generate-deck", middleware.LoggingMiddleware(middleware.AuthMiddleware(handlers.GenerateDeckHandler)))
	http.HandleFunc("/prompt-suggestion", middleware.LoggingMiddleware(middleware.AuthMiddleware(handlers.PromptSuggestionHandler)))

	log.Println("Server starting on localhost:8080")
	log.Fatal(http.ListenAndServe("0.0.0.0:8080", nil))
}
