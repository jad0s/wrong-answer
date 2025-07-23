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

	if len(l.Clients) == 0 {
		log.Println("No players connected in lobby:", l.ID)
		return
	}

	for _, client := range l.Clients {
		client.Answered = false
		client.Voted = false
	}

	if len(config.Questions) == 0 {
		log.Println("No questions loaded. Cannot start game in lobby:", l.ID)
		return
	}

	l.ImpostorConn = l.pickRandomConn(l.Clients)
	l.currentQuestionPair = config.Questions[rand.Intn(len(config.Questions))]

	for conn, client := range l.Clients {
		var questionText string
		if conn == l.ImpostorConn {
			questionText = fmt.Sprintf("You're the impostor. Your question is: %s", l.currentQuestionPair.Impostor)
			log.Printf("Impostor is: %s in lobby: %s\n", client.Username, l.ID)
		} else {
			questionText = fmt.Sprintf("Your question is: %s", l.currentQuestionPair.Normal)
		}
		payload, _ := json.Marshal(questionText)
		conn.WriteJSON(Message{
			Type:    "question",
			Payload: payload,
		})
	}

	duration, err := time.ParseDuration(config.Config["answer_timer"] + "s")
	if err != nil {
		log.Printf("Invalid timer format:%e in lobby: %s\n", err, l.ID)
		duration = AnswerDuration
	}
	startTimer(duration, l.allAnswered, func() {
		answerOnce.Do(func() {
			log.Println("Answer timer expired in lobby:", l.ID)
			l.revealAnswers()
			l.vote()
		})
	})
}

func (l *Lobby) vote() {
	clientsMu.Lock()
	for _, client := range l.Clients {
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
	startTimer(duration, l.allVoted, func() {
		voteOnce.Do(func() {
			log.Println("Vote timer expired.")
			l.revealImpostor(l.getClientUsername(l.ImpostorConn), l.getMostVoted(), l.currentQuestionPair.Impostor)
		})
	})
}

func (l *Lobby) revealAnswers() {
	clientsMu.Lock()
	defer clientsMu.Unlock()

	answers := make(map[string]string)
	for _, client := range l.Clients {
		answers[client.Username] = client.Answer
	}

	text := "All answers:\n"
	for user, ans := range answers {
		text += fmt.Sprintf("%s: %s\n", user, ans)
	}

	payload, _ := json.Marshal(text)
	for _, client := range l.Clients {
		if client.Conn == l.ImpostorConn {
			rawPayload, err := json.Marshal(l.currentQuestionPair.Normal)
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

func (l *Lobby) revealImpostor(impostor, voted, impostorQuestion string) {
	payloadStruct := RevealImpostorPayload{
		Impostor:         impostor,
		MostVoted:        voted,
		ImpostorQuestion: impostorQuestion,
	}
	payload, _ := json.Marshal(payloadStruct)

	clientsMu.Lock()
	for _, client := range l.Clients {
		client.Conn.WriteJSON(Message{
			Type:    "reveal_impostor",
			Payload: payload,
		})
	}
	clientsMu.Unlock()
}

func (l *Lobby) pickRandomConn(clients map[*websocket.Conn]*Client) *websocket.Conn {
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

func (l *Lobby) getClientUsername(conn *websocket.Conn) string {
	clientsMu.Lock()
	defer clientsMu.Unlock()
	if client, ok := l.Clients[conn]; ok {
		return client.Username
	}
	return "Unknown"
}

func (l *Lobby) getMostVoted() string {
	votesCount := make(map[string]int)

	clientsMu.Lock()
	defer clientsMu.Unlock()
	for _, client := range l.Clients {
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
