package main

import (
	"bufio"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/jad0s/wrong-answer/internal/config"
)

var ImpostorConn *websocket.Conn

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

var clients = map[*websocket.Conn]*Client{}

type Message struct {
	Type    string `json:"type"`
	Payload string `json:"payload"`
}

func handler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer func() {
		log.Println("Connection closed")
		delete(clients, conn)
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
			username := msg.Payload
			clients[conn] = &Client{Conn: conn, Username: username}
			log.Println("User joined: ", username)
		case "submit_answer":
			client, ok := clients[conn]
			if !ok {
				log.Println("Unknown connection tried sending a message")
				continue
			}
			log.Printf("%s answered: %s\n", client.Username, msg.Payload)
			client.Answer = msg.Payload
			client.Answered = true
			if allAnswered() {
				log.Println("All players answered. proceeding")
				revealAnswers()
				vote()
			}
		case "vote":
			client, ok := clients[conn]
			if !ok {
				log.Println("Unknown connection tried voting")
				continue
			}
			vote := strings.TrimSpace(msg.Payload)
			log.Printf("%s voted for %s\n", client.Username, vote)
			client.Voted = true
			client.Vote = vote
			if allVoted() {
				log.Println("All players voted. proceeding")
				mostVoted := getMostVoted()
				impostorUsername := clients[ImpostorConn].Username
				revealImpostor(mostVoted, impostorUsername)
			}
		default:
			log.Println("Unknown message type: ", msg.Type)
		}
	}

}

func main() {
	fmt.Println("Loaded config from:", config.ConfigPath)
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
	log.Printf("WebSocker server started on ws://localhost%s/ws", port)
	err := http.ListenAndServe(port, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func startGameRound() {
	for _, client := range clients {
		client.Answered = false
	}
	ImpostorConn = pickRandomConn(clients)
	if ImpostorConn == nil {
		log.Println("No players connected.")
		return
	}
	impostor_question := "imposter"
	question := "not imposter"
	for conn, client := range clients {
		if conn == ImpostorConn {
			question := fmt.Sprintf("You're the impostor. Your question is: %s", impostor_question)
			msg := Message{
				Type:    "question",
				Payload: question,
			}
			conn.WriteJSON(msg)
			log.Printf("Impostor is: %s\n", client.Username)
		} else {
			question := fmt.Sprintf("Your question is: %s", question)
			msg := Message{
				Type:    "question",
				Payload: question,
			}
			conn.WriteJSON(msg)
		}
	}
}

func pickRandomConn(clients map[*websocket.Conn]*Client) *websocket.Conn {
	rand.Seed(time.Now().UnixNano())

	conns := make([]*websocket.Conn, 0, len(clients))
	for conn := range clients {
		conns = append(conns, conn)
	}

	if len(conns) == 0 {
		return nil
	}

	idx := rand.Intn(len(conns))
	return conns[idx]
}

func allAnswered() bool {
	for _, client := range clients {
		if !client.Answered {
			return false
		}

	}
	return true
}

func revealAnswers() {

	answers := make(map[string]string)
	for _, client := range clients {
		answers[client.Username] = client.Answer
	}

	for _, client := range clients {
		msg := Message{
			Type:    "reveal_answers",
			Payload: formatAnswers(answers),
		}
		err := client.Conn.WriteJSON(msg)
		if err != nil {
			log.Println("Error sending reveal to", client.Username, err)
		}
	}
}

func formatAnswers(answers map[string]string) string {
	result := "All answers:\n"
	for user, ans := range answers {
		result += fmt.Sprintf("%s: %s\n", user, ans)
	}
	return result
}

func vote() {
	for _, client := range clients {
		client.Voted = false
		err := client.Conn.WriteJSON(Message{Type: "vote", Payload: "open"})
		if err != nil {
			log.Println("Error sending vote opening to", client.Username, err)
		}
	}
}

func allVoted() bool {
	for _, client := range clients {
		if !client.Voted {
			return false
		}
	}
	return true
}

func revealImpostor(impostor string, voted string) {
	var msg Message
	if impostor == voted {
		msg = Message{
			Type:    "reveal_impostor_success",
			Payload: impostor,
		}
	} else {
		msg = Message{
			Type:    "reveal_impostor_fail",
			Payload: impostor,
		}
	}

	for _, client := range clients {
		err := client.Conn.WriteJSON(msg)
		if err != nil {
			log.Println("Error sending impostor reveal to", client.Username, err)
		}
	}
}

func getMostVoted() string {
	votesCount := make(map[string]int)
	for _, client := range clients {
		vote := strings.TrimSpace(client.Vote)
		votesCount[vote]++
	}
	var maxCount int
	var mostVoted string
	for user, count := range votesCount {
		if count > maxCount {
			maxCount = count
			mostVoted = user
		}
	}
	return mostVoted
}
