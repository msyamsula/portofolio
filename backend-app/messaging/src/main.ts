import { websocket } from "./websocket/websocket.js"
import express from "express";
import http from "http";
import { Server } from "socket.io";

const PORT = process.env.PORT || "3000";

function main() {
    const app = express();
    const server = http.createServer(app);

    const io = new Server(server, {
        cors: {
            origin: "*",
        },
    });

    let ws = new websocket(io, server, PORT);
    ws.run();
}

main();