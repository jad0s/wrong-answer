"use strict";
// main.ts
// --- Build the right WS URL (ws:// for http, wss:// for https) ---
const WS_URL = (() => {
    const proto = location.protocol === "https:" ? "wss" : "ws";
    // e.g. "localhost:443" or "192.168.0.99:8080"
    return `${proto}://${location.host}/ws`;
})();
const ws = new WebSocket(WS_URL);
// --- State ---
let username = "";
let lobbyID = "";
let isLeader = false;
// --- UI elements ---
const usernameForm = document.getElementById("username-form");
const usernameInput = document.getElementById("username");
const lobbyControls = document.getElementById("lobby-controls");
const createBtn = document.getElementById("btn-create");
const joinBtn = document.getElementById("btn-join");
const lobbyIdInput = document.getElementById("lobby-id");
const startBtn = document.getElementById("btn-start");
const gameDiv = document.getElementById("game");
const messageBox = document.getElementById("message-box");
const questionDiv = document.getElementById("question");
const answerForm = document.getElementById("answer-form");
const answerInput = document.getElementById("answer");
const answersDiv = document.getElementById("answers");
const voteForm = document.getElementById("vote-form");
const voteInput = document.getElementById("vote");
const resultDiv = document.getElementById("result");
// --- Utils ---
function show(el) { el.classList.remove("hidden"); }
function hide(el) { el.classList.add("hidden"); }
function setText(el, text) { el.textContent = text; }
function payloadToString(p) {
    if (p == null)
        return "";
    if (typeof p === "string")
        return p;
    try {
        return JSON.stringify(p);
    }
    catch (_a) {
        return String(p);
    }
}
function renderAnswers(payload) {
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
        const rows = Object.entries(payload).map(([user, ans]) => `<div><b>${user}:</b> ${String(ans)}</div>`);
        answersDiv.innerHTML = rows.join("");
        return;
    }
    answersDiv.textContent = String(payload);
}
// Queue messages if ws isn’t open yet
const sendQueue = [];
ws.addEventListener("open", () => {
    if (sendQueue.length) {
        for (const m of sendQueue)
            ws.send(m);
        sendQueue.length = 0;
    }
});
function send(type, payload) {
    const msg = JSON.stringify({ type, payload });
    if (ws.readyState === WebSocket.OPEN)
        ws.send(msg);
    else
        sendQueue.push(msg);
}
// --- UI: Username submit ---
usernameForm.addEventListener("submit", (e) => {
    e.preventDefault();
    username = usernameInput.value.trim();
    if (!username)
        return;
    hide(usernameForm);
    show(lobbyControls);
    messageBox.textContent = "Create a lobby or join an existing code.";
});
// --- UI: Create Lobby ---
createBtn.addEventListener("click", () => {
    if (!username)
        return;
    send("create_lobby", { username });
});
// --- UI: Join Lobby ---
joinBtn.addEventListener("click", () => {
    if (!username)
        return;
    const id = lobbyIdInput.value.trim().toUpperCase();
    if (!id)
        return;
    send("join_lobby", { username, lobby_id: id });
});
// --- UI: Start Game (leader only) ---
startBtn.addEventListener("click", () => {
    if (!isLeader)
        return;
    startBtn.disabled = true; // avoid double-clicks
    send("start_game", {});
});
// --- UI: Submit Answer ---
answerForm.addEventListener("submit", (e) => {
    e.preventDefault();
    const ans = answerInput.value.trim();
    if (!ans)
        return;
    send("submit_answer", ans);
    answerInput.value = "";
    hide(answerForm);
});
// --- UI: Submit Vote ---
voteForm.addEventListener("submit", (e) => {
    e.preventDefault();
    const v = voteInput.value.trim();
    if (!v)
        return;
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
    var _a, _b, _c;
    let msg;
    try {
        msg = JSON.parse(evt.data);
    }
    catch (_d) {
        console.error("Invalid JSON:", evt.data);
        return;
    }
    console.log(msg)
    switch (msg.type) {
        
        // Lobby lifecycle
        case "lobby_created": {
            // payload: { lobby_id: string }
            lobbyID = ((_a = msg.payload) === null || _a === void 0 ? void 0 : _a.lobby_id) || lobbyID;
            isLeader = true;
            startBtn.disabled = false;
            show(startBtn);
            setText(messageBox, `Lobby ${lobbyID} created. Waiting for players…`);
            if (lobbyIdInput)
                lobbyIdInput.value = lobbyID; // convenience
            break;
        }
        case "lobby_joined": {
            // payload may have { lobby_id }
            lobbyID = ((_b = msg.payload) === null || _b === void 0 ? void 0 : _b.lobby_id) || lobbyID;
            isLeader = false;
            hide(startBtn);
            setText(messageBox, `Joined lobby ${lobbyID}. Waiting for leader to start…`);
            break;
        }
        case "join_failed": {
            // payload: { error: string }
            setText(messageBox, `Join failed: ${payloadToString(((_c = msg.payload) === null || _c === void 0 ? void 0 : _c.error) || msg.payload)}`);
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
            const r = msg.payload;
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
