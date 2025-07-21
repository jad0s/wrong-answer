var ws = new WebSocket("ws://192.168.0.155:8080/ws");
var username = "";
var usernameForm = document.getElementById("username-form");
var usernameInput = document.getElementById("username");
var gameDiv = document.getElementById("game");
var questionDiv = document.getElementById("question");
var answerForm = document.getElementById("answer-form");
var answerInput = document.getElementById("answer");
var voteForm = document.getElementById("vote-form");
var voteInput = document.getElementById("vote");
var messageBox = document.getElementById("message-box");
var answersDiv = document.getElementById("answers");
var resultDiv = document.getElementById("result");
usernameForm.addEventListener("submit", function (e) {
    e.preventDefault();
    username = usernameInput.value.trim();
    if (!username)
        return;
    var msg = { type: "set_username", payload: username };
    ws.send(JSON.stringify(msg));
    usernameForm.classList.add("hidden");
    gameDiv.classList.remove("hidden");
});
answerForm.addEventListener("submit", function (e) {
    e.preventDefault();
    var answer = answerInput.value.trim();
    if (!answer)
        return;
    var msg = { type: "submit_answer", payload: answer };
    ws.send(JSON.stringify(msg));
    answerForm.classList.add("hidden");
});
voteForm.addEventListener("submit", function (e) {
    e.preventDefault();
    var vote = voteInput.value.trim();
    if (!vote)
        return;
    var msg = { type: "vote", payload: vote };
    ws.send(JSON.stringify(msg));
    voteForm.classList.add("hidden");
});
ws.onmessage = function (e) {
    var msg;
    try {
        msg = JSON.parse(e.data);
    }
    catch (err) {
        console.error("Failed to parse message:", e.data);
        return;
    }
    switch (msg.type) {
        case "question":
            questionDiv.textContent = msg.payload;
            answerForm.classList.remove("hidden");
            break;
        case "reveal_answers":
            answersDiv.textContent = msg.payload;
            break;
        case "vote":
            messageBox.textContent = "Voting started! Type the name of the impostor.";
            messageBox.classList.remove("hidden");
            voteForm.classList.remove("hidden");
            break;
        case "reveal_impostor":
            var reveal = msg.payload;
            resultDiv.innerHTML =
                "<strong>Voting Result:</strong><br>" +
                    (reveal.impostor === reveal.most_voted
                        ? "You guessed correctly! The impostor was <b>".concat(reveal.impostor, "</b>.")
                        : "Wrong guess! You voted <b>".concat(reveal.most_voted, "</b>, but the impostor was <b>").concat(reveal.impostor, "</b>.")) +
                    "<br><br>Impostor's question was: ".concat(reveal.impostor_question);
            break;
        case "reveal_normal_question":
            messageBox.textContent = "The impostor saw a different question. The real one was: \"".concat(msg.payload, "\"");
            messageBox.classList.remove("hidden");
            break;
        default:
            console.log("Unknown message:", msg);
    }
};
