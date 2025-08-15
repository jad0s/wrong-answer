package server

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

func (s *Server) Handler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer func() {
		log.Println("Connection closed")
		clientsMu.Lock()
		_, lobby := s.FindClientByConn(conn)
		delete(lobby.Clients, conn)
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
		case "create_lobby":
			var req struct {
				Username string `json:"username"`
			}
			if err := json.Unmarshal(msg.Payload, &req); err != nil {
				log.Println("create_lobby: invalid payload:", err)
				break
			}
			lobbyID := s.CreateLobbyID()
			_ = s.JoinLobby(lobbyID, conn, req.Username)

			resp := struct {
				LobbyID string `json:"lobby_id"`
			}{LobbyID: lobbyID}
			data, _ := json.Marshal(resp)
			conn.WriteJSON(Message{
				Type:    "lobby_created",
				Payload: data,
			})

			log.Printf("Created lobby %s for user %s\n", lobbyID, req.Username)

		case "join_lobby":
			var req struct {
				Username string `json:"username"`
				LobbyID  string `json:"lobby_id"`
			}
			if err := json.Unmarshal(msg.Payload, &req); err != nil {
				log.Println("join_lobby: invalid payload:", err)
				break
			}

			client := s.JoinLobby(req.LobbyID, conn, req.Username)
			if client == nil {
				resp := struct {
					Error string `json:"error"`
				}{Error: "lobby not found"}
				data, _ := json.Marshal(resp)
				conn.WriteJSON(Message{
					Type:    "join_failed",
					Payload: data,
				})
				break
			}

			conn.WriteJSON(Message{
				Type: "lobby_joined",
			})

			log.Printf("User %s joined lobby %s\n", req.Username, req.LobbyID)
		case "start_game":
			client, lobby := s.FindClientByConn(conn)
			if client != nil && lobby != nil {
				lobby.StartGameRound()
			} else {
				log.Println("Client not found for given connection")
			}
		case "submit_answer":
			var answer string
			if err := json.Unmarshal(msg.Payload, &answer); err != nil {
				log.Println("Invalid answer payload:", err)
				break
			}
			clientsMu.Lock()
			client, lobby := s.FindClientByConn(conn)
			if client == nil || lobby == nil {
				log.Println("Client not found for given connection")
				break
			}

			client.Answer = answer
			client.Answered = true
			log.Printf("%s answered: %s\n", client.Username, answer)

			clientsMu.Unlock()

			if lobby.allAnswered() {
				log.Println("All players answered. Proceeding.")
				answerOnce.Do(func() {
					lobby.revealAnswers()
					lobby.vote()
				})
			}

		case "vote":
			var vote string
			if err := json.Unmarshal(msg.Payload, &vote); err != nil {
				log.Println("Invalid vote payload:", err)
				break
			}
			clientsMu.Lock()
			client, lobby := s.FindClientByConn(conn)
			if client == nil || lobby == nil {
				log.Println("Client not found for given connection")
				break
			}

			client.Vote = strings.TrimSpace(vote)
			client.Voted = true
			log.Printf("%s voted for %s\n", client.Username, client.Vote)

			clientsMu.Unlock()

			if lobby.allVoted() {
				log.Println("All players voted. Proceeding.")
				voteOnce.Do(func() {
					lobby.revealImpostor(lobby.getClientUsername(lobby.ImpostorConn), lobby.getMostVoted(), lobby.currentQuestionPair.Impostor)
				})
			}

		default:
			log.Println("Unknown message type:", msg.Type)
		}
	}
}
