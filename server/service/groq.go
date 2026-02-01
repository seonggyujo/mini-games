package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	GroqAPIURL   = "https://api.groq.com/openai/v1/chat/completions"
	GroqModel    = "openai/gpt-oss-120b"
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
	return c.ChatWithTemp(systemPrompt, userPrompt, 0.7)
}

// ChatWithTemp sends a chat request to Groq API with custom temperature
func (c *GroqClient) ChatWithTemp(systemPrompt, userPrompt string, temperature float64) (string, error) {
	if !c.IsAvailable() {
		return "", fmt.Errorf("Groq API key not configured")
	}

	reqBody := ChatRequest{
		Model: GroqModel,
		Messages: []ChatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
		Temperature: temperature,
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
	// 랜덤 시드 생성 (매번 다른 문장 유도)
	randomSeed := rand.Intn(90000) + 10000

	// 다양한 랜덤 요소들
	topics := []string{
		"여행", "음식", "취미", "가족", "친구", "직장", "학교", "건강",
		"운동", "쇼핑", "날씨", "계절", "동물", "음악", "영화",
		"책", "게임", "기술", "자연", "도시", "바다", "산",
		"커피", "요리", "청소", "빨래", "전화", "선물", "약속", "휴가",
	}
	situations := []string{
		"아침에", "점심시간에", "저녁에", "주말에", "휴일에",
		"비 오는 날", "눈 오는 날", "여름에", "겨울에", "봄에", "가을에",
		"출근길에", "퇴근 후", "집에서", "카페에서", "공원에서",
		"회사에서", "학교에서", "병원에서", "마트에서", "식당에서",
	}
	moods := []string{
		"기쁜", "슬픈", "피곤한", "신나는", "걱정되는",
		"편안한", "바쁜", "심심한", "배고픈", "졸린",
		"행복한", "설레는", "긴장한", "궁금한", "만족스러운",
	}
	actions := []string{
		"계획을 세우는", "추억하는", "결정하는", "도움을 주는", "배우는",
		"만나는", "기다리는", "준비하는", "즐기는", "고민하는",
		"찾는", "사는", "먹는", "마시는", "보는", "듣는", "말하는",
	}

	// 랜덤 요소 선택
	randomTopic := topics[rand.Intn(len(topics))]
	randomSituation := situations[rand.Intn(len(situations))]
	randomMood := moods[rand.Intn(len(moods))]
	randomAction := actions[rand.Intn(len(actions))]

	systemPrompt := `You are a Korean language expert. Generate a single Korean sentence for English translation practice.

Rules:
- Output ONLY the Korean sentence, nothing else
- No quotation marks, no explanation
- Make it natural and commonly used in daily life
- Do not include any English words
- MUST incorporate the given random elements (topic, situation, mood, action)
- Create a completely unique and unexpected sentence every time`

	var userPrompt string
	switch difficulty {
	case "easy":
		userPrompt = fmt.Sprintf(`[Seed: %d] [Topic: %s] [Situation: %s]

Generate a VERY SIMPLE Korean sentence (3-5 words only).
- Elementary school level vocabulary
- Simple present or past tense only
- Examples of this level: "나는 밥을 먹어요", "오늘 학교에 갔어요", "친구가 좋아요"
- Incorporate the topic "%s" naturally
- Keep it SHORT and EASY`, randomSeed, randomTopic, randomSituation, randomTopic)
	case "medium":
		userPrompt = fmt.Sprintf(`[Seed: %d] [Topic: %s] [Situation: %s] [Mood: %s]

Generate a simple Korean sentence (5-8 words).
- Use basic vocabulary (TOPIK level 1-2)
- Simple sentence structure
- Incorporate: topic "%s", situation "%s"
- Make it natural everyday conversation`, randomSeed, randomTopic, randomSituation, randomMood, randomTopic, randomSituation)
	case "hard":
		userPrompt = fmt.Sprintf(`[Seed: %d] [Topic: %s] [Situation: %s] [Mood: %s] [Action: %s]

Generate a medium difficulty Korean sentence (8-12 words).
- Use intermediate vocabulary (TOPIK level 3-4)
- Can include compound sentences or honorifics
- Creatively combine: topic "%s", situation "%s", mood "%s", action "%s"
- Make it specific and vivid`, randomSeed, randomTopic, randomSituation, randomMood, randomAction, randomTopic, randomSituation, randomMood, randomAction)
	default:
		userPrompt = fmt.Sprintf(`[Seed: %d] [Topic: %s] [Situation: %s]

Generate a simple Korean sentence (5-8 words) about "%s".
- Be creative and unique`, randomSeed, randomTopic, randomSituation, randomTopic)
	}

	// Temperature 1.0으로 최대 다양성 유도
	sentence, err := c.ChatWithTemp(systemPrompt, userPrompt, 1.0)
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
