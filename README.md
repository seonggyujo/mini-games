# Mini Games Arcade

미니 게임을 즐길 수 있는 웹 아케이드

## 데모

https://mini-games.duckdns.org

## 게임 목록

- 🐍 **Snake** - 클래식 뱀 게임
- 🃏 **Memory Card** - 카드 짝 맞추기
- 🏃 **Jump Runner** - 장애물 점프 러너
- ⚡ **Speed Click** - 빠른 클릭 (실시간 대전 지원)
- 🌐 **Translation Battle** - AI와 함께하는 1v1 영어 번역 대결

## 주요 기능

- 실시간 랭킹 시스템
- WebSocket 기반 실시간 대전 (Speed Click, Translation Battle)
- 닉네임 설정 및 세션 관리
- AI 기반 번역 평가 시스템 (Groq API)

## Translation Battle

1v1 실시간 영어 번역 대결 게임

### 게임 방식
- 3판 2선승제
- AI가 생성한 한국어 문장을 영어로 번역
- AI가 두 번역을 평가하여 승자 결정

### 평가 기준
| 항목 | 배점 | 설명 |
|------|------|------|
| 의미 전달 | 40점 | 원문의 의미를 정확히 전달했는지 |
| 문법 | 30점 | 영어 문법이 올바른지 |
| 자연스러움 | 30점 | 원어민이 사용하는 자연스러운 표현인지 |

### 난이도
| 난이도 | 설명 |
|--------|------|
| 쉬움 | 3-5 단어, 초등학생 수준 |
| 보통 | 5-8 단어, TOPIK 1-2급 |
| 어려움 | 8-12 단어, TOPIK 3-4급 |

### 특징
- 시간 종료 시 자동 제출
- 주제/상황/감정/행동 랜덤 조합으로 매번 다양한 문장 생성
- 라운드별 AI 피드백 및 모범 답안 제공
- 효과음 시스템 (음소거 가능)

## 기술 스택

- **Frontend**: React 18, Webpack 5
- **Backend**: Go
- **Database**: SQLite
- **실시간 통신**: WebSocket (gorilla/websocket)
- **AI API**: Groq API (openai/gpt-oss-120b)

## 프로젝트 구조

```
mini-games/
├── client/                 # 프론트엔드
│   ├── public/             # 정적 파일
│   ├── src/
│   │   ├── components/
│   │   │   ├── common/     # 공통 컴포넌트 (NavBar, NicknameModal, Toast)
│   │   │   └── games/      # 게임 컴포넌트
│   │   │       ├── Snake/
│   │   │       ├── MemoryCard/
│   │   │       ├── JumpRunner/
│   │   │       ├── SpeedClick/
│   │   │       └── TranslationBattle/
│   │   ├── hooks/          # 커스텀 훅 (useGameSession, useHighScore, useWebSocket, useTranslationWS, useSoundEffects)
│   │   ├── pages/          # 페이지
│   │   └── styles/         # 스타일
│   └── webpack.*.js        # Webpack 설정
│
└── server/                 # 백엔드
    ├── database/           # DB 초기화
    ├── handler/            # API 핸들러 (game, ranking, websocket)
    ├── middleware/         # 미들웨어 (cors, logging, ratelimit)
    ├── model/              # 데이터 모델 (game, room, score)
    ├── service/            # 비즈니스 로직 (게임별 서비스, room, ranking, groq, translationbattle)
    └── main.go             # 서버 진입점
```

## 환경 변수

| 변수 | 설명 |
|------|------|
| `GROQ_API_KEY` | Groq API 키 (Translation Battle에 필요) |
| `PORT` | 서버 포트 (기본값: 4001) |

## 로컬 개발

### 클라이언트
```bash
cd client
npm install
npm run dev
```

### 서버
```bash
cd server
export GROQ_API_KEY=your_api_key
CGO_ENABLED=1 go run .
```

## 배포

```bash
cd ~/mini-games
git pull origin main
cd client && npm run build
cd ../server && CGO_ENABLED=1 go build -o mini-games-server .
sudo systemctl restart mini-games
```
