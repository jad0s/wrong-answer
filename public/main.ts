// main.ts

// --- Build the right WS URL (ws:// for http, wss:// for https) ---
const WS_URL = (() => {
  const proto = location.protocol === "https:" ? "wss" : "ws";
  // e.g. "localhost:443" or "192.168.0.99:8080"
  return `${proto}://${location.host}/ws`;
})();

const ws = new WebSocket(WS_URL);

// --- Types ---
interface WsMessage<T = any> {
  type: string;
  payload: T;
}

type RevealImpostorPayload = {
  impostor: string;
  most_voted: string;
  impostor_question: string;
};

// --- State ---
let username = "";
let lobbyID = "";
let isLeader = false;

// --- UI elements ---
const usernameForm  = document.getElementById("username-form")  as HTMLFormElement;
const usernameInput = document.getElementById("username")       as HTMLInputElement;

const lobbyControls = document.getElementById("lobby-controls") as HTMLDivElement;
const createBtn     = document.getElementById("btn-create")     as HTMLButtonElement;
const joinBtn       = document.getElementById("btn-join")       as HTMLButtonElement;
const lobbyIdInput  = document.getElementById("lobby-id")       as HTMLInputElement;
const startBtn      = document.getElementById("btn-start")      as HTMLButtonElement;

const gameDiv       = document.getElementById("game")           as HTMLDivElement;
const messageBox    = document.getElementById("message-box")    as HTMLDivElement;
const questionDiv   = document.getElementById("question")       as HTMLDivElement;

const answerForm    = document.getElementById("answer-form")    as HTMLFormElement;
const answerInput   = document.getElementById("answer")         as HTMLInputElement;

const answersDiv    = document.getElementById("answers")        as HTMLDivElement;

const voteForm      = document.getElementById("vote-form")      as HTMLFormElement;
const voteInput     = document.getElementById("vote")           as HTMLInputElement;

const resultDiv     = document.getElementById("result")         as HTMLDivElement;

// --- Utils ---
function show(el: HTMLElement)  { el.classList.remove("hidden"); }
function hide(el: HTMLElement)  { el.classList.add("hidden"); }
function setText(el: HTMLElement, text: string) { el.textContent = text; }

function payloadToString(p: any): string {
  if (p == null) return "";
  if (typeof p === "string") return p;
  try { return JSON.stringify(p); } catch { return String(p); }
}

function renderAnswers(payload: any) {
  // Accepts string, array, or object map {username: answer}
  if (typeof payload === "string") {
    // If server sends a preformatted string with newlines
    answersDiv.innerHTML = `<pre>${payload.replace(/</g, "&lt;")}</pre>`;
    return;
  }
  if (Array.isArray(payload)) {
    answersDiv.innerHTML = payload.map(a => `<div>${String(a)}</div>`).join("");
    return;
  }
  if (typeof payload === "object") {
    const rows = Object.entries(payload).map(([user, ans]) =>
      `<div><b>${user}:</b> ${String(ans)}</div>`
    );
    answersDiv.innerHTML = rows.join("");
    return;
  }
  answersDiv.textContent = String(payload);
}

// Queue messages if ws isn’t open yet
const sendQueue: string[] = [];
ws.addEventListener("open", () => {
  if (sendQueue.length) {
    for (const m of sendQueue) ws.send(m);
    sendQueue.length = 0;
  }
});

function send(type: string, payload: any) {
  const msg = JSON.stringify({ type, payload });
  if (ws.readyState === WebSocket.OPEN) ws.send(msg);
  else sendQueue.push(msg);
}

// --- UI: Username submit ---
usernameForm.addEventListener("submit", (e) => {
  e.preventDefault();
  username = usernameInput.value.trim();
  if (!username) return;

  hide(usernameForm);
  show(lobbyControls);
  messageBox.textContent = "Create a lobby or join an existing code.";
});

// --- UI: Create Lobby ---
createBtn.addEventListener("click", () => {
  if (!username) return;
  send("create_lobby", { username });
});

// --- UI: Join Lobby ---
joinBtn.addEventListener("click", () => {
  if (!username) return;
  const id = lobbyIdInput.value.trim().toUpperCase();
  if (!id) return;
  send("join_lobby", { username, lobby_id: id });
});

// --- UI: Start Game (leader only) ---
startBtn.addEventListener("click", () => {
  if (!isLeader) return;
  startBtn.disabled = true; // avoid double-clicks
  send("start_game", {});
});

// --- UI: Submit Answer ---
answerForm.addEventListener("submit", (e) => {
  e.preventDefault();
  const ans = answerInput.value.trim();
  if (!ans) return;
  send("submit_answer", ans);
  answerInput.value = "";
  hide(answerForm);
});

// --- UI: Submit Vote ---
voteForm.addEventListener("submit", (e) => {
  e.preventDefault();
  const v = voteInput.value.trim();
  if (!v) return;
  send("vote", v);
  voteInput.value = "";
  hide(voteForm);
});

// --- WS Events ---
ws.addEventListener("close", () => {
  setText(messageBox, "Disconnected from server.");
});

ws.addEventListener("error", (e) => {
  console.error("WebSocket error:", e);
});

// --- WS Incoming Messages ---
ws.onmessage = (evt) => {
  let msg: WsMessage;
  try {
    msg = JSON.parse(evt.data);
  } catch {
    console.error("Invalid JSON:", evt.data);
    return;
  }

  switch (msg.type) {
    // Lobby lifecycle
    case "lobby_created": {
      // payload: { lobby_id: string }
      lobbyID = msg.payload?.lobby_id || lobbyID;
      isLeader = true;
      startBtn.disabled = false;
      show(startBtn);
      setText(messageBox, `Lobby ${lobbyID} created. Waiting for players…`);
      if (lobbyIdInput) lobbyIdInput.value = lobbyID; // convenience
      break;
    }

    case "lobby_joined": {
      // payload may have { lobby_id }
      lobbyID = msg.payload?.lobby_id || lobbyID;
      isLeader = false;
      hide(startBtn);
      setText(messageBox, `Joined lobby ${lobbyID}. Waiting for leader to start…`);
      break;
    }

    case "join_failed": {
      // payload: { error: string }
      setText(messageBox, `Join failed: ${payloadToString(msg.payload?.error || msg.payload)}`);
      break;
    }

    case "start_ack": {
      // Optional phase. If server sends it, switch UI now.
      hide(lobbyControls);
      show(gameDiv);
      setText(messageBox, "Game starting…");
      break;
    }

    // Game flow
    case "question": {
      // Some servers skip "start_ack" and go straight to question. Handle both.
      hide(lobbyControls);
      show(gameDiv);

      const q = payloadToString(msg.payload);
      questionDiv.textContent = q;
      show(answerForm);

      // Clear previous round state
      answersDiv.innerHTML = "";
      resultDiv.innerHTML = "";
      hide(voteForm);
      break;
    }

    case "reveal_answers": {
      renderAnswers(msg.payload);
      break;
    }

    case "vote": {
      // payload often "open"; don’t care. Just show UI.
      setText(messageBox, "Voting started! Enter the name of the impostor.");
      show(voteForm);
      break;
    }

    case "reveal_impostor": {
      const r = msg.payload as RevealImpostorPayload;
      const good = r && r.impostor === r.most_voted;

      resultDiv.innerHTML =
        `<strong>Voting Result:</strong><br>` +
        (good
          ? `You guessed correctly! The impostor was <b>${r.impostor}</b>.`
          : `Wrong guess! You voted <b>${r.most_voted}</b>, but the impostor was <b>${r.impostor}</b>.`) +
        `<br><br>Impostor's question was: ${r.impostor_question}`;

      break;
    }

    case "reveal_normal_question": {
      // payload: string (the real/normal question)
      setText(messageBox, `The impostor saw a different question. The real one was: "${payloadToString(msg.payload)}"`);
      show(messageBox);
      break;
    }

    // Optional generic error
    case "error": {
      setText(messageBox, `Error: ${payloadToString(msg.payload)}`);
      break;
    }

    default:
      console.log("Unknown message:", msg);
  }
};
