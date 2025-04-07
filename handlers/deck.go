package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"sanctum/utils"
)

const systemPrompt = `You are a helpful study aid that creates flashcard pairs in JSON format. For any topic provided, generate relevant question-answer pairs where "pattern" contains the prompt/question and "match" contains the corresponding answer. Format each flashcard as a JSON object with these exact fields:
{ "pattern": string, "match": string }

Return multiple flashcards as a plain array of these objects - do not wrap in any additional object. Ensure the content is accurate and educational. Only respond with the JSON array, no additional text.

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

	targetSize := 20 // Same as bash script
	var allCards []utils.Flashcard

	// Initial request to generate first set of cards
	initialMessages := []utils.Message{
		{
			Role:    "system",
			Content: systemPrompt,
		},
		{
			Role:    "user",
			Content: fmt.Sprintf("Generate 2-3 flashcards about %s.", req.Prompt),
		},
	}

	var initialCards []utils.Flashcard

	response, err := utils.MakeOpenAIChatRequest(initialMessages, utils.GetFlashcardSchema())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error processing initial request")
		return
	}

	var rawResponse any
	if err := json.Unmarshal([]byte(response), &rawResponse); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error parsing initial flashcards")
		return
	}

	fmt.Printf("Raw response: %+v\n", rawResponse)

	if responseMap, ok := rawResponse.(map[string]any); ok {
		if cardsRaw, ok := responseMap["cards"].([]any); ok {
			cardsJSON, err := json.Marshal(cardsRaw)
			if err != nil {
				respondWithError(w, http.StatusInternalServerError, "Error re-encoding cards")
				return
			}

			if err := json.Unmarshal(cardsJSON, &initialCards); err != nil {
				respondWithError(w, http.StatusInternalServerError, "Error parsing cards array")
				return
			}
		}
	}

	allCards = append(allCards, initialCards...)

	for len(allCards) < targetSize {
		currentDeckJSON, err := json.Marshal(allCards)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Error encoding current deck")
			return
		}

		expansionPrompt := fmt.Sprintf(`Here is my current deck of flashcards:

%s

Please generate 2-3 additional flashcards that expand the knowledge covered by this deck. Focus on related but new concepts.`, string(currentDeckJSON))

		messages := []utils.Message{
			{
				Role:    "system",
				Content: systemPrompt,
			},
			{
				Role:    "user",
				Content: expansionPrompt,
			},
		}

		response, err := utils.MakeOpenAIChatRequest(messages, utils.GetFlashcardSchema())
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Error processing expansion request")
			return
		}

		var rawResponse any
		if err := json.Unmarshal([]byte(response), &rawResponse); err != nil {
			respondWithError(w, http.StatusInternalServerError, "Error parsing initial flashcards")
			return
		}

		if responseMap, ok := rawResponse.(map[string]any); ok {
			if cardsRaw, ok := responseMap["cards"].([]any); ok {
				cardsJSON, err := json.Marshal(cardsRaw)
				if err != nil {
					respondWithError(w, http.StatusInternalServerError, "Error re-encoding cards")
					return
				}

				if err := json.Unmarshal(cardsJSON, &initialCards); err != nil {
					respondWithError(w, http.StatusInternalServerError, "Error parsing cards array")
					return
				}
			}
		}

		allCards = append(allCards, initialCards...)

		fmt.Printf("%d of %d cards...", len(allCards), targetSize)

		time.Sleep(time.Second / 5)
	}

	deck := utils.FlashcardDeck{
		Cards: allCards,
		Title: req.Prompt,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(deck); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error encoding final response")
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

	numericGrade, err := utils.Grade(pc, gradeRequest.Uuid.String(), gradeRequest.Answer)
	if err != nil {
		errMessage := fmt.Sprintf("Error grading answer: %v", err)
		respondWithError(w, http.StatusInternalServerError, errMessage)
		return
	}

	respondWithJSON(w, 200, map[string]any{
		"numericGrade": numericGrade,
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
