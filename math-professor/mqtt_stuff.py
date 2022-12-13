import boto3
import json

def publish(topic, payload):
    client = boto3.client('iot-data', region_name='eu-central-1')
    client.publish(
            topic=topic,
            qos=0,
            payload=json.dumps(payload)
        )
