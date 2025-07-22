package server

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/jad0s/wrong-answer/internal/config"
)

func (l *Lobby) StartGameRound() {
	answerOnce = sync.Once{}
	voteOnce = sync.Once{}

	clientsMu.Lock()
	defer clientsMu.Unlock()

	if len(clients) == 0 {
		log.Println("No players connected.")
		return
	}

	for _, client := range l.Clients {
		client.Answered = false
		client.Voted = false
	}

	if len(config.Questions) == 0 {
		log.Println("No questions loaded. Cannot start game.")
		return
	}

	ImpostorConn = pickRandomConn(l.Clients)
	currentQuestionPair = config.Questions[rand.Intn(len(config.Questions))]

	for conn, client := range l.Clients {
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
