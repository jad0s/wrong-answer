// main.ts
var ws = new WebSocket("wss://localhost:443/ws");
var username = "";
var lobbyID = "";
var isLeader = false;
// UI elements
var usernameForm = document.getElementById("username-form");
var usernameInput = document.getElementById("username");
var lobbyControls = document.getElementById("lobby-controls");
var createBtn = document.getElementById("btn-create");
var joinBtn = document.getElementById("btn-join");
var lobbyIdInput = document.getElementById("lobby-id");
var startBtn = document.getElementById("btn-start");
var gameDiv = document.getElementById("game");
var questionDiv = document.getElementById("question");
var answerForm = document.getElementById("answer-form");
var answerInput = document.getElementById("answer");
var voteForm = document.getElementById("vote-form");
var voteInput = document.getElementById("vote");
var messageBox = document.getElementById("message-box");
var answersDiv = document.getElementById("answers");
var resultDiv = document.getElementById("result");
// utility
function send(type, payload) {
    ws.send(JSON.stringify({ type: type, payload: payload }));
}
// after login
usernameForm.addEventListener("submit", function (e) {
    e.preventDefault();
    username = usernameInput.value.trim();
    if (!username)
        return;
    usernameForm.classList.add("hidden");
    lobbyControls.classList.remove("hidden");
});
// create lobby
createBtn.addEventListener("click", function () {
    send("create_lobby", { username: username });
});
// join lobby
joinBtn.addEventListener("click", function () {
    var id = lobbyIdInput.value.trim().toUpperCase();
    if (!id)
        return;
    send("join_lobby", { username: username, lobby_id: id });
});
// start game (only leader)
startBtn.addEventListener("click", function () {
    send("start_game", {});
});
// answer submission
answerForm.addEventListener("submit", function (e) {
    e.preventDefault();
    var ans = answerInput.value.trim();
    if (!ans)
        return;
    send("submit_answer", ans);
    answerForm.classList.add("hidden");
});
// vote submission
voteForm.addEventListener("submit", function (e) {
    e.preventDefault();
    var v = voteInput.value.trim();
    if (!v)
        return;
    send("vote", v);
    voteForm.classList.add("hidden");
});
// incoming messages
ws.onmessage = function (e) {
    var _a;
    var msg;
    try {
        msg = JSON.parse(e.data);
    }
    catch (_b) {
        console.error("Invalid JSON:", e.data);
        return;
    }
    switch (msg.type) {
        case "lobby_created":
            lobbyID = msg.payload.lobby_id;
            isLeader = true;
            messageBox.textContent = "Lobby ".concat(lobbyID, " created. Waiting for players\u2026");
            startBtn.classList.remove("hidden");
            break;
        case "lobby_joined":
            lobbyID = ((_a = msg.payload) === null || _a === void 0 ? void 0 : _a.lobby_id) || lobbyID;
            isLeader = false;
            messageBox.textContent = "Joined lobby ".concat(lobbyID, ". Waiting for leader to start\u2026");
            startBtn.classList.add("hidden");
            break;
        case "join_failed":
            messageBox.textContent = "Join failed: ".concat(msg.payload.error);
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
            var r = msg.payload;
            resultDiv.innerHTML =
                "<strong>Voting Result:</strong><br>" +
                    (r.impostor === r.most_voted
                        ? "You guessed correctly! The impostor was <b>".concat(r.impostor, "</b>.")
                        : "Wrong guess! You voted <b>".concat(r.most_voted, "</b>, but the impostor was <b>").concat(r.impostor, "</b>.")) +
                    "<br><br>Impostor's question was: ".concat(r.impostor_question);
            break;
        case "reveal_normal_question":
            messageBox.textContent =
                "The impostor saw a different question. The real one was: \"".concat(msg.payload, "\"");
            messageBox.classList.remove("hidden");
            break;
        default:
            console.log("Unknown message:", msg);
    }
};
