const messageContainer = document.getElementById("message-container");
const messageInput = document.getElementById("message-input");
const userNameBox = document.getElementById("user-name")
const room = document.getElementById("room-name")
const noTick = ""
const tick = "✔"
const doubleTick = "✔✔"
// Create a new WebSocket connection
// const host = "ws://0.0.0.0:8080"
const host = "wss://api.syamsul.online"
console.log(document.getElementById("room-name"));
var socket, url

function receiveMessage(event) {
    console.log("Message from server: ", event.data);
    // You can also handle the incoming message here
    data = JSON.parse(event.data)
    // const messageText = messageInput.value.trim();
    var text = data.text
    if (!text) {
        return
    }
    var sender = data.sender
    if (sender == userName || userName == "") {
        return
    }
    const messageHtml = `
            <div class="message other">
                <span>${text}</span>
                <span class="tick-button">${noTick}</span>
            </div>
        `;

    // Append both messages to the container
    messageContainer.innerHTML += messageHtml;

    // Scroll to the bottom
    messageContainer.scrollTop = messageContainer.scrollHeight;

    // Scroll to the bottom
    messageContainer.scrollTop = messageContainer.scrollHeight;
}
function connect() {
    if (socket) {
        socket.close()
    }

    if (roomName == "" || userName == "") {
        alert("please enter roomName and userName before connect")
        return
    }

    url = `${host}/chat/ws/${roomName}`
    socket = new WebSocket(url);

    // When the connection is established
    socket.addEventListener("open", function (event) {
        console.log("Connected to WebSocket server!");
        // Send a message to the server
        socket.send("Hello, Server!");
    });

    // When a message is received from the server
    socket.addEventListener("message", receiveMessage);

    // When the connection is closed
    socket.addEventListener("close", function (event) {
        console.log("Disconnected from WebSocket server!");
    });

    // When an error occurs
    socket.addEventListener("error", function (event) {
        console.error("WebSocket error: ", event);
    });
}

// alert("Welcome, please enter room and username before chatting with our community. Thanks")

console.log(url);

var userName = ""
var roomName = ""

room.addEventListener("change", function (event) {
    roomName = event.target.value
})

userNameBox.addEventListener("change", function (event) {
    userName = event.target.value
})

var myId = document.getElementById("user-name").innerHTML

function sendMessage() {
    if (userName == "" || roomName == "") {
        alert("please enter your username and roomName before sending the mesage")
        return
    }
    const messageText = messageInput.value.trim();
    if (messageText) {
        var data = JSON.stringify({
            text: messageText,
            sender: userName,
        })
        socket.send(data)
        const userMessageHTML = `
            <div class="message user">
                <span>${messageText}</span>
                <span class="tick-button">${noTick}</span>
            </div>
        `;

        // Create the other message with tick button using HTML literals
        // const otherMessageHTML = `
        //     <div class="message other">
        //         <span>${messageText}</span>
        //         <span class="tick-button">${tick}</span>
        //     </div>
        // `;

        // Append both messages to the container
        messageContainer.innerHTML += userMessageHTML;

        // Scroll to the bottom
        messageContainer.scrollTop = messageContainer.scrollHeight;

        // Scroll to the bottom
        messageContainer.scrollTop = messageContainer.scrollHeight;

        // Clear the input field
        messageInput.value = '';
    }
}

// Allow sending message by pressing Enter
messageInput.addEventListener("keypress", function (event) {
    if (event.key === "Enter") {
        sendMessage();
    }
});