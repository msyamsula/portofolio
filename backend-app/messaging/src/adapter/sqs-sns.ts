import { SNS } from "@aws-sdk/client-sns";
import { SQS } from "@aws-sdk/client-sqs";
import { createAdapter, PubSubAdapter, type AdapterOptions } from "@socket.io/aws-sqs-adapter";
import type { ClusterAdapterOptions } from "socket.io-adapter";

type adapter = (nsp: any) => PubSubAdapter;

export function newSqsSnsAdapter(opts: AdapterOptions & ClusterAdapterOptions) : adapter {
    const snsClient = new SNS()
    const sqsClient = new SQS()
    return createAdapter(snsClient, sqsClient, opts);
}

