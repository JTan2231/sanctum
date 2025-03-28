package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"sanctum/utils"
)

type DeckRequest struct {
	Prompt string `json:"prompt"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type Flashcard struct {
	Pattern string `json:"pattern"`
	Match   string `json:"match"`
}

type FlashcardDeck struct {
	Cards []Flashcard `json:"cards"`
	Title string      `json:"title"`
}

const systemPrompt = `You are a helpful study aid that creates flashcard pairs in JSON format. For any topic provided, generate relevant question-answer pairs where "pattern" contains the prompt/question and "match" contains the corresponding answer. Format each flashcard as a JSON object with these exact fields:
{ "pattern": string, "match": string }

Return multiple flashcards as an array of these objects. Keep both pattern and match concise - ideally under 15 words each. Ensure the content is accurate and educational. Only respond with the JSON array, no additional text.

Example format:
[
  {
    "pattern": "What is photosynthesis?",
    "match": "Process where plants convert sunlight, water and CO2 into glucose and oxygen"
  },
  {
    "pattern": "In what year did World War II end?",
    "match": "1945"
  }
]`

func respondWithError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	response := ErrorResponse{Error: message}
	json.NewEncoder(w).Encode(response)

	log.Println("ERROR: ", response)
}

func GenerateDeckHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req DeckRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Prompt == "" {
		respondWithError(w, http.StatusBadRequest, "Prompt cannot be empty")
		return
	}

	messages := []utils.Message{
		{
			Role:    "system",
			Content: systemPrompt,
		},
		{
			Role:    "user",
			Content: req.Prompt,
		},
	}

	response, err := utils.MakeOpenAIChatRequest(messages)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error processing request")
		return
	}

	var cards []Flashcard
	if err := json.Unmarshal([]byte(response), &cards); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error parsing flashcards")
		return
	}

	if len(cards) == 0 {
		respondWithError(w, http.StatusInternalServerError, "No flashcards generated")
		return
	}

	deck := FlashcardDeck{
		Cards: cards,
		Title: req.Prompt,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(deck); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error encoding response")
		return
	}
}
