const websocketHost = "wss://websocket.syamsul.online"
// const websocketHost = "ws://0.0.0.0:8080"
function getPairId() {
    return parseInt(localStorage.getItem("pairId"))
}

function getUsername() {
    return localStorage.getItem("username")
}

function getUserId() {
    return parseInt(localStorage.getItem("id"))
}
let socket = io(websocketHost);