import { Collection, MongoClient } from "mongodb";
import { createAdapter, MongoAdapter } from "@socket.io/mongo-adapter";

type adapter = (nsp: any) => MongoAdapter;

export async function getMonggoCollectionAdapter(db: string, collection: string, uri: string): Promise<Collection> {

    const mongoClient = new MongoClient(uri);

    await mongoClient.connect();

    try {
        await mongoClient.db(db).createCollection(collection, {
            capped: true,
            size: 1e6
        });
    } catch (e) {
        // collection already exists
    }
    return mongoClient.db(db).collection(collection);
}

export function newMongoAdapter(collection: Collection): adapter {
    return createAdapter(collection);
}







