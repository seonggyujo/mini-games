package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	"mini-games/middleware"
	"mini-games/model"
	"mini-games/service"
)

// Connection limits per IP
var (
	connPerIP    = make(map[string]int)
	connMu       sync.Mutex
	maxConnPerIP = 5
)

// Nickname validation regex for WebSocket (alphanumeric, Korean, underscore, hyphen, space)
var wsNicknameRegex = regexp.MustCompile(`^[a-zA-Z0-9가-힣_\-\s]+$`)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		// Use shared CORS origin validation
		return middleware.IsOriginAllowed(origin)
	},
}

var roomManager *service.RoomManager
var translationRoomManager *service.TranslationRoomManager

// InitWebSocket initializes the WebSocket handler
func InitWebSocket() {
	roomManager = service.NewRoomManager()
	translationRoomManager = service.NewTranslationRoomManager()
}

// validateNickname checks if the nickname is valid
func validateNickname(nickname string) bool {
	trimmed := strings.TrimSpace(nickname)
	if len(trimmed) == 0 || len(trimmed) > 20 {
		return false
	}
	return wsNicknameRegex.MatchString(trimmed)
}

// HandleBattleWS handles WebSocket connections for battle mode
func HandleBattleWS(w http.ResponseWriter, r *http.Request) {
	// Get client IP for connection limiting
	ip := middleware.GetClientIP(r)

	// Check connection limit per IP
	connMu.Lock()
	if connPerIP[ip] >= maxConnPerIP {
		connMu.Unlock()
		http.Error(w, "Too many connections from this IP", http.StatusTooManyRequests)
		return
	}
	connPerIP[ip]++
	connMu.Unlock()

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		// Decrement connection count on upgrade failure
		connMu.Lock()
		connPerIP[ip]--
		if connPerIP[ip] == 0 {
			delete(connPerIP, ip)
		}
		connMu.Unlock()
		return
	}

	// Set message size limit (4KB)
	conn.SetReadLimit(4096)

	player := &model.Player{
		Conn:       conn,
		LastActive: time.Now(),
	}

	var roomCode string
	var room *model.Room

	defer func() {
		conn.Close()
		if roomCode != "" {
			roomManager.RemovePlayerFromRoom(roomCode, player.Index)
		}
		// Decrement connection count on disconnect
		connMu.Lock()
		connPerIP[ip]--
		if connPerIP[ip] == 0 {
			delete(connPerIP, ip)
		}
		connMu.Unlock()
	}()

	// Set read deadline for ping/pong
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	// Start ping routine
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if err := conn.WriteControl(websocket.PingMessage, nil, time.Now().Add(10*time.Second)); err != nil {
					return
				}
			}
		}
	}()

	// Message handling loop
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		var msg struct {
			Type     string `json:"type"`
			RoomCode string `json:"roomCode,omitempty"`
			Nickname string `json:"nickname,omitempty"`
		}

		if err := json.Unmarshal(message, &msg); err != nil {
			player.SendJSON(model.ErrorMsg{Type: "error", Message: "잘못된 메시지 형식입니다"})
			continue
		}

		player.LastActive = time.Now()

		switch msg.Type {
		case "create":
			if !validateNickname(msg.Nickname) {
				player.SendJSON(model.ErrorMsg{Type: "error", Message: "닉네임은 1-20자의 한글, 영문, 숫자만 가능합니다"})
				continue
			}
			player.Nickname = strings.TrimSpace(msg.Nickname)
			room, roomCode = roomManager.CreateRoom(player)
			player.SendJSON(model.RoomCreatedMsg{Type: "room_created", RoomCode: roomCode})

		case "join":
			if !validateNickname(msg.Nickname) {
				player.SendJSON(model.ErrorMsg{Type: "error", Message: "닉네임은 1-20자의 한글, 영문, 숫자만 가능합니다"})
				continue
			}
			if msg.RoomCode == "" {
				player.SendJSON(model.ErrorMsg{Type: "error", Message: "방 코드를 입력해주세요"})
				continue
			}
			player.Nickname = strings.TrimSpace(msg.Nickname)
			var err error
			room, err = roomManager.JoinRoom(msg.RoomCode, player)
			if err != nil {
				player.SendJSON(model.ErrorMsg{Type: "error", Message: err.Error()})
				continue
			}
			roomCode = msg.RoomCode

			// Notify both players
			room.Players[0].SendJSON(model.OpponentJoinedMsg{Type: "opponent_joined", Nickname: player.Nickname})
			player.SendJSON(model.OpponentJoinedMsg{Type: "opponent_joined", Nickname: room.Players[0].Nickname})

			// Start game
			go roomManager.StartGame(room)

		case "click":
			if room == nil {
				continue
			}
			roomManager.HandleClick(room, player.Index)

		case "ready_rematch":
			if room == nil {
				continue
			}
			roomManager.HandleRematchReady(room, player.Index)

		case "leave":
			if roomCode != "" {
				roomManager.RemovePlayerFromRoom(roomCode, player.Index)
				roomCode = ""
				room = nil
			}

		default:
			player.SendJSON(model.ErrorMsg{Type: "error", Message: "알 수 없는 메시지 타입입니다"})
		}
	}
}

// HandleTranslationWS handles WebSocket connections for translation battle mode
func HandleTranslationWS(w http.ResponseWriter, r *http.Request) {
	// Get client IP for connection limiting
	ip := middleware.GetClientIP(r)

	// Check connection limit per IP
	connMu.Lock()
	if connPerIP[ip] >= maxConnPerIP {
		connMu.Unlock()
		http.Error(w, "Too many connections from this IP", http.StatusTooManyRequests)
		return
	}
	connPerIP[ip]++
	connMu.Unlock()

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		// Decrement connection count on upgrade failure
		connMu.Lock()
		connPerIP[ip]--
		if connPerIP[ip] == 0 {
			delete(connPerIP, ip)
		}
		connMu.Unlock()
		return
	}

	// Set message size limit (4KB)
	conn.SetReadLimit(4096)

	player := &model.Player{
		Conn:       conn,
		LastActive: time.Now(),
	}

	var roomCode string
	var room *model.TranslationRoom

	defer func() {
		conn.Close()
		if roomCode != "" {
			translationRoomManager.RemovePlayer(roomCode, player.Index)
		}
		// Decrement connection count on disconnect
		connMu.Lock()
		connPerIP[ip]--
		if connPerIP[ip] == 0 {
			delete(connPerIP, ip)
		}
		connMu.Unlock()
	}()

	// Set read deadline for ping/pong
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	// Start ping routine
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if err := conn.WriteControl(websocket.PingMessage, nil, time.Now().Add(10*time.Second)); err != nil {
					return
				}
			}
		}
	}()

	// Message handling loop
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		var msg struct {
			Type        string `json:"type"`
			RoomCode    string `json:"roomCode,omitempty"`
			Nickname    string `json:"nickname,omitempty"`
			Difficulty  string `json:"difficulty,omitempty"`
			Translation string `json:"translation,omitempty"`
		}

		if err := json.Unmarshal(message, &msg); err != nil {
			player.SendJSON(model.TErrorMsg{Type: "t_error", Message: "잘못된 메시지 형식입니다"})
			continue
		}

		player.LastActive = time.Now()

		switch msg.Type {
		case "t_create":
			if !validateNickname(msg.Nickname) {
				player.SendJSON(model.TErrorMsg{Type: "t_error", Message: "닉네임은 1-20자의 한글, 영문, 숫자만 가능합니다"})
				continue
			}
			player.Nickname = strings.TrimSpace(msg.Nickname)
			room, roomCode = translationRoomManager.CreateRoom(player)
			player.SendJSON(model.TRoomCreatedMsg{Type: "t_room_created", RoomCode: roomCode})

		case "t_join":
			if !validateNickname(msg.Nickname) {
				player.SendJSON(model.TErrorMsg{Type: "t_error", Message: "닉네임은 1-20자의 한글, 영문, 숫자만 가능합니다"})
				continue
			}
			if msg.RoomCode == "" {
				player.SendJSON(model.TErrorMsg{Type: "t_error", Message: "방 코드를 입력해주세요"})
				continue
			}
			player.Nickname = strings.TrimSpace(msg.Nickname)
			var err error
			room, err = translationRoomManager.JoinRoom(strings.ToUpper(msg.RoomCode), player)
			if err != nil {
				player.SendJSON(model.TErrorMsg{Type: "t_error", Message: err.Error()})
				continue
			}
			roomCode = strings.ToUpper(msg.RoomCode)

			// Notify both players
			room.Players[0].SendJSON(model.TOpponentJoinedMsg{Type: "t_opponent_joined", Nickname: player.Nickname})
			player.SendJSON(model.TOpponentJoinedMsg{Type: "t_opponent_joined", Nickname: room.Players[0].Nickname})

		case "t_start_game":
			if room == nil {
				player.SendJSON(model.TErrorMsg{Type: "t_error", Message: "방에 참가하지 않았습니다"})
				continue
			}
			// Only host (player 0) can start
			if player.Index != 0 {
				player.SendJSON(model.TErrorMsg{Type: "t_error", Message: "방장만 게임을 시작할 수 있습니다"})
				continue
			}
			// Parse difficulty
			difficulty := model.DifficultyMedium
			switch msg.Difficulty {
			case "easy":
				difficulty = model.DifficultyEasy
			case "hard":
				difficulty = model.DifficultyHard
			}
			go translationRoomManager.StartGame(room, difficulty)

		case "t_submit":
			if room == nil {
				continue
			}
			translationRoomManager.HandleSubmit(room, player.Index, msg.Translation)

		case "t_rematch":
			if room == nil {
				continue
			}
			translationRoomManager.HandleRematch(room, player.Index)

		case "t_leave":
			if roomCode != "" {
				translationRoomManager.RemovePlayer(roomCode, player.Index)
				roomCode = ""
				room = nil
			}

		default:
			player.SendJSON(model.TErrorMsg{Type: "t_error", Message: "알 수 없는 메시지 타입입니다"})
		}
	}
}
