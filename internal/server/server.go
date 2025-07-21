package server

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/jad0s/wrong-answer/internal/config"
)

type Server struct {
	Lobbies map[string]*Lobby
	Mu      sync.RWMutex
}

type Lobby struct {
	ID      string
	Clients map[*websocket.Conn]*Client
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

var ImpostorConn *websocket.Conn
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

var clients = map[*websocket.Conn]*Client{}
var clientsMu sync.Mutex

var currentQuestionPair config.QuestionPair
