package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

const CHAT_ENDPOINT string = "https://api.openai.com/v1/chat/completions"
const EMBED_ENDPOINT string = "https://api.openai.com/v1/embeddings"

func MakeOpenAIRequest[T ChatRequest | EmbedRequest](reqBody T, endpoint string) (*http.Response, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		fmt.Println("OPENAI_API_KEY environment variable not set")
		return nil, nil
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %v", err)
	}

	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %v", err)
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("api request failed with status: %v", res.StatusCode)
	}

	return res, nil
}

func GetFlashcardSchema() *ResponseFormat {
	return &ResponseFormat{
		Type: "json_schema",
		JSONSchema: JSONSchemaSpec{
			Name: "flashcards",
			Schema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"cards": map[string]any{
						"type": "array",
						"items": map[string]any{
							"type": "object",
							"properties": map[string]any{
								"pattern": map[string]any{"type": "string"},
								"match":   map[string]any{"type": "string"},
							},
							"required":             []string{"pattern", "match"},
							"additionalProperties": false,
						},
					},
				},
				"required":             []string{"cards"},
				"additionalProperties": false,
			},
		},
	}
}

func MakeOpenAIChatRequest(messages []Message, responseFormat *ResponseFormat) (string, error) {
	reqBody := ChatRequest{
		Model:          "gpt-4o",
		Messages:       messages,
		ResponseFormat: responseFormat,
	}

	res, err := MakeOpenAIRequest(reqBody, CHAT_ENDPOINT)
	if err != nil {
		return "", fmt.Errorf("error making request to OpenAI Chat endpoint: %v", err)
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	var chatResponse ChatResponse
	err = json.Unmarshal(body, &chatResponse)
	if err != nil {
		return "", fmt.Errorf("error parsing response: %v", err)
	}

	if len(chatResponse.Choices) == 0 {
		return "", fmt.Errorf("no response choices returned")
	}

	return chatResponse.Choices[0].Message.Content, nil
}

// NOTE: This needs to be guaranteed to return the embeddings in order as they were input
func MakeOpenAIEmbedRequest(text []string) (*[]EmbedData, error) {
	reqBody := EmbedRequest{
		Input: text,
		Model: "text-embedding-3-small",
	}

	res, err := MakeOpenAIRequest(reqBody, EMBED_ENDPOINT)
	if err != nil {
		return nil, fmt.Errorf("error making request to OpenAI Embed endpoint: %v", err)
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading bytes from response body: %v", err)
	}

	var embedResponse EmbedResponse
	err = json.Unmarshal(body, &embedResponse)

	if err != nil {
		return nil, fmt.Errorf("error parsing response: %v", err)
	}

	return &embedResponse.Data, nil
}
