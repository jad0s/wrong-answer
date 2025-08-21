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
	s := &server.Server{Lobbies: make(map[string]*server.Lobby)}
	mux.Handle("/ws", http.HandlerFunc(s.Handler))
	mux.Handle("/", http.FileServer(http.Dir("./public")))

	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			cmd := scanner.Text()
			if cmd == "start" {
				log.Println("Game starting")
			}
		}
	}()

	go func() {
		log.Printf("TLS server on https://localhost:443 and wss://localhost:443/ws\n")
		srv := &http.Server{
			Addr:    ":443",
			Handler: mux,
		}
		log.Fatal(srv.ListenAndServeTLS("cert.pem", "key.pem"))

	}()

	log.Println("QUIC (HTTP/3) listening on udp://:443")
	log.Fatal(http3.ListenAndServeQUIC(
		":443",
		"cert.pem",
		"key.pem",
		mux,
	))
}
