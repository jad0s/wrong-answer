package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/jad0s/wrong-answer/internal/config"
)

var ImpostorConn *websocket.Conn
var answerOnce sync.Once
var voteOnce sync.Once

const VoteDuration = 180 * time.Second
const AnswerDuration = 20 * time.Second

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type Client struct {
	Conn     *websocket.Conn
	Username string
	Answered bool
	Answer   string
	Voted    bool
	Vote     string
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

func handler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer func() {
		log.Println("Connection closed")
		clientsMu.Lock()
		delete(clients, conn)
		clientsMu.Unlock()
		conn.Close()
	}()

	log.Println("Client connected")

	for {
		var msg Message
		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Println("ReadJSON: ", err)
			break
		}

		switch msg.Type {
		case "set_username":
			var username string
			if err := json.Unmarshal(msg.Payload, &username); err != nil {
				log.Println("Invalid username payload:", err)
				break
			}
			clientsMu.Lock()
			clients[conn] = &Client{Conn: conn, Username: username}
			clientsMu.Unlock()
			log.Println("User joined:", username)

		case "submit_answer":
			var answer string
			if err := json.Unmarshal(msg.Payload, &answer); err != nil {
				log.Println("Invalid answer payload:", err)
				break
			}
			clientsMu.Lock()
			client, ok := clients[conn]
			if ok {
				client.Answer = answer
				client.Answered = true
				log.Printf("%s answered: %s\n", client.Username, answer)
			}
			clientsMu.Unlock()

			if allAnswered() {
				log.Println("All players answered. Proceeding.")
				answerOnce.Do(func() {
					revealAnswers()
					vote()
				})
			}

		case "vote":
			var vote string
			if err := json.Unmarshal(msg.Payload, &vote); err != nil {
				log.Println("Invalid vote payload:", err)
				break
			}
			clientsMu.Lock()
			client, ok := clients[conn]
			if ok {
				client.Vote = strings.TrimSpace(vote)
				client.Voted = true
				log.Printf("%s voted for %s\n", client.Username, client.Vote)
			}
			clientsMu.Unlock()

			if allVoted() {
				log.Println("All players voted. Proceeding.")
				voteOnce.Do(func() {
					revealImpostor(getClientUsername(ImpostorConn), getMostVoted(), currentQuestionPair.Impostor)
				})
			}

		default:
			log.Println("Unknown message type:", msg.Type)
		}
	}
}

func main() {
	err := config.LoadConfig(config.ConfigPath)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Timer is:", config.Config["answer_timer"])

	err = config.LoadQuestions(config.QuestionsPath)
	if err != nil {
		log.Fatal("Failed to load questions:", err)
	}

	http.HandleFunc("/ws", handler)

	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			cmd := scanner.Text()
			if cmd == "start" {
				log.Println("Game starting")
				startGameRound()
			}
		}
	}()

	port := ":8080"
	log.Printf("WebSocket server started on ws://localhost%s/ws", port)
	err = http.ListenAndServe(port, nil)
	if err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}

func startGameRound() {
	answerOnce = sync.Once{}
	voteOnce = sync.Once{}

	clientsMu.Lock()
	defer clientsMu.Unlock()

	if len(clients) == 0 {
		log.Println("No players connected.")
		return
	}

	for _, client := range clients {
		client.Answered = false
		client.Voted = false
	}

	if len(config.Questions) == 0 {
		log.Println("No questions loaded. Cannot start game.")
		return
	}

	ImpostorConn = pickRandomConn(clients)
	currentQuestionPair = config.Questions[rand.Intn(len(config.Questions))]

	for conn, client := range clients {
		var questionText string
		if conn == ImpostorConn {
			questionText = fmt.Sprintf("You're the impostor. Your question is: %s", currentQuestionPair.Impostor)
			log.Printf("Impostor is: %s\n", client.Username)
		} else {
			questionText = fmt.Sprintf("Your question is: %s", currentQuestionPair.Normal)
		}
		payload, _ := json.Marshal(questionText)
		conn.WriteJSON(Message{
			Type:    "question",
			Payload: payload,
		})
	}

	duration, err := time.ParseDuration(config.Config["answer_timer"] + "s")
	if err != nil {
		log.Println("Invalid timer format:", err)
		duration = AnswerDuration
	}
	startTimer(duration, allAnswered, func() {
		answerOnce.Do(func() {
			log.Println("Answer timer expired.")
			revealAnswers()
			vote()
		})
	})
}

func vote() {
	clientsMu.Lock()
	for _, client := range clients {
		client.Voted = false
		payload, _ := json.Marshal("open")
		client.Conn.WriteJSON(Message{
			Type:    "vote",
			Payload: payload,
		})
	}
	clientsMu.Unlock()

	duration, err := time.ParseDuration(config.Config["vote_timer"] + "s")
	if err != nil {
		log.Println("Invalid vote timer format:", err)
		duration = VoteDuration
	}
	startTimer(duration, allVoted, func() {
		voteOnce.Do(func() {
			log.Println("Vote timer expired.")
			revealImpostor(getClientUsername(ImpostorConn), getMostVoted(), currentQuestionPair.Impostor)
		})
	})
}

func revealAnswers() {
	clientsMu.Lock()
	defer clientsMu.Unlock()

	answers := make(map[string]string)
	for _, client := range clients {
		answers[client.Username] = client.Answer
	}

	text := "All answers:\n"
	for user, ans := range answers {
		text += fmt.Sprintf("%s: %s\n", user, ans)
	}

	payload, _ := json.Marshal(text)
	for _, client := range clients {
		if client.Conn == ImpostorConn {
			rawPayload, err := json.Marshal(currentQuestionPair.Normal)
			if err != nil {
				log.Println("Failed to marshal question:", err)
				continue
			}
			client.Conn.WriteJSON(Message{
				Type:    "reveal_normal_question",
				Payload: rawPayload,
			})

		}
		client.Conn.WriteJSON(Message{
			Type:    "reveal_answers",
			Payload: payload,
		})
	}
}

func revealImpostor(impostor, voted, impostorQuestion string) {
	payloadStruct := RevealImpostorPayload{
		Impostor:         impostor,
		MostVoted:        voted,
		ImpostorQuestion: impostorQuestion,
	}
	payload, _ := json.Marshal(payloadStruct)

	clientsMu.Lock()
	for _, client := range clients {
		client.Conn.WriteJSON(Message{
			Type:    "reveal_impostor",
			Payload: payload,
		})
	}
	clientsMu.Unlock()
}

func pickRandomConn(clients map[*websocket.Conn]*Client) *websocket.Conn {
	if len(clients) == 0 {
		return nil
	}
	conns := make([]*websocket.Conn, 0, len(clients))
	for conn := range clients {
		conns = append(conns, conn)
	}
	rand.Seed(time.Now().UnixNano())
	return conns[rand.Intn(len(conns))]
}

func getClientUsername(conn *websocket.Conn) string {
	clientsMu.Lock()
	defer clientsMu.Unlock()
	if client, ok := clients[conn]; ok {
		return client.Username
	}
	return "Unknown"
}

func getMostVoted() string {
	votesCount := make(map[string]int)

	clientsMu.Lock()
	defer clientsMu.Unlock()
	for _, client := range clients {
		votesCount[strings.TrimSpace(client.Vote)]++
	}

	var mostVoted string
	maxVotes := 0
	for user, count := range votesCount {
		if count > maxVotes {
			maxVotes = count
			mostVoted = user
		}
	}
	return mostVoted
}

func allAnswered() bool {
	clientsMu.Lock()
	defer clientsMu.Unlock()
	for _, client := range clients {
		if !client.Answered {
			return false
		}
	}
	return true
}

func allVoted() bool {
	clientsMu.Lock()
	defer clientsMu.Unlock()
	for _, client := range clients {
		if !client.Voted {
			return false
		}
	}
	return true
}

func startTimer(duration time.Duration, shouldExpireEarly func() bool, onExpire func()) {
	go func() {
		timer := time.NewTimer(duration)
		ticker := time.NewTicker(200 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-timer.C:
				log.Println("Timer expired.")
				onExpire()
				return
			case <-ticker.C:
				if shouldExpireEarly() {
					log.Println("All players done before timer.")
					timer.Stop()
					onExpire()
					return
				}
			}
		}
	}()
}
