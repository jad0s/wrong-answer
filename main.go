package main

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/jad0s/wrong-answer/internal/config"
	"github.com/jad0s/wrong-answer/internal/server"
	"github.com/quic-go/quic-go/http3"
)

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

	mux := http.NewServeMux()
	mux.Handle("/ws", http.HandlerFunc(server.Handler))
	mux.Handle("/", http.FileServer(http.Dir("./public")))

	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			cmd := scanner.Text()
			if cmd == "start" {
				log.Println("Game starting")
				server.StartGameRound()
			}
		}
	}()

	port := ":8080"
	log.Printf("WebSocket server started on ws://localhost%s/ws", port)
	go func() {
		log.Fatal(http.ListenAndServe(port, nil))
	}()

	log.Println("QUIC (HTTP/3) listening on udp://:443")
	if err := http3.ListenAndServeTLS(
		":443",
		"cert.pem", // your cert file
		"key.pem",  // your key file
		mux,
	); err != nil {
		log.Fatal(err)
	}

}
