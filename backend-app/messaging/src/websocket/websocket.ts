import http from "http";
import { Server } from "socket.io";
import { Events } from "./event.js";

export class websocket {
    io: Server;
    server: http.Server;
    port: string;

    constructor(io: Server, server: http.Server, port: string = "3000") {
        this.io = io;
        this.server = server;
        this.port = port
    }

    #start() {
        this.io.on(Events.Connection, (socket) => {
            console.log("client connected:", socket.id);


            socket.on(Events.Send, (msg: string) => {
                console.log("message from client:", msg);
                socket.broadcast.emit(Events.Send, msg); // broadcast to all clients
            });

            socket.on(Events.Disconnect, () => {
                console.log("client disconnected:", socket.id);
            });
        });
    }

    run() {
        this.#start();
        this.server.listen(3000, () => {
            console.log(`Server running on http://localhost:${this.port}`);
        });
    }

}



