// const host = "http://0.0.0.0:8000"
const host = "https://api.syamsul.online"

async function registerUser(username) {
    try {
        let response = await fetch(`${host}/user`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                username: username,
            })
        })

        let user = await response.json()
        if (user.data) {
            return user.data
        }

        return null
    } catch (error) {
        console.log(error, "goes here");
        return null
    }
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
    user = await response.json()
    console.log(user);
    if (user.data) {
        return user.data
    }

    return null

}

async function login(username) {
    let user = await getUser(username)
    if (user.username) {
        return user
    }

    user = await registerUser(username)
    if (user.username) {
        return user
    }

    return null

}

document.getElementById("login-form").addEventListener("submit", async function (event) {
    event.preventDefault();
    const username = document.getElementById("username").value;

    user = await login(username)
    console.log(user);
    if (user) {
        localStorage.setItem("username", user.username);
        localStorage.setItem("id", user.id);
        window.location.href = "chat.html"; // Redirect to chat page
        return
    }

    alert("try different name")
});

window.onload = function () {
    const username = localStorage.getItem("username");
    if (username) {
        // If username exists in localStorage, redirect to the chat page
        window.location.href = "chat.html";  // Redirect to chat page
    }
};