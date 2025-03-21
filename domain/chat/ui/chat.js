let friends = [
    {
        id: 1,
        username: "mantap"
    },
    {
        id: 2,
        username: "wow"
    },
    {
        id: 3,
        username: "ok"
    },
]

function exitChat() {
    localStorage.clear()
    window.location.href = "index.html"; // Redirect to chat page
}

// Function to search for messages
function searchMessages() {
    let input = document.getElementById('message-search-input');
    let filter = input.value.toLowerCase();
    let messages = document.getElementsByClassName('message');

    for (let i = 0; i < messages.length; i++) {
        let message = messages[i].textContent || messages[i].innerText;
        if (message.toLowerCase().includes(filter)) {
            messages[i].style.display = "";
        } else {
            messages[i].style.display = "none";
        }
    }
}

// Allow sending message by pressing Enter
const messageInput = document.getElementById("message-input");
messageInput.addEventListener("keypress", function (event) {
    if (event.key === "Enter") {
        sendMessage();
    }
});

// Function to send message
function sendMessage() {
    let messageInput = document.getElementById("message-input");
    let message = messageInput.value;
    if (message) {
        let chatBox = document.getElementById("chat-box");
        let newMessage = document.createElement("div");
        newMessage.className = "message sent"; // Initially mark the message as sent
        newMessage.textContent = message;

        // Add tick mark for the sent message
        let tickSent = document.createElement("span");
        tickSent.className = "status tick-sent";
        tickSent.textContent = "✓";
        newMessage.appendChild(tickSent);

        chatBox.appendChild(newMessage);
        messageInput.value = ""; // Clear the input field
        chatBox.scrollTop = chatBox.scrollHeight; // Scroll to the bottom

        messageInput.focus()
        // // Simulate message delivery and reading after a delay
        // setTimeout(() => {
        //     tickSent.className = "status tick-delivered";
        //     tickSent.textContent = "✓✓"; // Change to delivered
        // }, 2000);

        // setTimeout(() => {
        //     tickSent.className = "status tick-read";
        //     tickSent.textContent = "✓✓"; // Change to read
        // }, 4000);
    }
}

function searchFriends() {
    alert("Search friends functionality here!");
}

function switchUser(user) {
    document.getElementById("chat-header").textContent = user.username;
    localStorage.setItem("pairId", user.id)
    localStorage.setItem("pairUsername", user.username)
    // console.log(localStorage.getItem("pairId"));
    // console.log(localStorage.getItem("pairUsername"));
}

function refreshFriendList(){
    let friendTab = document.getElementById("users")
    friendTab.innerHTML = ""
    for (let friend of friends) {
        let f = document.createElement("li")
        f.innerHTML = friend.username
        f.addEventListener("click", () => switchUser(friend))
        friendTab.appendChild(f)
    }
}

window.onload = function () {
    messageInput.focus()
    document.getElementById("friendModal").style.display = "none";
    const username = localStorage.getItem("username");
    if (!username) {
        // If username exists in localStorage, redirect to the chat page
        window.location.href = "index.html";  // Redirect to chat page
        return
    }

    refreshFriendList()
    
};


function showFriendModal(){
    document.getElementById("friendModal").style.display = "flex";
    document.getElementById("friendName").value = "";
    document.getElementById("friendName").focus();
}

document.getElementById("friendModal").addEventListener("keypress", function (event) {
    let displayStatus = document.getElementById("friendModal").style.display
    if (event.key == "Enter" && displayStatus != "none"){
        submitFriend()
    }
})

// Example functions (addFriend, searchFriends, switchUser)
function addFriend(friend) {

    friends.push({
        id: friends.length+1,
        username: friend.username,
    })
    refreshFriendList()
    // alert("Add friend functionality here!");
}

// Close modal when clicking the close button
document.querySelector(".close").addEventListener("click", function() {
    document.getElementById("friendModal").style.display = "none";
});

// Close modal when clicking outside the modal
window.addEventListener("click", function(event) {
    let modal = document.getElementById("friendModal");
    if (event.target === modal) {
        modal.style.display = "none";
    }
});

// Function to handle adding a friend (modify as needed)
function submitFriend() {
    let friendName = document.getElementById("friendName").value;
    console.log(friendName);
    if (friendName) {
        document.getElementById("friendModal").style.display = "none";
        let user = {
            username: friendName,
            id: 0,
        }
        addFriend(user)
        switchUser(user)
    }
}