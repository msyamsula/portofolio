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
let nsqdAddress = ""
async function main() {
  let data = await getNsqd()
  // console.log(data.producers.length);
  if (data.producers.length > 0) {
    nsqdAddress = data.producers[0].hostname
  } else {
    console.log("there are no nsqd running");
    return
  }

  // console.log(nsqdAddress);
  const w = new Writer(nsqdAddress, 4150)
  w.connect()

  w.on('ready', async () => {

    console.log("ready", nsqdAddress);

    io.on("connection", (socket) => {
      // ...
      // console.log(socket.handshake.headers);
      //   console.log(socket.handshake.query);
      //   console.log(socket.handshake.url);

      // let userId = socket.handshake.headers["userid"]
      // let pairId = socket.handshake.headers["pairid"]

      // let room = createRoomName(userId, pairId)

      // console.log(room);
      // socket.join(room)

      socket.on("chat", (msg) => {
        // socket.to(room).emit("chat", msg)
        let receiverEvent = msg.receiverId
        socket.broadcast.emit(receiverEvent, msg)
        w.publish(topicSendMessage, msg, err => {
          console.log(err);
        })
      })


    });

    httpServer.listen(8080, "0.0.0.0");

    // console.log(err);
    // w.publish('sample_topic', 'it really tied the room together')
    // w.deferPublish('sample_topic', ['This message gonna arrive 1 sec later.'], 1000)
    // w.publish('sample_topic', [
    //     'Uh, excuse me. Mark it zero. Next frame.',
    //     'Smokey, this is not \'Nam. This is bowling. There are rules.'
    // ])
    // w.publish('sample_topic', 'Wu?', err => {
    //     if (err) { return console.error(err.message) }
    //     console.log('Message sent successfully')
    //     w.close()
    // })
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



