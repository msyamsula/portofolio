import http from "http";
import { Server, Namespace, type ServerOptions } from "socket.io";
import { Events } from "./event.js";
import express from "express";

class messagePayload {
    senderId: string;
    receiverId: string;
    content: string;
    timestamp: string;

    constructor(senderId: string, receiverId: string, content: string, timestamp: string) {
        this.senderId = senderId;
        this.receiverId = receiverId;
        this.content = content;
        this.timestamp = timestamp;
    }

}

interface listenEvent {
    SEND: (msg: messagePayload) => void;
}
interface emitEvent extends listenEvent {
    SEND: (msg: messagePayload) => void;
}
interface serverSideEmit extends emitEvent { }


export class websocket {
    io: Server<listenEvent, emitEvent, serverSideEmit>;
    port: string;
    server: http.Server;

    constructor(port: string = "3000", opt: Partial<ServerOptions>) {
        this.port = port

        const app = express();
        this.server = http.createServer(app);
        this.io = new Server(this.server, opt);
    }

    #start() {
        this.io.on(Events.Connection, (socket) => {
            let userId = socket.handshake.query.userId
            console.log("client connected:", socket.id);
            
            socket.join(userId!) // join the room
            console.log("client join rooms:", userId);


            socket.on(Events.Send, (msg: messagePayload) => {
                console.log("message from client:", msg);
                let receiver = msg.receiverId
                this.io.in(receiver).emit(Events.Send, msg) // room broadcast
                // this.io.emit(Events.Send, msg); // broadcast to all clients
                // if (ack){
                //     ack() // ack to sender
                // }
            });

            socket.on(Events.Disconnect, () => {
                console.log("client disconnected:", socket.id);
            });
        });
    }

    run() {


        this.#start(); // run the websocket

        this.server.listen(this.port, () => {
            console.log(`Server running on http://localhost:${this.port}`);
        }); // run http server
    }

}



