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

function handleChat(payload) {
    // const messageText = messageInput.value.trim();
    var text = payload.message
    if (!text) {
        return
    }
    var sender = payload.sender
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

function receiveMessage(event) {
    console.log("Message from server: ", event.data);
    // You can also handle the incoming message here
    data = JSON.parse(event.data)
    if (data.type == "chat") {
        handleChat(data.payload)
    }
}
function connect() {
    if (socket) {
        socket.close()
    }

    if (roomName == "" || userName == "") {
        alert("please enter roomName and userName before connect")
        return
    }

    url = `${host}/chat/ws/${roomName}?username=${userName}`
    socket = new WebSocket(url);

    // When the connection is established
    socket.addEventListener("open", function (event) {
        console.log("Connected to WebSocket server!");
        // Send a message to the server
        // socket.send("Hello, Server!");
    });

    // When a message is received from the server
    socket.addEventListener("message", receiveMessage);

    // When the connection is closed
    socket.addEventListener("close", function (event) {
        // alert(`Disconnected from WebSocket server! (${event.reason})`);
        console.log("reconnecting");
        socket = new WebSocket(url)
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
            type: "chat",
            payload: {
                message: messageText,
                sender: userName,
            }
        })
        socket.send(data)
        const userMessageHTML = `
            <div class="message user">
                <span>${messageText}</span>
                <span class="tick-button">${noTick}</span>
            </div>
        `;

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