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

	http.HandleFunc("/grade", middleware.LoggingMiddleware(middleware.AuthMiddleware(handlers.GradeHandler)))
	http.HandleFunc("/add-card", middleware.LoggingMiddleware(middleware.AuthMiddleware(handlers.AddCardHandler)))
	http.Handle("/remove-card", middleware.LoggingMiddleware(middleware.AuthMiddleware(handlers.RemoveCardHandler)))
	
	http.HandleFunc("/auth", middleware.LoggingMiddleware(handlers.AuthHandler))
	http.HandleFunc("/generate-deck", middleware.LoggingMiddleware(middleware.AuthMiddleware(handlers.GenerateDeckHandler)))
	http.HandleFunc("/prompt-suggestion", middleware.LoggingMiddleware(middleware.AuthMiddleware(handlers.PromptSuggestionHandler)))

	log.Println("Server starting on localhost:8080")
	log.Fatal(http.ListenAndServe("127.0.0.1:8080", nil))
}
