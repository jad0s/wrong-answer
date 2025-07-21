package server

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

func Handler(w http.ResponseWriter, r *http.Request) {
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
