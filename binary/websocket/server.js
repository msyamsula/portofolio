import { createServer } from "http";
import { Server } from "socket.io";

const httpServer = createServer();
const io = new Server(httpServer, {
  cors: {
    origin: "*",
    allowedHeaders: '*',
  }
});

import { Writer } from "nsqjs"
import { getNsqd } from "./util.js"

// const nsqd = "0.0.0.0"
// const nsqd = "0.0.0.0"
// const w = new Writer('127.0.0.1', 4150)
const topicSendMessage = "send_message"
const topicReadMessage = "read_message"
const topicUpdateUnread = "update_unread"
let nsqdAddress = ""
async function main() {
  let data = await getNsqd()
  if (data.producers.length > 0) {
    nsqdAddress = data.producers[0].hostname
  } else {
    console.log("there are no nsqd running");
    return
  }

  const w = new Writer(nsqdAddress, 4150)
  w.connect()

  w.on('ready', async () => {

    console.log("ready", nsqdAddress);

    io.on("connection", (socket) => {
      

      socket.on("chat", (msg) => {
        // socket.to(room).emit("chat", msg)
        let receiverEvent = msg.receiverId
        let senderEvent = msg.senderId
        socket.broadcast.emit(receiverEvent, msg)
        socket.emit(senderEvent, {
          subevent: "delivered",
          senderId: msg.senderId,
          receiverId: msg.receiverId,
        })
        w.publish(topicSendMessage, msg, err => {
          console.log(err);
        })
      })

      socket.on("read", (msg) => {
        w.publish(topicReadMessage, msg, err => {
          console.log(err);
        })

        socket.broadcast.emit(msg.senderId, {subevent: "read", senderId: msg.senderId, receiverId: msg.receiverId})
      })

      socket.on("userLogin", msg => {
        socket.broadcast.emit("userLogin", msg)
      })

      socket.on("userLogout", msg => {
        socket.broadcast.emit("userLogout", msg)
      })

      socket.on("updateUnread", msg => {
        w.publish(topicUpdateUnread, msg)
      })

      socket.on("addFriend", msg => {
        msg.subevent = "addFriend"
        socket.broadcast.emit(msg.receiverId, msg)
      })


    });

    httpServer.listen(8080, "0.0.0.0");


  })

  w.on("error", err => {
    console.log(err, "error");
  })

  w.on('closed', async () => {
    console.log('Writer closed')
    setTimeout(main, 5000)
  })


}

main()



