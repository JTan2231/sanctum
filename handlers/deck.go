package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"

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

// TODO: Retry logic, error handling, etc.
func GenerateDeckHandler(w http.ResponseWriter, r *http.Request) {
	// Set headers for SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	flusher, ok := w.(http.Flusher)
	if !ok {
		respondWithError(w, http.StatusInternalServerError, "Streaming unsupported")
		return
	}

	// Helper function to send SSE updates
	sendUpdate := func(eventType string, data interface{}) {
		update, err := json.Marshal(data)
		if err != nil {
			log.Printf("Error marshaling update: %v", err)
			return
		}
		fmt.Fprintf(w, "event: %s\ndata: %s\n\n", eventType, update)
		flusher.Flush()
	}

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

	// TODO: This needs to be configurable somehow
	targetSize := 20
	var allCards []utils.Flashcard

	sendUpdate("status", map[string]interface{}{
		"message":  "Starting generation...",
		"progress": 1,
	})

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

	// TODO: I think at some point we'll want to do something like parallelize this to speed things along
	//       I don't think the embeddings have any dependencies except the cards to which they're directly linked
	for i := range initialCards {
		initialCards[i].Uuid = uuid.New().String()
	}

	err = addCardsToPinecone(initialCards)
	if err != nil {
		log.Println("Error adding initial cards to Pinecone:", err)
		respondWithError(w, http.StatusInternalServerError, "Error adding card to Pinecone")
		return
	}

	allCards = append(allCards, initialCards...)

	sendUpdate("status", map[string]interface{}{
		"message":  "Initial cards generated",
		"progress": float64(len(allCards)) / float64(targetSize) * 100,
	})

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

		// TODO: The error handling here needs to be more robust
		//       probably something like having a local backlog of un-indexed cards
		for i := range initialCards {
			initialCards[i].Uuid = uuid.New().String()
		}

		err = addCardsToPinecone(initialCards)
		if err != nil {
			log.Println("Error adding new cards to Pinecone:", err)
			respondWithError(w, http.StatusInternalServerError, "Error adding cards to Pinecone")
			return
		}

		allCards = append(allCards, initialCards...)

		// TODO: I think eventually we'll want this to be some sort of keep-alive
		fmt.Printf("%d of %d cards...", len(allCards), targetSize)

		sendUpdate("status", map[string]interface{}{
			"message":  "Initial cards generated",
			"progress": float64(len(allCards)) / float64(targetSize) * 100,
		})

		time.Sleep(time.Second / 5)
	}

	deck := utils.FlashcardDeck{
		Cards: allCards,
		Title: req.Prompt,
	}

	log.Println("Returning deck")

	sendUpdate("complete", map[string]interface{}{
		"message":  "Deck generation complete",
		"progress": 100,
		"deck": utils.FlashcardDeck{
			Cards: allCards,
			Title: req.Prompt,
		},
	})

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(deck); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error encoding final response")
		return
	}
}

func GradeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

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

	log.Println("grading with body:", gradeRequest)

	if gradeRequest.Answer == "" {
		respondWithError(w, http.StatusInternalServerError, "An answer must be provided")
		return
	}

	pc, err := utils.GetPineconeClient()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error connecting to pinecone client")
		return
	}

	numericGrade, err := utils.Grade(pc, gradeRequest.Uuid, gradeRequest.Answer)
	if err != nil {
		errMessage := fmt.Sprintf("Error grading answer: %v", err)
		respondWithError(w, http.StatusInternalServerError, errMessage)
		return
	}

	respondWithJSON(w, 200, map[string]any{
		"numericGrade": numericGrade,
	})
}

func addCardsToPinecone(cards []utils.Flashcard) error {
	for _, card := range cards {
		if card.Uuid == "" {
			return fmt.Errorf("Flashcard UUID is not set")
		}
	}

	pc, err := utils.GetPineconeClient()
	if err != nil {
		return err
	}

	_, err = pc.AddCards(cards)
	if err != nil {
		return err
	}

	return nil
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
		respondWithError(w, http.StatusInternalServerError, "Error unmarshaling request")
		return
	}

	card.Uuid = uuid.New().String()

	err = addCardsToPinecone([]utils.Flashcard{card})
	if err != nil {
		log.Println("Error adding card to Pinecone:", err)
		respondWithError(w, http.StatusInternalServerError, "Error adding card to Pinecone")
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

	pc, err := utils.GetPineconeClient()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error connecting to pinecone client")
		return
	}

	_, err = pc.RemoveCard(card.Uuid)
	if err != nil {
		errMessage := fmt.Sprintf("Error removing card: %v", err)
		respondWithError(w, http.StatusInternalServerError, errMessage)
		return
	}

	respondWithJSON(w, 200, map[string]string{"message": "Card removed successfully"})
}
