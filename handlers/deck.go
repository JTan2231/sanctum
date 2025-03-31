package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"sanctum/utils"
)

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

	response := utils.ErrorResponse{Error: message}
	json.NewEncoder(w).Encode(response)

	log.Println("ERROR: ", response)
}

func respondWithJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(payload)
}

func GenerateDeckHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req utils.DeckRequest
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

	var cards []utils.Flashcard
	if err := json.Unmarshal([]byte(response), &cards); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error parsing flashcards")
		return
	}

	if len(cards) == 0 {
		respondWithError(w, http.StatusInternalServerError, "No flashcards generated")
		return
	}

	deck := utils.FlashcardDeck{
		Cards: cards,
		Title: req.Prompt,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(deck); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error encoding response")
		return
	}
}

func GradeHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error decoding request")
		return
	}

	var gradeRequest utils.GradeRequest
	err = json.Unmarshal(body, &gradeRequest)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error decoding request")
		return
	}

	if gradeRequest.Answer == "" {
		respondWithError(w, http.StatusInternalServerError, "An answer must be provided")
		return
	}

	pc, err := utils.InitPineconeClient("sanctum2")
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error connecting to pinecone client")
		return
	}

	numericGrade,letterGrade,err := utils.Grade(pc,gradeRequest.Uuid.String(),gradeRequest.Answer)
	if err != nil {
		errMessage := fmt.Sprintf("Error grading answer: %v", err)
		respondWithError(w, http.StatusInternalServerError, errMessage)
		return
	}

	respondWithJSON(w, 200, map[string]interface{}{
		"letterGrade" :  letterGrade,
		"numericGrade" : numericGrade,
	})
}

func AddCardHandler(w http.ResponseWriter, r *http.Request) {

	defer r.Body.Close()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error decoding request")
		return
	}

	var card utils.Flashcard
	err = json.Unmarshal(body, &card)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error decoding request")
		return
	}

	pc, err := utils.InitPineconeClient("sanctum2")
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error connecting to pinecone client")
		return
	}

	_, err = pc.AddCard(card)
	if err != nil {
		errMessage := fmt.Sprintf("Error adding card to database: %v", err)
		respondWithError(w, http.StatusInternalServerError, errMessage)
		return
	}

	respondWithJSON(w, 200, map[string]string{"message": "Card added successfully"})
}

func RemoveCardHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error decoding request")
		return
	}

	var card utils.Flashcard
	err = json.Unmarshal(body, &card)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error decoding request")
		return
	}

	pc, err := utils.InitPineconeClient("sanctum2")
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error connecting to pinecone client")
		return
	}

	_, err = pc.RemoveCard(card.Uuid.String())
	if err != nil {
		errMessage := fmt.Sprintf("Error removing card: %v", err)
		respondWithError(w, http.StatusInternalServerError, errMessage)
		return
	}

	respondWithJSON(w, 200, map[string]string{"message": "Card removed successfully"})
}