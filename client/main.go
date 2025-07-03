package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/gorilla/websocket"
)

type Message struct {
	Type    string `json:"type"`
	Payload string `json:"payload"`
}

func main() {
	fmt.Print("Enter your username: ")
	reader := bufio.NewReader(os.Stdin)
	username, _ := reader.ReadString('\n')
	username = strings.TrimSpace(username)

	url := "ws://localhost:8080/ws"
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		log.Fatal("Dial error:", err)
	}
	defer conn.Close()

	// Send set_username message
	if err := conn.WriteJSON(Message{Type: "set_username", Payload: username}); err != nil {
		log.Fatal("Set username error:", err)
	}
	fmt.Println("Connected! Waiting for game...")

	inputChan := make(chan bool, 1)
	voteChan := make(chan bool, 1) // signal for voting phase

	go func() {
		for {
			var msg Message
			err := conn.ReadJSON(&msg)
			if err != nil {
				log.Println("Read error:", err)
				os.Exit(0)
			}

			// Debug print to help see incoming messages
			fmt.Printf("DEBUG: Received message type: '%s', payload: '%s'\n", msg.Type, msg.Payload)

			switch strings.TrimSpace(msg.Type) {
			case "question":
				fmt.Println("\nQuestion:", msg.Payload)
				fmt.Print("Your answer: ")
				inputChan <- true
			case "reveal_answers":
				fmt.Println("\n--- All Answers ---")
				fmt.Println(msg.Payload)
				fmt.Println("-------------------")
			case "vote":
				fmt.Println("\nVoting started! Enter the username of the player you think is the impostor:")
				voteChan <- true
			case "reveal_impostor_success":
				fmt.Printf("\nVoting Result: You guessed correctly! The impostor was: %s\n", msg.Payload)
			case "reveal_impostor_fail":
				fmt.Printf("\nVoting Result: Wrong guess! The impostor was: %s\n", msg.Payload)
			default:
				fmt.Println("Unknown message:", msg)
			}
		}
	}()

	for {
		select {
		case <-inputChan:
			answer, _ := reader.ReadString('\n')
			answer = strings.TrimSpace(answer)
			if answer == "" {
				fmt.Print("Please enter a non-empty answer: ")
				inputChan <- true
				continue
			}
			err := conn.WriteJSON(Message{Type: "submit_answer", Payload: answer})
			if err != nil {
				log.Println("Write error:", err)
				return
			}
			fmt.Println("Answer submitted, waiting for others...")

		case <-voteChan:
			vote, _ := reader.ReadString('\n')
			vote = strings.TrimSpace(vote)
			if vote == "" {
				fmt.Print("Please enter a username to vote for: ")
				voteChan <- true
				continue
			}
			err := conn.WriteJSON(Message{Type: "vote", Payload: vote})
			if err != nil {
				log.Println("Write error:", err)
				return
			}
			fmt.Println("Vote submitted, waiting for others...")
		}
	}
}
