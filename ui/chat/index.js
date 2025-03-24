async function login(username) {
    let user = await getUser(username)
    if (user) {
        return user
    }

    user = await registerUser(username)
    if (user) {
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