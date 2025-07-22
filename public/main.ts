// main.ts

const ws = new WebSocket("wss://localhost/ws");

interface Message {
  type: string;
  payload: any;
}

let username = "";
let lobbyID = "";
let isLeader = false;

// UI elements
const usernameForm     = document.getElementById("username-form")    as HTMLFormElement;
const usernameInput    = document.getElementById("username")         as HTMLInputElement;

const lobbyControls    = document.getElementById("lobby-controls")   as HTMLDivElement;
const createBtn        = document.getElementById("btn-create")       as HTMLButtonElement;
const joinBtn          = document.getElementById("btn-join")         as HTMLButtonElement;
const lobbyIdInput     = document.getElementById("lobby-id")         as HTMLInputElement;
const startBtn         = document.getElementById("btn-start")        as HTMLButtonElement;

const gameDiv          = document.getElementById("game")             as HTMLDivElement;
const questionDiv      = document.getElementById("question")         as HTMLDivElement;
const answerForm       = document.getElementById("answer-form")      as HTMLFormElement;
const answerInput      = document.getElementById("answer")           as HTMLInputElement;

const voteForm         = document.getElementById("vote-form")        as HTMLFormElement;
const voteInput        = document.getElementById("vote")             as HTMLInputElement;

const messageBox       = document.getElementById("message-box")      as HTMLDivElement;
const answersDiv       = document.getElementById("answers")          as HTMLDivElement;
const resultDiv        = document.getElementById("result")           as HTMLDivElement;

// utility
function send(type: string, payload: any) {
  ws.send(JSON.stringify({ type, payload }));
}

// after login
usernameForm.addEventListener("submit", e => {
  e.preventDefault();
  username = usernameInput.value.trim();
  if (!username) return;
  usernameForm.classList.add("hidden");
  lobbyControls.classList.remove("hidden");
});

// create lobby
createBtn.addEventListener("click", () => {
  send("create_lobby", { username });
});

// join lobby
joinBtn.addEventListener("click", () => {
  const id = lobbyIdInput.value.trim().toUpperCase();
  if (!id) return;
  send("join_lobby", { username, lobby_id: id });
});

// start game (only leader)
startBtn.addEventListener("click", () => {
  send("start_game", {});
});

// answer submission
answerForm.addEventListener("submit", e => {
  e.preventDefault();
  const ans = answerInput.value.trim();
  if (!ans) return;
  send("submit_answer", ans);
  answerForm.classList.add("hidden");
});

// vote submission
voteForm.addEventListener("submit", e => {
  e.preventDefault();
  const v = voteInput.value.trim();
  if (!v) return;
  send("vote", v);
  voteForm.classList.add("hidden");
});

// incoming messages
ws.onmessage = e => {
  let msg: Message;
  try {
    msg = JSON.parse(e.data);
  } catch {
    console.error("Invalid JSON:", e.data);
    return;
  }

  switch (msg.type) {
    case "lobby_created":
      lobbyID = msg.payload.lobby_id;
      isLeader = true;
      messageBox.textContent = `Lobby ${lobbyID} created. Waiting for players…`;
      startBtn.classList.remove("hidden");
      break;

    case "lobby_joined":
      lobbyID = msg.payload?.lobby_id || lobbyID;
      isLeader = false;
      messageBox.textContent = `Joined lobby ${lobbyID}. Waiting for leader to start…`;
      startBtn.classList.add("hidden");
      break;

    case "join_failed":
      messageBox.textContent = `Join failed: ${msg.payload.error}`;
      break;

    case "start_ack":
      // leader pressed start; game will begin soon
      lobbyControls.classList.add("hidden");
      gameDiv.classList.remove("hidden");
      break;

    case "question":
      questionDiv.textContent = msg.payload;
      answerForm.classList.remove("hidden");
      break;

    case "reveal_answers":
      answersDiv.textContent = msg.payload;
      break;

    case "vote":
      messageBox.textContent = "Voting started! Enter the name of the impostor.";
      messageBox.classList.remove("hidden");
      voteForm.classList.remove("hidden");
      break;

    case "reveal_impostor":
      const r = msg.payload as {
        impostor: string;
        most_voted: string;
        impostor_question: string;
      };
      resultDiv.innerHTML =
        `<strong>Voting Result:</strong><br>` +
        (r.impostor === r.most_voted
          ? `You guessed correctly! The impostor was <b>${r.impostor}</b>.`
          : `Wrong guess! You voted <b>${r.most_voted}</b>, but the impostor was <b>${r.impostor}</b>.`) +
        `<br><br>Impostor's question was: ${r.impostor_question}`;
      break;

    case "reveal_normal_question":
      messageBox.textContent =
        `The impostor saw a different question. The real one was: "${msg.payload}"`;
      messageBox.classList.remove("hidden");
      break;

    default:
      console.log("Unknown message:", msg);
  }
};
