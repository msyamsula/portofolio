import { websocket } from "./websocket/websocket.js"
import type { ServerOptions } from "socket.io";
import { getMonggoCollectionAdapter, newMongoAdapter } from "./adapter/mongo.js";
import { newSqsSnsAdapter } from "./adapter/sqs-sns.js";
import { SnsPublisher } from "./sns/publisher.js";

const ENVIRONMENT = process.env.ENVIRONMENT
const PORT = process.env.PORT || "12000";
const MONGO_ADAPTER_DB = process.env.MONGO_DB || "my-mongodb"
const MONGO_ADAPTER_COLLECTION = process.env.MONGO_COLLECTION || "socket-io-adapter"
const MONGO_ADAPTER_URI = process.env.MONGO_URI || "mongodb://localhost:27017"
const AWS_ACCESS_KEY_ID = process.env.AWS_ACCESS_KEY_ID || ""
const AWS_SECRET_ACCESS_KEY = process.env.AWS_SECRET_ACCESS_KEY || ""
const AWS_REGION = process.env.AWS_REGION || ""
const SOCKET_SNS_TOPIC = process.env.SOCKET_SNS_TOPIC || ""
const SQS_PREFIX_QUEUE = process.env.SQS_PREFIX_QUEUE || ""
const ADAPTER_TYPE = process.env.ADAPTER_TYPE || ""
const PERSISTENCE_SNS_TOPIC_ARN = process.env.PERSISTENCE_SNS_TOPIC_ARN || ""

function showEnv() {
    console.log("ENVIRONMENT:", ENVIRONMENT);
    console.log("PORT:", PORT);
    console.log("MONGO_ADAPTER_DB:", MONGO_ADAPTER_DB);
    console.log("MONGO_ADAPTER_COLLECTION:", MONGO_ADAPTER_COLLECTION);
    console.log("MONGO_ADAPTER_URI:", MONGO_ADAPTER_URI);
    console.log("AWS_SECRET_ACCESS_KEY:", AWS_SECRET_ACCESS_KEY);
    console.log("AWS_ACCESS_KEY_ID:", AWS_ACCESS_KEY_ID);
    console.log("AWS_REGION:", AWS_REGION);
    console.log("SQS_PREFIX_QUEUE:", SQS_PREFIX_QUEUE);
    console.log("SOCKET_SNS_TOPIC:", SOCKET_SNS_TOPIC);
    console.log("ADAPTER_TYPE:", ADAPTER_TYPE);
    console.log("PERSISTENCE_SNS_TOPIC_ARN:", PERSISTENCE_SNS_TOPIC_ARN);

}

async function getAdapter(cfg: Partial<ServerOptions>): Promise<Partial<ServerOptions>> {
    switch (ADAPTER_TYPE) {
        case "mongo":
            let collection = await getMonggoCollectionAdapter(MONGO_ADAPTER_DB, MONGO_ADAPTER_COLLECTION, MONGO_ADAPTER_URI)
            cfg.adapter = newMongoAdapter(collection, {})
            break;
        case "sqs":
            cfg.adapter = newSqsSnsAdapter({
                topicName: SOCKET_SNS_TOPIC,
                queuePrefix: SQS_PREFIX_QUEUE,
            })
            break;

        default:
            break;
    }

    return cfg
}

async function main() {
    if (ENVIRONMENT != "production") {
        showEnv()
    }

    let serverConfig: Partial<ServerOptions> = {
        cors: {
            origin: "*",
        },
        connectionStateRecovery: {
            maxDisconnectionDuration: 2 * 60 * 1000,
            skipMiddlewares: true
        },
        transports: ["websocket"],
    }
    serverConfig = await getAdapter(serverConfig) // select adapter
    console.log("server adapter", serverConfig.adapter);
    let publisher = new SnsPublisher(AWS_REGION, PERSISTENCE_SNS_TOPIC_ARN)
    

    let ws = new websocket(PORT, publisher, serverConfig);

    ws.run();
}

await main();