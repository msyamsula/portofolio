// const host = "http://0.0.0.0:8000"
const host = "https://api.syamsul.online"
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

function createPi(p) {
    let n = p.length
    let i = 0
    let pi = Array(n+1).fill(0)
    pi[0]=-1
    let j = -1
    while (i < n) {
        while(j<n && p[j] != p[i]) {
            j = pi[j]
        }
        i++
        j++
        p[i]=j
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

function switchUser(user) {
    document.getElementById("chat-header").textContent = user.username;
    localStorage.setItem("pairId", user.id)
    localStorage.setItem("pairUsername", user.username)
}

function refreshFriendList() {
    let friendTab = document.getElementById("users")
    friendTab.innerHTML = ""
    for (let friend of friendsOnDisplay) {
        let f = document.createElement("li")
        f.innerHTML = friend.username
        f.addEventListener("click", () => switchUser(friend))
        friendTab.appendChild(f)
    }
}

async function getFriends(id) {
    let response = await fetch(`${host}/user/friend?id=${id}`, {
        method: 'GET',
        headers: {
            'Content-Type': 'application/json'
        }
    })

    let user = await response.json()
    if (user.data) {
        return user.data
    }

    return null
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
        // If username exists in localStorage, redirect to the chat page
        window.location.href = "index.html";  // Redirect to chat page
        return
    }

    await populateFriends()

};


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
        id: friends.length + 1,
        username: friend.username,
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

async function connectWithUser(idA, idB) {
    idA = parseInt(idA, 10)
    let edge = {
        small_id: idA,
        big_id: idB
    }
    let response = await fetch(`${host}/user/friend`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify(edge)
    })

    let data = await response.json()
    if (data) {
        return data
    }
    return null
}

async function getUser(username) {
    const query = new URLSearchParams({
        username: username
    }).toString()
    response = await fetch(`${host}/user?${query}`, {
        method: 'GET',
        headers: {
            'Content-Type': 'application/json'
        }
    })
    let user = await response.json()
    if (user) {
        return user
    }

    return null

}

// Function to handle adding a friend (modify as needed)
async function submitFriend() {
    let friendName = document.getElementById("friendName").value;
    if (friendName) {
        document.getElementById("friendModal").style.display = "none";

        let newFriend = await getUser(friendName)
        if (!newFriend.error) {
            let r = await connectWithUser(id, newFriend.data.id)
            if (r != null && !r.error) {
                addFriend(newFriend.data)
                switchUser(newFriend.data)
                return
            } else {
                alert(r.error)
            }
        } else {
            alert(newFriend.error)
        }
    }
}