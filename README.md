# Wrong Answer

A local multiplayer deception game in the terminal. One player receives a slightly different question — and everyone has to figure out who it was based on their answer.

The game is currently in a very early state, and not ready for production. I'm currently focusing on the server side, the frontend is AI generated and wonky.
Once I feel the server is in a good enough place, I will rewrite the frontend myself. 

## Features

- WebSocket server for real-time multiplayer
- Web server serving the client interface
- Round-based question → answer → vote → reveal flow
- Configurable questions, usernames, and server URLs

## Building

### Build from source (Go 1.21+)

```bash
git clone https://github.com/jad0s/wrong-answer
cd wrong-answer

go build

```

## Running the Game

### Start the server

```bash
./wrong-answer
```



## Configuration

The server loads configuration from:

```
~/.config/wrong-answer-server/config.yaml
```
All options must be wrapped in quotes.

Example:

```yaml
port: "8080"
answer_timer: "20"
vote_timer: "180"
```


## Custom Questions

You can create your own question set in a JSON file, which is stored in the same directory as config.yaml, under the name "questions.json". Format:

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

- No player limit, but game is designed for 3+ players

## License

MIT License

## Author

https://github.com/jad0s
