package handlers

import (
	"encoding/json"
	"net/http"
	"sanctum/utils"
)

type PromptRequest struct {
	Prompt string `json:"prompt"`
}

type PromptResponse struct {
	Suggestion string `json:"suggestion, omitempty"`
	Error      string `json:"error, omitempty"`
}

func PromptSuggestionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req PromptRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Prompt == "" {
		http.Error(w, "Prompt cannot be empty", http.StatusBadRequest)
		return
	}

	messages := []utils.Message{
		{
			Role: "system",
			Content: `You are a helpful study aid assistant that analyzes topic requests and suggests ways to make better flashcards. When given a topic, provide specific suggestions for how to refine or expand the request to generate more effective study materials.

Consider aspects like:
- Scope and specificity of the topic
- Key subtopics that might be valuable to include
- Different angles of study (definitions, applications, examples, etc.)
- Level of detail needed
- Common areas students often miss
- Ways to break down complex topics

Return suggestions as a JSON array where each suggestion is an object with a "suggestion" field. Format must be:
[
  {"suggestion": string},
  {"suggestion": string},
  ...
]

Example format:
[
  {"suggestion": "Specify which historical period of Ancient Rome you want to focus on"},
  {"suggestion": "Include both theoretical concepts and practical applications of calculus"},
  {"suggestion": "Break down 'biology' into specific systems: circulatory, respiratory, etc."}
]

Keep suggestions clear, specific, and focused on improving the flashcard learning experience. Only respond with the JSON array, no additional text.`,
		},
		{
			Role:    "user",
			Content: req.Prompt,
		},
	}

	w.Header().Set("Content-Type", "application/json")

	response, err := utils.MakeOpenAIChatRequest(messages, nil)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(PromptResponse{
			Error: "Error processing request",
		})
		return
	}

	json.NewEncoder(w).Encode(PromptResponse{
		Suggestion: response,
	})
}
