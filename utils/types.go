package utils

import (
	"context"

	"github.com/google/uuid"
	"github.com/pinecone-io/go-pinecone/v3/pinecone"
)

/*-----------------------------------------------------*/

type DeckRequest struct {
	Prompt string `json:"prompt"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type Flashcard struct {
	Pattern string     `json:"pattern"`
	Match   string     `json:"match"`
	Uuid    uuid.UUID  `json:"uuid"` 
}

type FlashcardDeck struct {
	Cards []Flashcard `json:"cards"`
	Title string      `json:"title"`
}

/*-----------------------------------------------------*/

type PineconeAPIKey     string
type PineconeNameSpace  string

type PineconeClient struct {
	Ctx  	   context.Context
	Client     *pinecone.Client
	Index      *pinecone.IndexConnection
}

type IndexMetrics struct {
	VectorCount  int
	Dimension    int
}

/*-----------------------------------------------------*/

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

type Choice struct {
	Message struct {
		Content string `json:"content"`
	} `json:"message"`
}

type ChatResponse struct {
	Choices []Choice `json:"choices"`
}

/*-----------------------------------------------------*/