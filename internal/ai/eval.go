package ai

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/codehia/goflash/internal/types"
	"go.uber.org/zap"
)

var (
	logger     = zap.Must(zap.NewDevelopment())
	sugar      = logger.Sugar()
	httpClient = &http.Client{Timeout: 72 * time.Second}
)

const (
	systemPrompt = `
    	You are a strict but fair technical interviewer evaluating a candidate's
    	answer to a system design question.

    	You will receive the question, the reference answer, and the candidate's
    	answer. Your job is to score the candidate's answer and provide feedback.

    	SCORING RULES:
    	- Score out of 10. Be strict: a 10 requires the candidate to have covered
    	  every key point in the reference answer with accuracy.
    	- Deduct points for missing concepts, incorrect statements, and vague
    	  answers that lack concrete detail.
    	- Do not award points for padding or restating the question.
	
    	FEEDBACK RULES:
    	- Keep feedback to 2-3 sentences maximum.
    	- Be specific: name the exact concepts that were missing or wrong.
    	- If the answer is good, say what made it good in one sentence, then
    	  name what was missing.
	
    	OUTPUT RULES:
    	- Respond ONLY with valid JSON. No preamble, no markdown fences.
    	- Match the response shape exactly.
	
    	RESPONSE SHAPE:
    	{
    	  "score": <integer 0-10>,
    	  "feedback": "<2-3 sentences, specific and actionable>"
    	}
	`
)

func createSystemPrompt() (types.APIMessage, error) {
	return types.NewMessage(systemPrompt, "system")
}

func createUserPrompt(question, correctAnswer, userAnswer string) (types.APIMessage, error) {
	content := fmt.Sprintf(`
		Question: %s
		Reference answer: %s
		Candidate answer: %s`, question, correctAnswer, userAnswer)
	return types.NewMessage(content, "user")
}

func createPayload(question, correctAnswer, userAnswer string) (types.RequestPayload, error) {
	systemPromptMessage, err := createSystemPrompt()
	if err != nil {
		return types.RequestPayload{}, err
	}

	userPromptMessage, err := createUserPrompt(question, correctAnswer, userAnswer)
	if err != nil {
		return types.RequestPayload{}, err
	}

	promptMessages := []types.APIMessage{systemPromptMessage, userPromptMessage}
	return types.NewRequestPayload(promptMessages)
}

func makeRequest(payloadData types.RequestPayload, cfg types.Config) (types.EvalResult, error) {
	responseBody, err := MakeRequest(payloadData, cfg)
	if err != nil {
		sugar.Errorw("request failed", "error", err)
		return types.EvalResult{}, err
	}
	var response types.EvalResult

	if err := json.Unmarshal(responseBody, &response); err != nil {
		sugar.Errorw("marshalling failed, retrying", "error", err)
		// return makeRequest(payloadData, cfg)
	}

	err = response.Validate()
	if err != nil {
		sugar.Warnw("validation failed", "error", err, "response", string(responseBody))
		// return makeRequest(payloadData, cfg)
	}
	return response, nil
}

func Evaluate(question, correctAnswer, userAnswer string) (types.EvalResult, error) {
	cfg, err := types.NewConfig()
	if err != nil {
		sugar.Errorw("config Creation failed", "error", err)
		os.Exit(1)
	}
	payloadData, err := createPayload(question, correctAnswer, userAnswer)
	if err != nil {
		sugar.Errorw("failed to create payload", "error", err)
	}
	return makeRequest(payloadData, cfg)
}
