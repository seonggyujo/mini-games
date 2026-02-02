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

// SentenceWithTense represents a sentence with its tense
type SentenceWithTense struct {
	Sentence string `json:"sentence"`
	Tense    string `json:"tense"` // "현재형", "과거형", "미래형"
}

// GenerateSentence generates a Korean sentence for translation practice
func (c *GroqClient) GenerateSentence(difficulty string) (string, error) {
	result, err := c.GenerateSentenceWithTense(difficulty)
	if err != nil {
		return "", err
	}
	return result.Sentence, nil
}

// GenerateSentenceWithTense generates a Korean sentence with tense information
func (c *GroqClient) GenerateSentenceWithTense(difficulty string) (*SentenceWithTense, error) {
	// 랜덤 시드 생성 (매번 다른 문장 유도)
	randomSeed := rand.Intn(90000) + 10000

	// 다양한 랜덤 주제들 (60개 이상)
	topics := []string{
		// 일상생활
		"아침 루틴", "저녁 일과", "주말 계획", "평일 일상", "출퇴근",
		// 음식/요리
		"아침식사", "점심", "저녁식사", "간식", "디저트", "요리", "배달음식", "외식", "카페",
		// 관계
		"가족", "친구", "직장동료", "이웃", "선생님", "선후배",
		// 장소
		"학교", "회사", "마트", "병원", "공원", "도서관", "헬스장", "영화관", "백화점",
		// 활동
		"운동", "독서", "게임", "쇼핑", "청소", "빨래", "산책", "등산", "수영", "자전거",
		// 감정/상태
		"피곤함", "기쁨", "걱정", "설렘", "그리움", "후회", "감사", "짜증",
		// 계절/날씨
		"봄", "여름", "가을", "겨울", "비오는 날", "눈오는 날", "더운 날", "추운 날",
		// 이벤트
		"생일", "졸업", "입학", "여행", "휴가", "명절", "기념일", "면접", "시험",
		// 기타
		"건강", "다이어트", "취미", "반려동물", "식물", "음악", "영화", "드라마", "뉴스",
	}

	// 다양한 상황 목록
	situations := []string{
		"친구에게 말할 때", "부모님께 말씀드릴 때", "혼잣말할 때", "일기를 쓸 때",
		"SNS에 올릴 때", "선생님께 여쭤볼 때", "동생에게 말할 때", "직장 동료와 대화할 때",
	}

	// 다양한 화자 유형
	speakers := []string{
		"학생", "직장인", "주부", "대학생", "고등학생", "아이", "선생님", "회사원",
	}

	// 랜덤 요소 선택
	randomTopic := topics[rand.Intn(len(topics))]
	randomSituation := situations[rand.Intn(len(situations))]
	randomSpeaker := speakers[rand.Intn(len(speakers))]

	systemPrompt := `당신은 한국어 원어민입니다. 영어 번역 연습을 위한 한국어 문장을 하나 생성하세요.

규칙:
- 반드시 JSON 형식으로만 응답하세요: {"sentence": "문장", "tense": "시제"}
- tense는 반드시 "현재형", "과거형", "미래형" 중 하나여야 합니다
- **주어를 반드시 명시하세요** (예: "나는", "우리는", "그녀는", "철수는", "엄마가", "친구들이" 등)
- 주어가 생략된 문장은 절대 만들지 마세요
- 실제 한국인이 일상에서 자주 사용하는 자연스러운 표현을 사용하세요
- 번역체나 어색한 표현을 절대 사용하지 마세요
- 영어 단어를 포함하지 마세요
- 매번 완전히 새롭고 다른 문장을 만드세요`

	var userPrompt string
	switch difficulty {
	case "easy":
		userPrompt = fmt.Sprintf(`[시드: %d] 주제: %s, 화자: %s

아주 쉬운 한국어 문장 (4-6 단어)을 만드세요.
- 주어를 반드시 포함하세요 (나는, 우리는, 엄마가, 친구가 등)
- 초등학생도 이해할 수 있는 쉬운 단어만 사용
- 단순한 문장 구조 (주어 + 목적어/부사 + 동사)
- 현재형, 과거형, 미래형 중 하나 선택
- 좋은 예시: {"sentence": "나는 어제 학교에 갔어요", "tense": "과거형"}
- 좋은 예시: {"sentence": "엄마가 맛있는 밥을 해요", "tense": "현재형"}
- 좋은 예시: {"sentence": "우리는 내일 공원에 갈 거예요", "tense": "미래형"}
- 나쁜 예시: {"sentence": "오늘 날씨가 좋아요", "tense": "현재형"} (주어 불명확)`, randomSeed, randomTopic, randomSpeaker)
	case "medium":
		userPrompt = fmt.Sprintf(`[시드: %d] 주제: %s, 상황: %s, 화자: %s

중급 난이도의 한국어 문장 (8-12 단어)을 만드세요.
- 주어를 반드시 포함하세요
- 고등학생/성인 수준의 어휘 사용
- 복문 또는 연결어미 사용 (예: -아서, -으면, -지만, -는데, -고)
- 일상에서 실제로 쓰는 자연스러운 표현
- 현재형, 과거형, 미래형 중 하나 선택 (메인 동사 기준)
- 좋은 예시: {"sentence": "나는 어제 비가 와서 우산을 들고 나갔는데 결국 안 썼어요", "tense": "과거형"}
- 좋은 예시: {"sentence": "우리 팀이 이번 프로젝트를 성공하면 다 같이 회식하기로 했어요", "tense": "과거형"}
- 좋은 예시: {"sentence": "동생이 요즘 시험 준비하느라 너무 바빠서 밥도 제대로 못 먹어요", "tense": "현재형"}`, randomSeed, randomTopic, randomSituation, randomSpeaker)
	case "hard":
		userPrompt = fmt.Sprintf(`[시드: %d] 주제: %s, 상황: %s, 화자: %s

고급 난이도의 한국어 문장 (10-15 단어)을 만드세요.
- 주어를 반드시 포함하세요
- 성인/비즈니스 수준의 어휘 사용
- 다음 요소 중 하나 이상 포함:
  * 관용 표현 (예: 발이 넓다, 손이 크다, 눈이 높다, 입이 무겁다)
  * 속담이나 격언 활용 (예: ~라는 말처럼, ~다더니)
  * 복잡한 연결 구조 (-았/었더니, -고 나서, -ㄴ/는 바람에, -기는커녕)
  * 높임말/겸양어 사용
- 현재형, 과거형, 미래형 중 하나 선택 (메인 동사 기준)
- 좋은 예시: {"sentence": "부장님께서 급한 일이 생기셔서 오늘 회의를 내일로 미루자고 하셨어요", "tense": "과거형"}
- 좋은 예시: {"sentence": "발 없는 말이 천 리 간다더니 제가 한 말이 벌써 다른 부서까지 퍼졌대요", "tense": "과거형"}
- 좋은 예시: {"sentence": "그 친구는 입이 무거워서 비밀을 말해도 절대 다른 사람에게 말하지 않아요", "tense": "현재형"}`, randomSeed, randomTopic, randomSituation, randomSpeaker)
	default:
		userPrompt = fmt.Sprintf(`[시드: %d] 주제: %s, 화자: %s

적당한 난이도의 자연스러운 한국어 문장 (6-10 단어)을 만드세요.
- 주어를 반드시 포함하세요
- 일상에서 자주 쓰는 표현 사용
JSON 형식으로 응답: {"sentence": "문장", "tense": "시제"}`, randomSeed, randomTopic, randomSpeaker)
	}

	// Temperature 1.0으로 최대 다양성 유도
	response, err := c.ChatWithTemp(systemPrompt, userPrompt, 1.0)
	if err != nil {
		// Return fallback sentence if API fails
		return c.getFallbackSentenceWithTense(difficulty), nil
	}

	// Try to extract JSON from the response
	jsonStart := strings.Index(response, "{")
	jsonEnd := strings.LastIndex(response, "}")
	if jsonStart >= 0 && jsonEnd > jsonStart {
		response = response[jsonStart : jsonEnd+1]
	}

	var result SentenceWithTense
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		// If parsing fails, try to use response as plain text
		sentence := strings.Trim(response, "\"'")
		return &SentenceWithTense{
			Sentence: sentence,
			Tense:    "현재형", // Default tense
		}, nil
	}

	// Validate tense
	if result.Tense != "현재형" && result.Tense != "과거형" && result.Tense != "미래형" {
		result.Tense = "현재형" // Default to present tense
	}

	return &result, nil
}

// getFallbackSentence returns a default sentence when API fails
func (c *GroqClient) getFallbackSentence(difficulty string) string {
	result := c.getFallbackSentenceWithTense(difficulty)
	return result.Sentence
}

// getFallbackSentenceWithTense returns a default sentence with tense when API fails
func (c *GroqClient) getFallbackSentenceWithTense(difficulty string) *SentenceWithTense {
	switch difficulty {
	case "easy":
		return &SentenceWithTense{Sentence: "오늘 날씨가 좋습니다.", Tense: "현재형"}
	case "hard":
		return &SentenceWithTense{Sentence: "급할수록 돌아가라는 말처럼 서두르지 않고 차근차근 준비하는 것이 중요합니다.", Tense: "현재형"}
	default:
		return &SentenceWithTense{Sentence: "요즘 바쁜 일상 속에서도 건강을 챙기는 것이 중요해요.", Tense: "현재형"}
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
	systemPrompt := `You are an expert English-Korean translator and native English speaker.
Evaluate two English translations of the given Korean sentence.
You must respond in valid JSON format only, no other text.

Scoring guidelines:
- meaning (0-40): Does it accurately convey the original meaning? Full marks for complete accuracy.
- grammar (0-30): Is the English grammar correct? Deduct for errors.
- naturalness (0-30): Does it sound natural to native speakers? Consider word choice and phrasing.

IMPORTANT for model answer:
- The model answer MUST be what a native English speaker would actually say in real conversation
- Use common, everyday expressions - NOT overly formal or textbook-like translations
- Prefer natural contractions (I'm, don't, can't, wasn't) when appropriate for casual speech
- Use natural word order and phrasing that sounds conversational
- Avoid literal word-for-word translations that sound awkward in English

If a translation is empty or clearly not an attempt, give 0 for all scores.`

	userPrompt := fmt.Sprintf(`Korean sentence: "%s"

Translation 1: "%s"
Translation 2: "%s"

Evaluate each translation and provide:
- Brief feedback for each translation (in Korean, 1-2 sentences, be specific about what's good or needs improvement)
- A model answer that sounds completely natural - exactly how a native English speaker would express this in everyday conversation (NOT a textbook translation)

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
