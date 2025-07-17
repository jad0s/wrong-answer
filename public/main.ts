const ws = new WebSocket("ws://192.168.0.155:8080/ws");

interface Message {
  type: string;
  payload: any;
}

let username = "";

const usernameForm = document.getElementById("username-form") as HTMLFormElement;
const usernameInput = document.getElementById("username") as HTMLInputElement;

const gameDiv = document.getElementById("game")!;
const questionDiv = document.getElementById("question")!;
const answerForm = document.getElementById("answer-form") as HTMLFormElement;
const answerInput = document.getElementById("answer") as HTMLInputElement;

const voteForm = document.getElementById("vote-form") as HTMLFormElement;
const voteInput = document.getElementById("vote") as HTMLInputElement;

const messageBox = document.getElementById("message-box")!;
const answersDiv = document.getElementById("answers")!;
const resultDiv = document.getElementById("result")!;

usernameForm.addEventListener("submit", (e) => {
  e.preventDefault();
  username = usernameInput.value.trim();
  if (!username) return;
  const msg: Message = { type: "set_username", payload: username };
  ws.send(JSON.stringify(msg));
  usernameForm.classList.add("hidden");
  gameDiv.classList.remove("hidden");
});

answerForm.addEventListener("submit", (e) => {
  e.preventDefault();
  const answer = answerInput.value.trim();
  if (!answer) return;
  const msg: Message = { type: "submit_answer", payload: answer };
  ws.send(JSON.stringify(msg));
  answerForm.classList.add("hidden");
});

voteForm.addEventListener("submit", (e) => {
  e.preventDefault();
  const vote = voteInput.value.trim();
  if (!vote) return;
  const msg: Message = { type: "vote", payload: vote };
  ws.send(JSON.stringify(msg));
  voteForm.classList.add("hidden");
});

ws.onmessage = (e) => {
  let msg: Message;

  try {
    msg = JSON.parse(e.data);
  } catch (err) {
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
      const reveal = msg.payload as {
        impostor: string;
        most_voted: string;
        impostor_question: string;
      };
      resultDiv.innerHTML =
        `<strong>Voting Result:</strong><br>` +
        (reveal.impostor === reveal.most_voted
          ? `You guessed correctly! The impostor was <b>${reveal.impostor}</b>.`
          : `Wrong guess! You voted <b>${reveal.most_voted}</b>, but the impostor was <b>${reveal.impostor}</b>.`) +
        `<br><br>Impostor's question was: ${reveal.impostor_question}`;
      break;

    case "reveal_normal_question":
      messageBox.textContent = `The impostor saw a different question. The real one was: "${msg.payload}"`;
      messageBox.classList.remove("hidden");
      break;

    default:
      console.log("Unknown message:", msg);
  }
};
