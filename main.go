package main

import (
	"log"
	"net/http"
	"fmt"

	"github.com/google/uuid"

	"sanctum/database"
	"sanctum/handlers"
	"sanctum/middleware"
	"sanctum/utils"
)

func main() {
	err := database.InitDB()
	if err != nil {
		log.Fatal(err)
	}

	pc,error := utils.InitPineconeClient("sanctum2")
	if error != nil {
		log.Fatalf("Could not initialize pinecone client: %v",err)
	}

	/*--------------------TESTING---------------------------------------------------*/

	testId := uuid.New()
	card := utils.Flashcard{
		Pattern: "What city is the eiffel tower in?",
		Match:   "Paris",
		Uuid:    testId,
	}

	pc.AddCard(card)

	fmt.Println(utils.Grade(pc, testId.String(), "France"))


	/*-----------------------------------------------------------------------------------------*/
	
	http.HandleFunc("/auth", middleware.LoggingMiddleware(handlers.AuthHandler))
	http.HandleFunc("/generate-deck", middleware.LoggingMiddleware(middleware.AuthMiddleware(handlers.GenerateDeckHandler)))
	http.HandleFunc("/prompt-suggestion", middleware.LoggingMiddleware(middleware.AuthMiddleware(handlers.PromptSuggestionHandler)))

	log.Println("Server starting on localhost:8080")
	log.Fatal(http.ListenAndServe("127.0.0.1:8080", nil))
}
