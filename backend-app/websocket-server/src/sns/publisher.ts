import { SNSClient, PublishCommand } from "@aws-sdk/client-sns";


export class SnsPublisher {
    private client: SNSClient;
    private topicArn: string;
    constructor(region: string, topicArn: string) {
        this.client = new SNSClient({ region: region });
        this.topicArn = topicArn
    }

    async publish(message: string) {
        const input = {
            TopicArn: this.topicArn,
            Message: message,
        };

        await this.client.send(new PublishCommand(input));
    }
}


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

