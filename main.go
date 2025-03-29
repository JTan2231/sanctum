package main

import (
	"log"
	"net/http"
	"fmt"

	"sanctum/database"
	"sanctum/handlers"
	"sanctum/middleware"
	"sanctum/utils"

	"github.com/google/uuid"
)

func main() {
	err := database.InitDB()
	if err != nil {
		log.Fatal(err)
	}

	/*--------------------TESTING PC WRAPPER---------------------------------------------------*/

	pc,error := utils.InitPineconeClient("sanctum2")
	if error != nil {
		log.Fatalf("Could not initialize pinecone client: %v",err)
	}

	fmt.Println(pc.AddCard(
		utils.Flashcard{
			Uuid: 	 uuid.New(),
			Pattern: "What city is the Eiffel Tower in",
			Match:   "Paris",
		},
	))

	metrics, _ := pc.IndexMetrics()
	fmt.Println(metrics)

	/*-----------------------------------------------------------------------------------------*/
	
	http.HandleFunc("/auth", middleware.LoggingMiddleware(handlers.AuthHandler))
	http.HandleFunc("/generate-deck", middleware.LoggingMiddleware(middleware.AuthMiddleware(handlers.GenerateDeckHandler)))
	http.HandleFunc("/prompt-suggestion", middleware.LoggingMiddleware(middleware.AuthMiddleware(handlers.PromptSuggestionHandler)))

	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
