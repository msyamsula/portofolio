import http from "http";
import { Server, Namespace, type ServerOptions } from "socket.io";
import { Events } from "./event.js";
import express from "express";
import { SnsPublisher } from "../sns/publisher.js"

class messagePayload {
    id: string;
    senderId: number;
    receiverId: number;
    conversationId: string;
    data: string;
    createTime?: string;
    event?: string;

    constructor(id: string, senderId: number, receiverId: number, conversationId: string, data: string, createTime: string) {
        this.id = id
        this.senderId = senderId
        this.receiverId = receiverId
        this.conversationId = conversationId
        this.data = data
        this.createTime = createTime
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
    snsPublisher: SnsPublisher;

    constructor(port: string = "3000", publisher: SnsPublisher, opt: Partial<ServerOptions>) {
        this.port = port

        const app = express();
        this.server = http.createServer(app);
        this.io = new Server(this.server, opt);
        this.snsPublisher = publisher
    }

    #start() {
        this.io.on(Events.Connection, (socket) => {
            let userId = socket.handshake.query.userId

            socket.join(userId!) // join the room


            socket.on(Events.Send, (msg: messagePayload) => {
                let receiver = `${msg.receiverId}`
                this.io.in(receiver).emit(Events.Send, msg) // room broadcast

                // publish to sns to persist the data
                msg.event = Events.Send
                this.snsPublisher.publish(JSON.stringify(msg))
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



