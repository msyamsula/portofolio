import { websocket } from "./websocket/websocket.js"
import { newSqsSnsAdapter } from "./adapter/sqs-sns.js";
import { newMongoAdapter, getMonggoCollectionAdapter } from "./adapter/mongo.js";

const ENVIRONMENT = process.env.ENVIRONMENT
const PORT = process.env.PORT || "3000";
const MONGO_ADAPTER_DB = process.env.MONGO_DB || "my-mongodb"
const MONGO_ADAPTER_COLLECTION = process.env.MONGO_COLLECTION || "socket-io-adapter"
const MONGO_ADAPTER_URI = process.env.MONGO_URI || "mongodb://localhost:27017"
const AWS_ACCESS_KEY_ID = process.env.AWS_ACCESS_KEY_ID || ""
const AWS_SECRET_ACCESS_KEY = process.env.AWS_SECRET_ACCESS_KEY || ""
const AWS_REGION = process.env.AWS_REGION || ""
const SNS_TOPIC = process.env.SNS_TOPIC || ""
const SQS_PREFIX_QUEUE = process.env.SQS_PREFIX_QUEUE || ""

function showEnv() {
    console.log(ENVIRONMENT);
    console.log(PORT);
    console.log(MONGO_ADAPTER_DB)
    console.log(MONGO_ADAPTER_COLLECTION)
    console.log(MONGO_ADAPTER_URI)
    console.log(AWS_SECRET_ACCESS_KEY);
    console.log(AWS_ACCESS_KEY_ID);
    console.log(AWS_REGION);
    console.log(SQS_PREFIX_QUEUE);
    console.log(SNS_TOPIC);
    
}


async function main() {
    if (ENVIRONMENT != "production") {
        showEnv()
    }


    let adapter = newSqsSnsAdapter({
        topicName: SNS_TOPIC,
        queuePrefix: SQS_PREFIX_QUEUE,
    });

    // const mongoCollectionAdapter = await getMonggoCollectionAdapter(MONGO_ADAPTER_DB, MONGO_ADAPTER_COLLECTION, MONGO_ADAPTER_URI);
    // let adapter = newMongoAdapter(mongoCollectionAdapter);


    let ws = new websocket(PORT, {
        cors: {
            origin: "*",
        },
        connectionStateRecovery: {
            maxDisconnectionDuration: 2 * 60 * 1000,
            skipMiddlewares: true
        },
        transports: ["websocket"],
        adapter: adapter,
    });

    ws.run();
}

await main();