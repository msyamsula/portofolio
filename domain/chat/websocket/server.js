import { createServer } from "http";
import { Server } from "socket.io";

const httpServer = createServer();
const io = new Server(httpServer, { /* options */ });

io.on("connection", async (socket) => {
  // ...
//   console.log(socket.handshake.headers);
//   console.log(socket.handshake.query);
//   console.log(socket.handshake.url);

  let room = socket.handshake.headers["room"]
  console.log(room);
  socket.join(room)

  socket.on("chat", (msg)=>{
    io.in(room).emit(msg)
  })

  socket.broadcast.emit("chat", room)
  socket.emit("chat", "hello world")

});

httpServer.listen(8080, "0.0.0.0");