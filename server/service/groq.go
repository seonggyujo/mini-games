package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	GroqAPIURL   = "https://api.groq.com/openai/v1/chat/completions"
	GroqModel    = "llama-3.1-70b-versatile"
	GroqTimeout  = 30 * time.Second
)

// GroqClient handles communication with Groq API
type GroqClient struct {
	apiKey     string
	httpClient *http.Client
}

// NewGroqClient creates a new Groq API client
func NewGroqClient() *GroqClient {
	apiKey := os.Getenv("GROQ_API_KEY")
	if apiKey == "" {
		fmt.Println("Warning: GROQ_API_KEY not set")
	}
	
	return &GroqClient{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: GroqTimeout,
		},
	}
}

// IsAvailable checks if the Groq API is configured
func (c *GroqClient) IsAvailable() bool {
	return c.apiKey != ""
}

// ChatRequest represents a request to Groq API
type ChatRequest struct {
	Model       string        `json:"model"`
	Messages    []ChatMessage `json:"messages"`
	Temperature float64       `json:"temperature"`
	MaxTokens   int           `json:"max_tokens"`
}

// ChatMessage represents a message in the chat
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatResponse represents a response from Groq API
type ChatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// Chat sends a chat request to Groq API
func (c *GroqClient) Chat(systemPrompt, userPrompt string) (string, error) {
	if !c.IsAvailable() {
		return "", fmt.Errorf("Groq API key not configured")
	}

	reqBody := ChatRequest{
		Model: GroqModel,
		Messages: []ChatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
		Temperature: 0.7,
		MaxTokens:   1024,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", GroqAPIURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	var chatResp ChatResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if chatResp.Error != nil {
		return "", fmt.Errorf("API error: %s", chatResp.Error.Message)
	}

	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("no response from API")
	}

	return strings.TrimSpace(chatResp.Choices[0].Message.Content), nil
}

// GenerateSentence generates a Korean sentence for translation practice
func (c *GroqClient) GenerateSentence(difficulty string) (string, error) {
	systemPrompt := `You are a Korean language expert. Generate a single Korean sentence for English translation practice.

Rules:
- Output ONLY the Korean sentence, nothing else
- No quotation marks, no explanation
- Make it natural and commonly used in daily life
- Do not include any English words`

	var userPrompt string
	switch difficulty {
	case "easy":
		userPrompt = `Generate a simple Korean sentence (5-8 words).
- Use basic vocabulary (TOPIK level 1-2)
- Simple present or past tense
- Topics: greetings, weather, food, family, daily activities`
	case "medium":
		userPrompt = `Generate a medium difficulty Korean sentence (8-12 words).
- Use intermediate vocabulary (TOPIK level 3-4)
- Can include compound sentences, honorifics
- Topics: work, travel, opinions, emotions, plans`
	case "hard":
		userPrompt = `Generate a challenging Korean sentence (12-18 words).
- Use advanced vocabulary (TOPIK level 5-6)
- Include idioms, proverbs, or complex grammar
- Topics: society, culture, abstract concepts, formal situations`
	default:
		userPrompt = `Generate a medium difficulty Korean sentence (8-12 words).`
	}

	sentence, err := c.Chat(systemPrompt, userPrompt)
	if err != nil {
		// Return fallback sentence if API fails
		return c.getFallbackSentence(difficulty), nil
	}

	// Clean up the sentence (remove quotes if any)
	sentence = strings.Trim(sentence, "\"'")
	return sentence, nil
}

// getFallbackSentence returns a default sentence when API fails
func (c *GroqClient) getFallbackSentence(difficulty string) string {
	switch difficulty {
	case "easy":
		return "오늘 날씨가 좋습니다."
	case "hard":
		return "급할수록 돌아가라는 말처럼 서두르지 않고 차근차근 준비하는 것이 중요합니다."
	default:
		return "요즘 바쁜 일상 속에서도 건강을 챙기는 것이 중요해요."
	}
}

// EvaluationResult represents the AI evaluation of translations
type EvaluationResult struct {
	Translation1 TranslationEval `json:"translation1"`
	Translation2 TranslationEval `json:"translation2"`
	ModelAnswer  string          `json:"modelAnswer"`
}

// TranslationEval represents evaluation scores for a single translation
type TranslationEval struct {
	Meaning     int    `json:"meaning"`
	Grammar     int    `json:"grammar"`
	Naturalness int    `json:"naturalness"`
	Feedback    string `json:"feedback"`
}

// EvaluateTranslations evaluates two translations of a Korean sentence
func (c *GroqClient) EvaluateTranslations(korean, trans1, trans2 string) (*EvaluationResult, error) {
	systemPrompt := `You are an expert English-Korean translator and language evaluator.
Evaluate two English translations of the given Korean sentence.
You must respond in valid JSON format only, no other text.

Scoring guidelines:
- meaning (0-40): Does it accurately convey the original meaning? Full marks for complete accuracy.
- grammar (0-30): Is the English grammar correct? Deduct for errors.
- naturalness (0-30): Does it sound natural to native speakers? Consider word choice and phrasing.

If a translation is empty or clearly not an attempt, give 0 for all scores.`

	userPrompt := fmt.Sprintf(`Korean sentence: "%s"

Translation 1: "%s"
Translation 2: "%s"

Evaluate each translation and provide:
- Brief feedback for each translation (in Korean, 1-2 sentences)
- A model answer (the best possible translation)

Respond ONLY in this exact JSON format:
{
  "translation1": { "meaning": N, "grammar": N, "naturalness": N, "feedback": "..." },
  "translation2": { "meaning": N, "grammar": N, "naturalness": N, "feedback": "..." },
  "modelAnswer": "..."
}`, korean, trans1, trans2)

	response, err := c.Chat(systemPrompt, userPrompt)
	if err != nil {
		// Return fallback evaluation if API fails
		return c.getFallbackEvaluation(trans1, trans2), nil
	}

	// Try to extract JSON from the response (in case there's extra text)
	jsonStart := strings.Index(response, "{")
	jsonEnd := strings.LastIndex(response, "}")
	if jsonStart >= 0 && jsonEnd > jsonStart {
		response = response[jsonStart : jsonEnd+1]
	}

	var result EvaluationResult
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		// Return fallback evaluation if parsing fails
		return c.getFallbackEvaluation(trans1, trans2), nil
	}

	// Validate and clamp scores
	result.Translation1 = c.validateScores(result.Translation1)
	result.Translation2 = c.validateScores(result.Translation2)

	return &result, nil
}

// validateScores ensures scores are within valid ranges
func (c *GroqClient) validateScores(eval TranslationEval) TranslationEval {
	clamp := func(val, min, max int) int {
		if val < min {
			return min
		}
		if val > max {
			return max
		}
		return val
	}

	eval.Meaning = clamp(eval.Meaning, 0, 40)
	eval.Grammar = clamp(eval.Grammar, 0, 30)
	eval.Naturalness = clamp(eval.Naturalness, 0, 30)
	return eval
}

// getFallbackEvaluation returns default evaluation when API fails
func (c *GroqClient) getFallbackEvaluation(trans1, trans2 string) *EvaluationResult {
	eval1 := TranslationEval{Meaning: 25, Grammar: 20, Naturalness: 15, Feedback: "평가를 수행할 수 없습니다."}
	eval2 := TranslationEval{Meaning: 25, Grammar: 20, Naturalness: 15, Feedback: "평가를 수행할 수 없습니다."}

	// Give 0 scores for empty translations
	if strings.TrimSpace(trans1) == "" {
		eval1 = TranslationEval{Meaning: 0, Grammar: 0, Naturalness: 0, Feedback: "번역이 제출되지 않았습니다."}
	}
	if strings.TrimSpace(trans2) == "" {
		eval2 = TranslationEval{Meaning: 0, Grammar: 0, Naturalness: 0, Feedback: "번역이 제출되지 않았습니다."}
	}

	return &EvaluationResult{
		Translation1: eval1,
		Translation2: eval2,
		ModelAnswer:  "(모범 답안을 생성할 수 없습니다)",
	}
}
