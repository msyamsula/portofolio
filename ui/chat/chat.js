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
let friendsOnDisplay = []
let username = ""
let id = 0

let messageOnDisplay = []

async function exitChat() {
    await registerUser(getUsername(), false)
    let msg = {
        userId: getUserId(),
    }
    socket.emit("userLogout", msg)
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
        let pairId = getPairId()
        let senderId = getUserId()
        let msg = {
            sender_id: senderId,
            receiver_id: pairId,
            senderId: senderId,
            receiverId: pairId,
            text: message,
        }

        // emit to websocket
        socket.emit("chat", msg)


        messageOnDisplay.push(msg)
        let chatBox = document.getElementById("chat-box");
        let newMessage = document.createElement("div");
        newMessage.className = "message user"; // Initially mark the message as sent
        newMessage.textContent = message;


        chatBox.appendChild(newMessage);
        messageInput.value = ""; // Clear the input field
        chatBox.scrollTop = chatBox.scrollHeight; // Scroll to the bottom

        messageInput.focus()
    }
}

// Function to send message
function receiveMessage(msg) {
    // let messageInput = document.getElementById("message-input");
    let message = msg.text;
    let pairId = getPairId()
    if (message && pairId == msg.senderId) {
        let chatBox = document.getElementById("chat-box");
        let newMessage = document.createElement("div");
        newMessage.className = "message pair"; // Initially mark the message as sent
        newMessage.textContent = message;

        // Add tick mark for the sent message
        // let tickSent = document.createElement("span");
        // tickSent.className = "status tick-sent";
        // tickSent.textContent = "✓";
        // newMessage.appendChild(tickSent);

        msg.sender_id = msg.senderId
        msg.receiver_id = msg.receiverId
        messageOnDisplay.push(msg)
        chatBox.appendChild(newMessage);
        // messageInput.value = ""; // Clear the input field
        chatBox.scrollTop = chatBox.scrollHeight; // Scroll to the bottom

    }
}

function createPi(p) {
    let n = p.length
    let i = 0
    let pi = Array(n + 1).fill(0)
    pi[0] = -1
    let j = -1
    while (i < n) {
        while (j < n && p[j] != p[i]) {
            j = pi[j]
        }
        i++
        j++
        p[i] = j
    }

    return pi
}

function kmp(s, p) {
    let n = s.length
    let i = 0
    let j = 0
    let pi = createPi(p)
    while (i < n) {
        while (j >= 0 && s[i] != p[j]) {
            j = pi[j]
        }
        i++
        j++
        if (j == p.length) {
            return true
        }
    }

    return false
}

let typingTimer
const typingInterval = 500
function searchFriends() {
    clearTimeout(typingTimer)
    typingTimer = setTimeout(() => {
        let pattern = document.getElementById("friend-search-input").value
        friendsOnDisplay = []
        if (pattern) {
            for (let f of friends) {
                if (kmp(f.username, pattern)) {
                    friendsOnDisplay.push(f)
                }
            }
        } else {
            friendsOnDisplay = friends
        }
        refreshFriendList()
    }, typingInterval);
}

function refreshConversation() {
    let chatBox = document.getElementById("chat-box");
    chatBox.innerHTML = ""
    for (m of messageOnDisplay) {
        let newMessage = document.createElement("div");
        if (m.sender_id == id) {
            newMessage.className = "message user";
        } else {
            newMessage.className = "message pair";
        }
        newMessage.textContent = m.text;


        if (m.sender_id == getUserId()) {
            let tickSent = document.createElement("span");
            tickSent.className = "status tick-sent";
            if (!m.is_read) {
                tickSent.textContent = "✓";
            } else {
                tickSent.textContent = "✓✓";
            }
            newMessage.appendChild(tickSent);
        }

        chatBox.appendChild(newMessage);
    }


    messageInput.value = ""; // Clear the input field
    chatBox.scrollTop = chatBox.scrollHeight; // Scroll to the bottom

    messageInput.focus()
}

async function switchUser(user) {
    document.getElementById("chat-header").textContent = user.username;
    localStorage.setItem("pairId", user.id)
    localStorage.setItem("pairUsername", user.username)

    // send read event
    let event = {
        receiverId: getUserId(),
        senderId: getPairId(),
    }
    socket.emit("read", event)

    let conversation = await getConversation(id, user.id)
    if (conversation) {
        messageOnDisplay = conversation
    } else {
        messageOnDisplay = []
    }

    refreshConversation()

    // send update unread
    let msg = {
        senderId: getPairId(),
        receiverId: getUserId(),
        unread: 0,
    }
    updateFriendUnread(msg, msg.unread)
}

function refreshFriendList() {
    let friendTab = document.getElementById("users")
    friendTab.innerHTML = ""
    for (let friend of friendsOnDisplay) {
        let f = document.createElement("li")
        f.id = `${friend.id}`
        if (friend.online) {
            f.className = "online"
        } else {
            f.className = "offline"
        }
        f.innerHTML = friend.username
        if (friend.unread) {
            let unreadCount = `<span class="badge">${friend.unread}</span>`
            f.innerHTML += unreadCount
        }
        f.addEventListener("click", () => {
            f.innerHTML = friend.username
            switchUser({
                id: f.id,
                username: friend.username,
            })
        })
        friendTab.appendChild(f)
    }
}

async function populateFriends() {
    friends = await getFriends(id)
    if (!friends) {
        friends = []
    }
    friendsOnDisplay = friends
    refreshFriendList()
}

window.onload = async function () {
    messageInput.focus()
    document.getElementById("friendModal").style.display = "none";
    username = localStorage.getItem("username");
    id = localStorage.getItem("id")
    if (!username && !id) {
        // If username does not exists in localStorage, redirect to the index page
        window.location.href = "index.html";  // Redirect to index page
        return
    }

    await populateFriends()


    socket.on("connect", () => {
        console.log(socket.id);
    })

    socket.on(id, (msg) => {
        if (msg.subevent) {
            // check subevent, only refresh if msg.senderId == userId and msg.receiverId == pairId
            if (msg.subevent == "delivered") {
                if (msg.senderId == getUserId() && msg.receiverId == getPairId()) {
                    updateDelivered()
                }
            } else if (msg.subevent == "read") {
                if (msg.senderId == getUserId() && msg.receiverId == getPairId()) {
                    updateRead()
                }
            } else if (msg.subevent == "addFriend") {
                let newFriend = {
                    username: msg.username,
                    id: msg.senderId,
                    online: msg.online,
                    unread: msg.unread,
                }
                // friends.push(newFriend)
                friendsOnDisplay.push(newFriend)
                refreshFriendList()
            }
            return
        }

        // no sub event mean incoming message
        if (msg.senderId == getPairId() && msg.receiverId == getUserId()) {
            // this is a direct read, user is face to face
            let read = {
                senderId: msg.senderId,
                receiverId: msg.receiverId,
            }
            socket.emit("read", read)
        } else {
            let unread
            for (f of friends) {
                if (f.id == msg.senderId) {
                    unread = f.unread
                }
            }
            unread++
            updateFriendUnread(msg, unread)
        }
        receiveMessage(msg)
    })

    socket.on("userLogin", msg => {
        updateFriendOnlineStatus(true, msg.userId)
    })

    socket.on("userLogout", msg => {
        updateFriendOnlineStatus(false, msg.userId)
    })

};

function updateFriendUnread(msg, unread) {
    for (f of friends) {
        if (f.id == msg.senderId) {
            f.unread = unread
            unreadPlus = f.unread
        }
    }

    for (f of friendsOnDisplay) {
        if (f.id == msg.senderId) {
            f.unread = unread
        }
    }

    refreshFriendList()
    let unreadMsg = {
        senderId: msg.senderId,
        receiverId: getUserId(),
        unread: unread,
    }
    socket.emit("updateUnread", unreadMsg)
}

function updateFriendOnlineStatus(status, userId) {
    for (f of friends) {
        if (f.id == userId) {
            f.online = status
        }
    }

    for (f of friendsOnDisplay) {
        if (f.id == userId) {
            f.online = status
        }
    }

    refreshFriendList()
}

function updateDelivered() {
    let chatbox = document.getElementById("chat-box")

    for (let i = messageOnDisplay.length - 1; i >= 0; i--) {
        if (messageOnDisplay[i].is_read != undefined) {
            break
        }

        messageOnDisplay[i].is_read = false
        let newMessage = chatbox.children[i]
        if (messageOnDisplay[i].sender_id != getUserId()) {
            // do nothing for pair message
            continue
        }

        let tickSent = document.createElement("span");
        tickSent.className = "status tick-sent";
        tickSent.textContent = "✓";
        newMessage.appendChild(tickSent);

    }

}

function updateRead() {
    let chatbox = document.getElementById("chat-box")

    for (let i = messageOnDisplay.length - 1; i >= 0; i--) {
        if (messageOnDisplay[i].is_read) {
            break
        }

        messageOnDisplay[i].is_read = true
        let newMessage = chatbox.children[i]
        if (messageOnDisplay[i].sender_id != getUserId()) {
            // do nothing for pair message
            continue
        }

        newMessage.querySelector("span").textContent = "✓✓"

    }

}


function showFriendModal() {
    document.getElementById("friendModal").style.display = "flex";
    document.getElementById("friendName").value = "";
    document.getElementById("friendName").focus();
}

document.getElementById("friendModal").addEventListener("keypress", async function (event) {
    let displayStatus = document.getElementById("friendModal").style.display
    if (event.key == "Enter" && displayStatus != "none") {
        await submitFriend()
    }
})

// Example functions (addFriend, searchFriends, switchUser)
function addFriend(friend) {
    friends.push({
        id: friend.id,
        username: friend.username,
        online: friend.online
    })
    friendsOnDisplay = friends
    refreshFriendList()
    // alert("Add friend functionality here!");
}

// Close modal when clicking the close button
document.querySelector(".close").addEventListener("click", function () {
    document.getElementById("friendModal").style.display = "none";
});

// Close modal when clicking outside the modal
window.addEventListener("click", function (event) {
    let modal = document.getElementById("friendModal");
    if (event.target === modal) {
        modal.style.display = "none";
    }
});

// Function to handle adding a friend (modify as needed)
async function submitFriend() {
    let friendName = document.getElementById("friendName").value;
    if (friendName) {
        document.getElementById("friendModal").style.display = "none";

        let newFriend = await getUser(friendName)
        if (newFriend) {
            let r = await connectWithUser(id, newFriend.id)
            if (r != null && !r.error) {
                addFriend(newFriend)
                switchUser(newFriend)

                // notify other party when he is added as a friend
                let msg = {
                    senderId: getUserId(),
                    receiverId: getPairId(),
                    username: getUsername(),
                    online: true,
                    unread: 0,
                }
                socket.emit("addFriend", msg)
                return
            } else {
                alert("can't add user, please try again")
            }
        } else {
            alert("user not found")
        }
    }
}


