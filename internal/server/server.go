package server

import (
	"crypto/rand"
	"encoding/json"
	"math/big"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/jad0s/wrong-answer/internal/config"
)

const idLength = 6
const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

type Server struct {
	Lobbies map[string]*Lobby
	Mu      sync.RWMutex
}

type Lobby struct {
	ID                  string
	Clients             map[*websocket.Conn]*Client
	ImpostorConn        *websocket.Conn
	currentQuestionPair config.QuestionPair
	Mu                  sync.RWMutex
	Leader              *Client
}

type Client struct {
	Conn     *websocket.Conn
	Username string
	Lobby    *Lobby
	Answered bool
	Answer   string
	Voted    bool
	Vote     string
}

var answerOnce sync.Once
var voteOnce sync.Once

const VoteDuration = 180 * time.Second
const AnswerDuration = 20 * time.Second

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Message struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type RevealImpostorPayload struct {
	Impostor         string `json:"impostor"`
	MostVoted        string `json:"most_voted"`
	ImpostorQuestion string `json:"impostor_question"`
}

var clientsMu sync.Mutex

func (s *Server) JoinLobby(id string, conn *websocket.Conn, username string) *Client {
	s.Mu.Lock()
	lobby, exists := s.Lobbies[id]
	if !exists {
		lobby = &Lobby{ID: id, Clients: make(map[*websocket.Conn]*Client)}
		s.Lobbies[id] = lobby
	}
	s.Mu.Unlock()

	client := &Client{Conn: conn, Username: username, Lobby: lobby}

	lobby.Mu.Lock()
	lobby.Clients[conn] = client
	if !exists {
		lobby.Leader = client
	}
	lobby.Mu.Unlock()

	return client
}

func GenerateLobbyID() string {
	id := make([]byte, idLength)
	for i := range id {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		id[i] = charset[n.Int64()]
	}
	return string(id)
}

func (s *Server) CreateLobbyID() string {
	for {
		id := GenerateLobbyID()
		s.Mu.RLock()
		_, exists := s.Lobbies[id]
		s.Mu.RUnlock()
		if !exists {
			return id
		}
	}
}

func (s *Server) FindClientByConn(conn *websocket.Conn) (*Client, *Lobby) {
	for _, lobby := range s.Lobbies {
		if client, ok := lobby.Clients[conn]; ok {
			return client, lobby
		}
	}
	return nil, nil
}
