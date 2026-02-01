# Translation Battle - 구현 계획서

> AI가 생성한 한국어 문장을 영어로 번역하는 1:1 대결 게임

---

## 📋 구현 진행 상황

### 현재 상태: ✅ 구현 완료 (테스트 필요)

| 단계 | 상태 | 완료일 |
|------|------|--------|
| Phase 1: 서버 | ✅ 완료 | 2026-02-01 |
| Phase 2: 클라이언트 로비 | ✅ 완료 | 2026-02-01 |
| Phase 3: 클라이언트 게임 | ✅ 완료 | 2026-02-01 |
| Phase 4: 통합 | ✅ 완료 | 2026-02-01 |
| Phase 5: 테스트 | ⏳ 대기 | - |

### 생성된 파일 목록

#### 서버 (Go)
| 파일 | 설명 | 코드 라인 |
|------|------|----------|
| `server/service/groq.go` | Groq API 클라이언트 (문장 생성, 번역 평가) | ~260 |
| `server/service/translationbattle.go` | 게임 로직 (방 관리, 라운드 진행, 평가) | ~470 |
| `server/model/game.go` | Translation Battle 모델 추가 | +50 |
| `server/model/room.go` | TranslationRoom + 15개 메시지 타입 | +120 |
| `server/handler/websocket.go` | HandleTranslationWS 핸들러 추가 | +150 |
| `server/main.go` | `/ws/translation` 라우트 추가 | +2 |

#### 클라이언트 (React)
| 파일 | 설명 | 코드 라인 |
|------|------|----------|
| `client/src/hooks/useTranslationWS.js` | WebSocket 훅 (번역 게임용) | ~110 |
| `client/src/components/games/TranslationBattle/TranslationLobby.jsx` | 로비 컴포넌트 | ~330 |
| `client/src/components/games/TranslationBattle/TranslationLobby.css` | 로비 스타일 | ~340 |
| `client/src/components/games/TranslationBattle/TranslationBattle.jsx` | 게임 컴포넌트 | ~230 |
| `client/src/components/games/TranslationBattle/TranslationBattle.css` | 게임 스타일 | ~340 |
| `client/src/pages/HomePage.jsx` | 게임 목록에 추가 | +7 |
| `client/src/pages/GamePage.jsx` | 컴포넌트 매핑 추가 | +2 |

### 주요 구현 기능

1. **방 생성/참가 시스템**
   - 6자리 방 코드 생성
   - 닉네임 입력 및 유효성 검사
   - WebSocket 연결 관리

2. **게임 진행**
   - 3판 2선승제
   - 60초 타이머
   - 난이도 선택 (Easy/Medium/Hard)
   - 실시간 상대 제출 상태 표시

3. **AI 연동 (Groq API)**
   - 난이도별 한국어 문장 생성
   - 두 번역 동시 평가
   - 점수 breakdown (의미 40 + 문법 30 + 자연스러움 30 = 100점)
   - 피드백 및 모범 답안 제공

4. **결과 표시**
   - 라운드별 상세 점수 비교
   - 양쪽 번역 및 피드백 표시
   - 모범 답안 제공
   - 재대결 기능

5. **에러 처리**
   - API 실패 시 fallback 문장/점수 사용
   - 연결 끊김 시 상대방 알림
   - 방 타임아웃 자동 정리

### 테스트 방법

```powershell
# 1. Groq API 키 설정
$env:GROQ_API_KEY="gsk_your_api_key_here"

# 2. 서버 실행
cd server
go run .

# 3. 클라이언트 실행 (새 터미널)
cd client
npm start

# 4. 테스트
# - 브라우저 2개 열기
# - 한쪽에서 방 생성, 다른 쪽에서 코드로 참가
# - 난이도 선택 후 게임 시작
# - 번역 제출 및 결과 확인
```

### 남은 작업

- [ ] 실제 게임 플로우 테스트
- [ ] Groq API 연동 테스트
- [ ] 에러 케이스 테스트 (연결 끊김, API 실패 등)
- [ ] 모바일 반응형 테스트

---

## 1. 게임 개요

### 1.1 기본 정보
| 항목 | 내용 |
|------|------|
| 게임 ID | `translation-battle` |
| 아이콘 | 🔤 |
| 모드 | 멀티플레이어 전용 (1v1) |
| 라운드 | 3판 2선승제 |
| 제한시간 | 60초/라운드 |
| 난이도 | Easy / Medium / Hard |
| AI | Groq API (llama-3.1-70b-versatile) |

### 1.2 점수 체계
| 항목 | 배점 | 설명 |
|------|------|------|
| 의미 전달 (Meaning) | 40점 | 원문의 의미를 정확히 전달했는가 |
| 문법 (Grammar) | 30점 | 영어 문법이 올바른가 |
| 자연스러움 (Naturalness) | 30점 | 원어민이 사용하는 자연스러운 표현인가 |
| **총점** | **100점** | |

### 1.3 핵심 요구사항
1. **양 플레이어 동일 결과 표시** - AI 평가 결과가 양쪽에 정확히 동일하게 전송
2. **실시간 제출 상태** - 상대 제출 시 "응답 완료" 표시 (답은 비공개)
3. **교육적 피드백** - 점수 breakdown, 각 플레이어 피드백, 모범 답안 제공

---

## 2. 게임 흐름

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              GAME FLOW                                       │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  [1. 로비]                                                                   │
│      │                                                                       │
│      ├─ 플레이어 A: 방 생성 (닉네임 입력)                                    │
│      │      └─ 방 코드 생성 (6자리)                                          │
│      │                                                                       │
│      └─ 플레이어 B: 방 코드로 참가                                           │
│                                                                              │
│  [2. 대기실]                                                                 │
│      │                                                                       │
│      └─ 방장(A)이 난이도 선택 후 "게임 시작"                                 │
│                                                                              │
│  [3. 라운드 시작] ◄─────────────────────────────────────────────────────┐   │
│      │                                                                   │   │
│      ├─ 서버: AI로 한국어 문장 생성                                      │   │
│      ├─ 양쪽에 동일한 문장 전송                                          │   │
│      └─ 60초 카운트다운 시작                                             │   │
│                                                                          │   │
│  [4. 번역 단계]                                                          │   │
│      │                                                                   │   │
│      ├─ 각자 영어 번역 작성                                              │   │
│      ├─ 제출 시 상대에게 "응답 완료" 표시                                │   │
│      └─ 둘 다 제출 OR 시간 초과 → 평가 단계                              │   │
│                                                                          │   │
│  [5. 평가 단계]                                                          │   │
│      │                                                                   │   │
│      ├─ "AI가 평가 중..." 표시                                           │   │
│      ├─ 서버: AI로 양쪽 번역 평가 (1회 호출)                             │   │
│      └─ 동일한 결과를 양쪽에 전송                                        │   │
│                                                                          │   │
│  [6. 결과 표시]                                                          │   │
│      │                                                                   │   │
│      ├─ 점수 breakdown (의미/문법/자연스러움)                            │   │
│      ├─ 각 플레이어 피드백                                               │   │
│      ├─ 모범 답안                                                        │   │
│      └─ 라운드 승자 표시                                                 │   │
│                                                                          │   │
│  [7. 분기]                                                               │   │
│      │                                                                   │   │
│      ├─ 2승 달성 → [8. 최종 결과]                                        │   │
│      └─ 아직 → 5초 후 다음 라운드 ──────────────────────────────────────┘   │
│                                                                              │
│  [8. 최종 결과]                                                              │
│      │                                                                       │
│      ├─ 최종 승패 표시                                                       │
│      ├─ 총점 비교                                                            │
│      └─ 재대결 / 나가기 선택                                                 │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## 3. 파일 구조

### 3.1 새로 생성할 파일 (6개)

```
server/
├── service/
│   ├── groq.go                 # Groq API 클라이언트
│   └── translationbattle.go    # 게임 로직 (방 관리, 라운드 진행)

client/src/components/games/TranslationBattle/
├── TranslationLobby.jsx        # 로비 화면 (방 생성/참가/대기/결과)
├── TranslationLobby.css        # 로비 스타일
├── TranslationBattle.jsx       # 게임 화면 (번역/평가/결과)
└── TranslationBattle.css       # 게임 스타일
```

### 3.2 수정할 파일 (6개)

```
server/
├── model/game.go               # Translation Battle 모델 추가
├── model/room.go               # WebSocket 메시지 타입 추가
├── handler/websocket.go        # Translation WebSocket 핸들러 추가
└── main.go                     # /ws/translation 라우트 추가

client/src/
├── pages/HomePage.jsx          # 게임 목록에 추가
└── pages/GamePage.jsx          # 컴포넌트 매핑 추가
```

---

## 4. 데이터 모델 (Go)

### 4.1 server/model/game.go에 추가

```go
// ========== Translation Battle Models ==========

// TranslationDifficulty 난이도
type TranslationDifficulty string

const (
    DifficultyEasy   TranslationDifficulty = "easy"
    DifficultyMedium TranslationDifficulty = "medium"
    DifficultyHard   TranslationDifficulty = "hard"
)

// TranslationRound 라운드 정보
type TranslationRound struct {
    Number        int                    `json:"number"`        // 1, 2, 3
    Sentence      string                 `json:"sentence"`      // 한국어 문장
    Translations  [2]string              `json:"translations"`  // 플레이어 번역 [player0, player1]
    Submitted     [2]bool                `json:"submitted"`     // 제출 여부
    SubmitTimes   [2]time.Time           `json:"submitTimes"`   // 제출 시간
    Scores        [2]TranslationScore    `json:"scores"`        // 평가 점수
    ModelAnswer   string                 `json:"modelAnswer"`   // 모범 답안
    Winner        int                    `json:"winner"`        // -1: draw, 0: player0, 1: player1
    StartTime     time.Time              `json:"startTime"`
}

// TranslationScore 번역 평가 점수
type TranslationScore struct {
    Meaning      int    `json:"meaning"`      // 0-40
    Grammar      int    `json:"grammar"`      // 0-30
    Naturalness  int    `json:"naturalness"`  // 0-30
    Total        int    `json:"total"`        // 0-100
    Feedback     string `json:"feedback"`     // AI 피드백
}

// TranslationGameState 게임 상태
type TranslationGameState string

const (
    TStateWaiting    TranslationGameState = "waiting"     // 대기 중
    TStateReady      TranslationGameState = "ready"       // 게임 시작 대기
    TStateCountdown  TranslationGameState = "countdown"   // 카운트다운
    TStatePlaying    TranslationGameState = "playing"     // 번역 중
    TStateEvaluating TranslationGameState = "evaluating"  // AI 평가 중
    TStateResult     TranslationGameState = "result"      // 결과 표시
    TStateFinished   TranslationGameState = "finished"    // 게임 종료
)
```

### 4.2 server/model/room.go에 추가

```go
// ========== Translation Battle Room ==========

// TranslationRoom 번역 배틀 방
type TranslationRoom struct {
    Code       string                `json:"code"`
    Players    [2]*Player            `json:"players"`
    State      TranslationGameState  `json:"state"`
    Difficulty TranslationDifficulty `json:"difficulty"`
    
    Rounds     []TranslationRound    `json:"rounds"`
    CurrentRnd int                   `json:"currentRound"`  // 0, 1, 2
    Wins       [2]int                `json:"wins"`          // 각 플레이어 승리 횟수
    
    CreatedAt  time.Time
    Mu         sync.RWMutex
    StopGame   chan struct{}
}
```

---

## 5. WebSocket 메시지 프로토콜

### 5.1 Client → Server

| Type | 필드 | 설명 |
|------|------|------|
| `t_create` | nickname | 방 생성 |
| `t_join` | nickname, roomCode | 방 참가 |
| `t_start_game` | difficulty | 게임 시작 (방장만) |
| `t_submit` | translation | 번역 제출 |
| `t_rematch` | - | 재대결 요청 |
| `t_leave` | - | 방 나가기 |

### 5.2 Server → Client

| Type | 필드 | 설명 |
|------|------|------|
| `t_room_created` | roomCode | 방 생성 완료 |
| `t_opponent_joined` | nickname | 상대 입장 |
| `t_game_start` | difficulty | 게임 시작 |
| `t_countdown` | count | 카운트다운 (3,2,1) |
| `t_round_start` | round, sentence | 라운드 시작 |
| `t_time_update` | timeLeft | 남은 시간 |
| `t_opponent_submitted` | - | 상대 제출 완료 |
| `t_evaluating` | - | 평가 중 |
| `t_round_result` | (상세) | 라운드 결과 |
| `t_game_over` | (상세) | 게임 종료 |
| `t_opponent_ready` | - | 상대 재대결 준비 |
| `t_rematch_start` | - | 재대결 시작 |
| `t_opponent_left` | - | 상대 퇴장 |
| `t_error` | message | 에러 |

### 5.3 t_round_result 상세

```json
{
  "type": "t_round_result",
  "round": 1,
  "sentence": "오늘 날씨가 좋습니다",
  "myTranslation": "The weather is nice today.",
  "opponentTranslation": "Today's weather is good.",
  "myScore": {
    "meaning": 38,
    "grammar": 28,
    "naturalness": 27,
    "total": 93,
    "feedback": "자연스러운 표현입니다."
  },
  "opponentScore": {
    "meaning": 35,
    "grammar": 25,
    "naturalness": 22,
    "total": 82,
    "feedback": "Today's weather보다 The weather today가 더 자연스럽습니다."
  },
  "modelAnswer": "The weather is nice today.",
  "roundWinner": "me",
  "totalWins": [1, 0],
  "isGameOver": false
}
```

---

## 6. Groq API 통합

### 6.1 환경 변수

```bash
GROQ_API_KEY=gsk_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
```

### 6.2 API 제한 (Free Tier)

| 항목 | 제한 |
|------|------|
| Requests/minute | 30 |
| Requests/day | 14,400 |
| 게임당 API 호출 | 최대 6회 (문장 생성 3회 + 평가 3회) |
| 동시 게임 가능 | 약 5개/분 |

### 6.3 server/service/groq.go 구조

```go
package service

type GroqClient struct {
    apiKey     string
    httpClient *http.Client
}

func NewGroqClient() *GroqClient

// Chat - 기본 API 호출
func (c *GroqClient) Chat(systemPrompt, userPrompt string) (string, error)

// GenerateSentence - 난이도에 맞는 한국어 문장 생성
func (c *GroqClient) GenerateSentence(difficulty string) (string, error)

// EvaluateTranslations - 두 번역을 동시에 평가
func (c *GroqClient) EvaluateTranslations(korean, trans1, trans2 string) (*EvaluationResult, error)

type EvaluationResult struct {
    Translation1 struct {
        Meaning     int    `json:"meaning"`
        Grammar     int    `json:"grammar"`
        Naturalness int    `json:"naturalness"`
        Feedback    string `json:"feedback"`
    } `json:"translation1"`
    Translation2 struct {
        Meaning     int    `json:"meaning"`
        Grammar     int    `json:"grammar"`
        Naturalness int    `json:"naturalness"`
        Feedback    string `json:"feedback"`
    } `json:"translation2"`
    ModelAnswer string `json:"modelAnswer"`
}
```

### 6.4 AI 프롬프트

#### 문장 생성 프롬프트

```
System: You are a Korean language expert. Generate a single Korean sentence for English translation practice.

Rules:
- Output ONLY the Korean sentence, nothing else
- No quotation marks, no explanation
- Make it natural and commonly used in daily life

User (Easy):
Generate a simple Korean sentence (5-8 words).
- Use basic vocabulary (TOPIK level 1-2)
- Simple present or past tense

User (Medium):
Generate a medium difficulty Korean sentence (8-12 words).
- Use intermediate vocabulary (TOPIK level 3-4)
- Can include compound sentences, honorifics

User (Hard):
Generate a challenging Korean sentence (12-18 words).
- Use advanced vocabulary (TOPIK level 5-6)
- Include idioms, proverbs, or complex grammar
```

#### 평가 프롬프트

```
System: You are an expert English-Korean translator and language evaluator.
Evaluate two English translations of the given Korean sentence.
You must respond in valid JSON format only, no other text.

User:
Korean sentence: "한국어 문장"

Translation 1: "번역 1"
Translation 2: "번역 2"

Evaluate each translation on:
1. meaning (0-40): Does it accurately convey the original meaning?
2. grammar (0-30): Is the English grammar correct?
3. naturalness (0-30): Does it sound natural to native speakers?

Also provide:
- Brief feedback for each translation (in Korean, 1-2 sentences)
- A model answer (the best possible translation)

Respond in JSON format:
{
  "translation1": { "meaning": N, "grammar": N, "naturalness": N, "feedback": "..." },
  "translation2": { "meaning": N, "grammar": N, "naturalness": N, "feedback": "..." },
  "modelAnswer": "..."
}
```

---

## 7. 게임 서비스

### 7.1 server/service/translationbattle.go 구조

```go
package service

const (
    TRoundDuration = 60 * time.Second
    TRoomTimeout   = 10 * time.Minute
)

type TranslationRoomManager struct {
    rooms      map[string]*model.TranslationRoom
    mu         sync.RWMutex
    groqClient *GroqClient
}

func NewTranslationRoomManager() *TranslationRoomManager

// 방 관리
func (rm *TranslationRoomManager) CreateRoom(player *model.Player) (*model.TranslationRoom, string)
func (rm *TranslationRoomManager) JoinRoom(code string, player *model.Player) (*model.TranslationRoom, error)
func (rm *TranslationRoomManager) RemovePlayer(code string, playerIndex int)

// 게임 진행
func (rm *TranslationRoomManager) StartGame(room *model.TranslationRoom, difficulty model.TranslationDifficulty)
func (rm *TranslationRoomManager) startRound(room *model.TranslationRoom)
func (rm *TranslationRoomManager) roundTimer(room *model.TranslationRoom)
func (rm *TranslationRoomManager) HandleSubmit(room *model.TranslationRoom, playerIndex int, translation string)
func (rm *TranslationRoomManager) evaluateRound(room *model.TranslationRoom)
func (rm *TranslationRoomManager) sendRoundResult(room *model.TranslationRoom, round *model.TranslationRound, isGameOver bool)
func (rm *TranslationRoomManager) endGame(room *model.TranslationRoom)
func (rm *TranslationRoomManager) HandleRematch(room *model.TranslationRoom, playerIndex int)

// 정리
func (rm *TranslationRoomManager) cleanupRoutine()
```

---

## 8. WebSocket 핸들러

### 8.1 server/handler/websocket.go에 추가

```go
var translationRoomManager *service.TranslationRoomManager

func InitWebSocket() {
    roomManager = service.NewRoomManager()
    translationRoomManager = service.NewTranslationRoomManager()  // 추가
}

func HandleTranslationWS(w http.ResponseWriter, r *http.Request) {
    // 기존 HandleBattleWS와 유사한 구조
    // 메시지 타입: t_create, t_join, t_start_game, t_submit, t_rematch, t_leave
}
```

### 8.2 server/main.go에 추가

```go
// mainMux 부분에 추가
wsTranslationHandler := middleware.RateLimit(http.HandlerFunc(handler.HandleTranslationWS))
mainMux.Handle("/ws/translation", wsTranslationHandler)
```

---

## 9. 클라이언트 컴포넌트

### 9.1 TranslationLobby.jsx 화면 구성

```
┌─────────────────────────────────────────┐
│           Translation Battle            │
│      AI와 함께하는 영어 번역 대결       │
├─────────────────────────────────────────┤
│                                         │
│  닉네임: [______________]               │
│                                         │
│       [ 방 만들기 ]                     │
│                                         │
│         ─── 또는 ───                    │
│                                         │
│  방 코드: [______]                      │
│                                         │
│       [ 참가하기 ]                      │
│                                         │
│       [ 돌아가기 ]                      │
│                                         │
└─────────────────────────────────────────┘

┌─────────────────────────────────────────┐
│            대기 중...                   │
├─────────────────────────────────────────┤
│                                         │
│           방 코드                       │
│          [ ABCDEF ]                     │
│                                         │
│    친구에게 이 코드를 공유하세요!       │
│                                         │
│           ⏳ 대기 중                    │
│                                         │
│          [ 취소 ]                       │
│                                         │
└─────────────────────────────────────────┘

┌─────────────────────────────────────────┐
│           난이도 선택                   │
│          vs 상대닉네임                  │
├─────────────────────────────────────────┤
│                                         │
│      [ 쉬움 (Easy) ]                    │
│                                         │
│      [● 보통 (Medium) ]                 │
│                                         │
│      [ 어려움 (Hard) ]                  │
│                                         │
│         [ 게임 시작 ]                   │
│                                         │
└─────────────────────────────────────────┘
```

### 9.2 TranslationBattle.jsx 화면 구성

```
┌─────────────────────────────────────────────────────────────┐
│  Round 1/3          [ 45초 ]          나: 0승 vs 상대: 0승  │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ┌─────────────────────────────────────────────────────┐   │
│  │ 한국어 문장                                          │   │
│  │                                                      │   │
│  │   오늘 날씨가 정말 좋아서 산책하기 좋은 것 같아요   │   │
│  │                                                      │   │
│  └─────────────────────────────────────────────────────┘   │
│                                                             │
│  ┌─────────────────────────────────────────────────────┐   │
│  │ 영어 번역                                            │   │
│  │                                                      │   │
│  │   The weather is really nice today, so it seems     │   │
│  │   like a good day for a walk.                       │   │
│  │                                                      │   │
│  │                                          125/500    │   │
│  └─────────────────────────────────────────────────────┘   │
│                                                             │
│                 [ 제출 (Ctrl+Enter) ]                       │
│                                                             │
│         나: 제출 완료    상대: 응답 중...                   │
│                                                             │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│                  Round 1 결과 - 승리!                       │
│                  나: 1승 vs 상대: 0승                       │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  원문: 오늘 날씨가 정말 좋아서 산책하기 좋은 것 같아요     │
│                                                             │
│  ┌─────────────────────┐  ┌─────────────────────┐          │
│  │ 나 (93점)           │  │ 상대 (82점)         │          │
│  │                     │  │                     │          │
│  │ The weather is...   │  │ Today weather is... │          │
│  │                     │  │                     │          │
│  │ 의미: 38/40         │  │ 의미: 35/40         │          │
│  │ 문법: 28/30         │  │ 문법: 22/30         │          │
│  │ 자연: 27/30         │  │ 자연: 25/30         │          │
│  │                     │  │                     │          │
│  │ "자연스러운 표현"   │  │ "Today보다 The가    │          │
│  │                     │  │  더 자연스럽습니다" │          │
│  └─────────────────────┘  └─────────────────────┘          │
│                                                             │
│  모범 답안: The weather is really nice today, so it seems  │
│            like a good day for a walk.                      │
│                                                             │
│              다음 라운드가 곧 시작됩니다... (5초)           │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

---

## 10. 홈페이지/게임페이지 수정

### 10.1 client/src/pages/HomePage.jsx

```jsx
// games 배열에 추가
{
  id: 'translation-battle',
  title: 'Translation Battle',
  description: 'AI와 함께하는 영어 번역 대결!',
  icon: '🔤',
  available: true
}
```

### 10.2 client/src/pages/GamePage.jsx

```jsx
// import 추가
import TranslationLobby from '../components/games/TranslationBattle/TranslationLobby';

// gameComponents에 추가
const gameComponents = {
  'jump-runner': JumpRunner,
  'speed-click': SpeedClick,
  'snake': Snake,
  'memory-card': MemoryCard,
  'translation-battle': TranslationLobby  // 추가
};
```

---

## 11. 구현 체크리스트

### Phase 1: 서버 (우선순위 높음)
- [x] `server/service/groq.go` 생성
- [x] `server/model/game.go`에 Translation Battle 모델 추가
- [x] `server/model/room.go`에 TranslationRoom 및 메시지 타입 추가
- [x] `server/service/translationbattle.go` 생성
- [x] `server/handler/websocket.go`에 HandleTranslationWS 추가
- [x] `server/main.go`에 `/ws/translation` 라우트 추가

### Phase 2: 클라이언트 로비 (우선순위 중간)
- [x] `TranslationLobby.jsx` 생성
- [x] `TranslationLobby.css` 생성

### Phase 3: 클라이언트 게임 (우선순위 중간)
- [x] `TranslationBattle.jsx` 생성
- [x] `TranslationBattle.css` 생성

### Phase 4: 통합 (우선순위 중간)
- [x] `HomePage.jsx` 수정 - 게임 목록에 추가
- [x] `GamePage.jsx` 수정 - 컴포넌트 매핑 추가
- [x] `useTranslationWS.js` 훅 추가

### Phase 5: 테스트
- [ ] WebSocket 연결 테스트
- [ ] Groq API 연동 테스트
- [ ] 전체 게임 플로우 테스트
- [ ] 에러 핸들링 테스트

---

## 12. 환경 설정

### 12.1 서버 실행 전 필요사항

```bash
# Windows (PowerShell)
$env:GROQ_API_KEY="gsk_xxxxxxxxxxxxxxxx"

# Windows (CMD)
set GROQ_API_KEY=gsk_xxxxxxxxxxxxxxxx

# Linux/Mac
export GROQ_API_KEY=gsk_xxxxxxxxxxxxxxxx
```

### 12.2 Groq API 키 발급
1. https://console.groq.com 접속
2. 회원가입/로그인
3. API Keys 메뉴에서 키 생성
4. 환경 변수에 설정

---

## 13. 에러 핸들링

### 13.1 Groq API 에러
- API 키 미설정: 기본 문장 사용 ("오늘 날씨가 좋습니다.")
- Rate limit: 재시도 또는 대기
- 응답 파싱 실패: 기본 점수 (50점) 부여

### 13.2 WebSocket 에러
- 연결 끊김: 상대에게 알림, 방 상태 리셋
- 메시지 파싱 실패: 에러 메시지 전송

### 13.3 게임 로직 에러
- 이중 제출: 무시
- 잘못된 상태에서의 요청: 무시
- 방장 아닌 사람의 게임 시작: 무시

---

## 14. 향후 개선 사항 (Optional)

1. **랭킹 시스템**: 승률 기반 ELO 레이팅
2. **히스토리**: 게임 기록 저장 및 복습
3. **난이도 자동 조절**: 플레이어 실력에 맞춰 조정
4. **힌트 시스템**: 어려울 때 AI 힌트 제공 (감점)
5. **문장 카테고리**: 일상, 비즈니스, 관용어 등 선택
6. **싱글 플레이 모드**: AI와 1:1 연습 모드
7. **관전 모드**: 다른 플레이어 게임 관전

---

## 15. 참고 파일

### 15.1 기존 코드 패턴 참조
| 파일 | 참고 내용 |
|------|----------|
| `server/service/room.go` | 방 관리, 게임 루프 패턴 |
| `server/handler/websocket.go` | WebSocket 핸들러 구조 |
| `client/src/components/games/SpeedClick/BattleLobby.jsx` | 로비 UI 패턴 |
| `client/src/components/games/SpeedClick/SpeedClickBattle.jsx` | 게임 UI 패턴 |
| `client/src/hooks/useWebSocket.js` | WebSocket 훅 패턴 |

---

**계획서 작성 완료: 2026-02-01**
