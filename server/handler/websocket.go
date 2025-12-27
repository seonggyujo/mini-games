package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"

	"mini-games/model"
	"mini-games/service"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for development
	},
}

var roomManager *service.RoomManager

// InitWebSocket initializes the WebSocket handler
func InitWebSocket() {
	roomManager = service.NewRoomManager()
}

// HandleBattleWS handles WebSocket connections for battle mode
func HandleBattleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

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
			if msg.Nickname == "" {
				player.SendJSON(model.ErrorMsg{Type: "error", Message: "닉네임을 입력해주세요"})
				continue
			}
			player.Nickname = msg.Nickname
			room, roomCode = roomManager.CreateRoom(player)
			player.SendJSON(model.RoomCreatedMsg{Type: "room_created", RoomCode: roomCode})

		case "join":
			if msg.Nickname == "" {
				player.SendJSON(model.ErrorMsg{Type: "error", Message: "닉네임을 입력해주세요"})
				continue
			}
			if msg.RoomCode == "" {
				player.SendJSON(model.ErrorMsg{Type: "error", Message: "방 코드를 입력해주세요"})
				continue
			}
			player.Nickname = msg.Nickname
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
