# Mini Games Arcade

미니 게임을 즐길 수 있는 웹 아케이드

## 데모

https://mini-games.duckdns.org

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
│   │   ├── components/     # React 컴포넌트
│   │   ├── pages/          # 페이지
│   │   └── styles/         # 스타일
│   └── webpack.*.js        # Webpack 설정
│
└── server/                 # 백엔드
    ├── database/           # DB 초기화
    ├── handler/            # API 핸들러
    ├── middleware/         # 미들웨어
    ├── model/              # 데이터 모델
    ├── service/            # 비즈니스 로직
    └── main.go             # 서버 진입점
```
