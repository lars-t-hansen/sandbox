# -*- fill-column: 100 -*-

# This is just an experiment to play around with AWS Lambda, AWS IoT, and Python, it doesn't amount
# to anything.

import math
import boto3
import mqtt_stuff

# An AWS IoT MQTT rule transforms this message
#
#    topic:   math/<function>/question
#    payload: {"arg": something}
#
# into this event for the below code to process:
#
#    event:   {"function": function,
#              "arg":      something}
#
# The output from the present program is posted by MQTT:
#
#    topic:   math/<function>/answer
#    payload: {"function": function,
#              "input":    arg,
#              "output":   result,
#              "found":    boolean}
#
# Supported functions are "fib", "fact", and "sin".  Results are cached per function and input;
# "found" is True if a cached result was returned.

def lambda_handler(event, context):
    function = event["function"]
    arg = event["arg"]
    result = -1337
    found = False
    db = boto3.client('dynamodb', region_name='eu-central-1')
    key = {"function": {"S": function},
           "arg":      {"N": str(arg)}}
    resp = db.get_item(TableName='math_data',Key=key)
    if resp is not None and "Item" in resp:
        found = True
        result = float(resp["Item"]["output"]["N"])
    elif function == "fib":
        result = fib(arg)
    elif function == "fact":
        result = fact(arg)
    elif function == "sin":
        result = math.sin(arg)

    if not found:
        db.put_item(TableName='math_data', Item={**key, "output":{"N":str(result)}})

    mqtt_stuff.publish(f"math/{function}/answer",
                       {"function": function,
                       "input":    arg,
                       "output":   result,
                       "found":    found})

def fib(n):
    if n < 2:
        return n
    return fib(n-1) + fib(n-2)

def fact(n):
    if n < 2:
        return n
    return n * fact(n-1)
