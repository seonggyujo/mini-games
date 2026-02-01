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

	// 다양한 랜덤 주제들
	topics := []string{
		"여행", "음식", "취미", "가족", "친구", "직장", "학교", "건강",
		"운동", "쇼핑", "날씨", "계절", "동물", "음악", "영화",
		"책", "게임", "기술", "자연", "도시", "바다", "산",
		"커피", "요리", "청소", "빨래", "전화", "선물", "약속", "휴가",
	}

	// 랜덤 주제 선택
	randomTopic := topics[rand.Intn(len(topics))]

	systemPrompt := `당신은 한국어 원어민입니다. 영어 번역 연습을 위한 한국어 문장을 하나 생성하세요.

규칙:
- 한국어 문장만 출력하세요 (따옴표, 설명, 번호 없이)
- 실제 한국인이 일상에서 자주 사용하는 자연스러운 표현을 사용하세요
- 번역체나 어색한 표현을 절대 사용하지 마세요
- 영어 단어를 포함하지 마세요
- 주어진 주제를 자연스럽게 반영하되, 억지로 넣지 마세요`

	var userPrompt string
	switch difficulty {
	case "easy":
		userPrompt = fmt.Sprintf(`[시드: %d] 주제: %s

아주 쉬운 한국어 문장 (3-5 단어)을 만드세요.
- 초등학생도 이해할 수 있는 쉬운 단어
- 현재형 또는 과거형
- 좋은 예시: "오늘 날씨가 좋아요", "친구랑 밥 먹었어요", "강아지가 귀여워요", "내일 학교 가요"
- 나쁜 예시: "~하는 것이 좋다", "~할 수 있다면" (너무 복잡함)`, randomSeed, randomTopic)
	case "medium":
		userPrompt = fmt.Sprintf(`[시드: %d] 주제: %s

적당한 난이도의 한국어 문장 (5-8 단어)을 만드세요.
- 중학생 수준의 어휘
- 일상 대화에서 흔히 쓰는 표현
- 좋은 예시: "주말에 친구들이랑 영화 보러 갈 거예요", "요즘 너무 바빠서 운동을 못 해요"
- 나쁜 예시: "~에 대해 생각해 보면", "~라고 할 수 있다" (번역체)`, randomSeed, randomTopic)
	case "hard":
		userPrompt = fmt.Sprintf(`[시드: %d] 주제: %s

조금 어려운 한국어 문장 (8-12 단어)을 만드세요.
- 고등학생/성인 수준의 어휘
- 복문이나 연결어미 사용 가능
- 존댓말/반말 자유롭게
- 좋은 예시: "어제 밤새 공부했더니 오늘 하루 종일 졸려서 집중이 안 돼요"
- 나쁜 예시: "~하는 것이 중요하다고 생각된다" (논문체, 번역체)`, randomSeed, randomTopic)
	default:
		userPrompt = fmt.Sprintf(`[시드: %d] 주제: %s

적당한 난이도의 자연스러운 한국어 문장 (5-8 단어)을 만드세요.`, randomSeed, randomTopic)
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
