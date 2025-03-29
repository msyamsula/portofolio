async function login(username) {
    user = await registerUser(username, true)
    if (!user.error) {
        return user
    }

    return null

}

document.getElementById("login-form").addEventListener("submit", async function (event) {
    event.preventDefault();
    const username = document.getElementById("username").value;

    user = await login(username)
    if (user) {
        localStorage.setItem("username", user.username);
        localStorage.setItem("id", user.id);
        let msg = {
            userId: getUserId()
        }
        socket.emit("userLogin", msg)
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