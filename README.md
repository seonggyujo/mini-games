# Mini Games Arcade

미니 게임을 즐길 수 있는 웹 아케이드

## 데모

https://mini-games.duckdns.org

## 게임 목록

- 🐍 **Snake** - 클래식 뱀 게임
- 🃏 **Memory Card** - 카드 짝 맞추기
- 🏃 **Jump Runner** - 장애물 점프 러너
- ⚡ **Speed Click** - 빠른 클릭 (실시간 대전 지원)

## 주요 기능

- 실시간 랭킹 시스템
- WebSocket 기반 실시간 대전 (Speed Click)
- 닉네임 설정 및 세션 관리

## 기술 스택

- **Frontend**: React, Webpack
- **Backend**: Go
- **Database**: SQLite

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
│   │   │       └── SpeedClick/
│   │   ├── hooks/          # 커스텀 훅 (useGameSession, useHighScore, useWebSocket)
│   │   ├── pages/          # 페이지
│   │   └── styles/         # 스타일
│   └── webpack.*.js        # Webpack 설정
│
└── server/                 # 백엔드
    ├── database/           # DB 초기화
    ├── handler/            # API 핸들러 (game, ranking, websocket)
    ├── middleware/         # 미들웨어 (cors, logging, ratelimit)
    ├── model/              # 데이터 모델 (game, room, score)
    ├── service/            # 비즈니스 로직 (게임별 서비스, room, ranking)
    └── main.go             # 서버 진입점
```
