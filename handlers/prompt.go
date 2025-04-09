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
	EnhancedPrompt string `json:"enhancedPrompt,omitempty"`
	Error          string `json:"error,omitempty"`
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
			Content: `You are an AI assistant that enhances study prompts to create more effective flashcard content. 
Your task is to take the user's input prompt and expand it into a more detailed, comprehensive version.

Consider:
- Adding specific subtopics
- Including relevant terminology
- Expanding scope where beneficial
- Adding context and related concepts
- Ensuring appropriate detail level
- Breaking down complex topics into manageable parts

Respond with ONLY the enhanced prompt text. Do not include explanations or metadata.
The enhanced prompt should be a direct replacement for the original, ready to use for flashcard creation.

Example:
Input: "Ancient Rome"
Output: "Ancient Rome (753 BCE - 476 CE), including: political structure (Republic and Empire), key historical figures, major battles, social classes, cultural developments, architectural achievements, and the factors leading to its rise and fall"`,
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
		EnhancedPrompt: response,
	})
}
