# Wrong Answer

A local multiplayer deception game in the terminal. One player receives a slightly different question — and everyone has to figure out who it was based on their answer.

## How to Play

1. Start the server.
2. Each player runs the client and joins with a username.
3. All players receive a question — except one impostor.
4. Everyone submits a one-word answer.
5. All answers are revealed.
6. Everyone votes on who they think the impostor is.
7. The game reveals whether the vote was correct.

## Features

- WebSocket server for real-time multiplayer
- Terminal client with minimal dependencies
- Round-based question → answer → vote → reveal flow
- Auto-update system (optional, via config)
- Configurable questions, usernames, and server URLs

## Building

### Build from source (Go 1.21+)

```bash
git clone https://github.com/yourname/wrong-answer.git
cd wrong-answer

# Build the server
go build -o wrong-answer-server ./server

# Build the client
go build -o wrong-answer-client ./client
```

## Running the Game

### Start the server

```bash
./wrong-answer-server
```

In the server terminal, type `start` to begin a round after all players have joined.

### Run the client

```bash
./wrong-answer-client
```

The client will prompt for a username, then wait for a question.

## Configuration

The client loads configuration from:

```
~/.config/wrong-answer/config.yaml
```

Example:

```yaml
username: jados
server: ws://localhost:8080/ws
auto_update: true
```

If `auto_update` is enabled, the client will check GitHub for newer releases and offer to update itself.

## Installation (Optional)

### Fedora COPR

```bash
sudo dnf copr enable jados/wrong-answer
sudo dnf install wrong-answer-client wrong-answer-server
```

The COPR build disables auto-update by default.

## Custom Questions

You can create your own question set in a JSON file. Format:

```json
[
  {
    "normal": "What’s your favorite fruit?",
    "impostor": "What’s your favorite vegetable?"
  },
  {
    "normal": "Where do you go on vacation?",
    "impostor": "Where do you go to work?"
  }
]
```

The server can be configured to load these instead of hardcoded questions.

## Notes

- Auto-update is disabled in the Fedora COPR release.
- The game is local only — for remote games, use a tunnel (e.g. Tailscale, localtunnel, playit.gg).
- No player limit, but game is designed for 3+ players

## License

MIT License

## Author

https://github.com/jad0s
