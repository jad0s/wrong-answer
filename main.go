package main

import (
	"bufio"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type Client struct {
	Conn     *websocket.Conn
	Username string
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
				log.Println("Unknown connectiion tried sending a message")
				continue
			}
			log.Printf("%s answered: %s\n", client.Username, msg.Payload)
		default:
			log.Println("Unknown message type: ", msg.Type)
		}
	}

}

func main() {
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
	impostorConn := pickRandomConn(clients)
	if impostorConn == nil {
		log.Println("No players connected.")
		return
	}
	for conn, client := range clients {
		if conn == impostorConn {
			conn.WriteJSON(Message{Type: "question", Payload: "You're the impostor. Your number is 13."})
			log.Printf("Impostor is: %s\n", client.Username)
		} else {
			conn.WriteJSON(Message{Type: "question", Payload: "Your number is 42."})
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
