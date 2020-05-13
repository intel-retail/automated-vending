#!/usr/bin/env python3

# Copyright © 2020 Intel Corporation. All rights reserved.
# SPDX-License-Identifier: BSD-3-Clause

import os
import sys
import time
import json
import sched
import random
import collections
from argparse import ArgumentParser

import paho.mqtt.client as mqtt

CMD_HEARTBEAT='inferenceHeartbeat'
CMD_HEARTBEAT_RESPONSE='inferencePong'
CMD_CV_TRIGGER='inferenceDoorStatus'
CMD_SKU_DELTA='inferenceSkuDelta'
STATUS_DOOR_OPEN='false'
STATUS_DOOR_CLOSE='true'

def build_argparser():
  parser = ArgumentParser()
  parser.add_argument("-b", "--broker", help="MQTT broker IP", required=True, type=str)
  parser.add_argument("-p", "--port", help="MQTT broker port", required=True, type=int)
  parser.add_argument("-s", "--test_sequence", help="JSON file specifying the sequence of responses to send back to EdgeX", required=True, type=str)
  parser.add_argument("-k", "--keepalive", help="MQTT keepalive", default=60, type=int)
  parser.add_argument("-c", "--command_topic", help="Command topic. MQTT topic from which EdgeX commands are received.", \
    default="Inference/CommandTopic", type=str)
  parser.add_argument("-r", "--response_topic", help="Response topic. MQTT topic to which EdgeX commands are acknowledged.", \
    default="Inference/ResponseTopic", type=str)
  parser.add_argument("-d", "--data_topic", help="Response topic. MQTT topic to which SKU delta are published.", \
    default="Inference/DataTopic", type=str)
  parser.add_argument("-D", "--delay", help="Number of seconds to delay SKU delta response",
                       default=5, type=int)

  return parser

class SKU_Deltas():
  def __init__(self, filename):
        with open(filename, 'r') as fdin:
          try:
            self.sku_deltas = json.load(fdin)
          except json.JSONDecodeError as e:
            print(f'Error parsing {filename}!')
            print(f'{e.msg} at position {e.pos}')
            sys.exit(-1)
        if not isinstance(self.sku_deltas, list):
              self.sku_deltas = [self.sku_deltas]
        self.idx = 0
  def __next__(self):
    item_list = []
    for k, v in self.sku_deltas[self.idx].items():
      item_list.append({'SKU':k, 'delta':v})
    self.idx = (self.idx + 1)%len(self.sku_deltas)
    return item_list

def main():

  sch = sched.scheduler(time.time, time.sleep)
  args = build_argparser().parse_args()
  sku_deltas = SKU_Deltas(args.test_sequence)

  def on_connect(client, userdata, flags, rc):
    print(f"Connected to broker {args.broker}:{args.port}")
    client.subscribe(args.command_topic)

  def publish_sku_delta():
    """
    Publish SKU delta to EdgeX
    """
    sku_delta_json = json.dumps(next(sku_deltas))

    msg_dict = {
      "name":"Inference-MQTT-device",
      "cmd":CMD_SKU_DELTA,
      "method":"get",
      CMD_SKU_DELTA: sku_delta_json
    }
    print(f"Sending SKU Delta on {args.data_topic} {msg_dict}")
    client.publish(args.data_topic, payload=json.dumps(msg_dict))

  def handle_ping(msg_dict):
    """
    Response to heartbeat messages
    """
    msg_dict[CMD_HEARTBEAT] = CMD_HEARTBEAT_RESPONSE
    return msg_dict

  def handle_trigger(msg_dict):
    """
    Response to the trigger command and schedule the
    SKU delta message for args.delay seconds
    """
    # Only send SKU delta on door close
    if msg_dict[CMD_CV_TRIGGER] == STATUS_DOOR_CLOSE:
      sch.enter(args.delay, 1, publish_sku_delta)
    msg_dict[CMD_CV_TRIGGER] = "Got it!"
    return msg_dict

  def on_message(client, userdata, msg):
    """
    This callback is invoke for every MQTT messages we received
    """
    if msg.topic == args.command_topic:
      msg_dict = json.loads(msg.payload)
      print(f"Message received on {msg.topic} {str(msg.payload)}")

      if msg_dict["cmd"] == CMD_HEARTBEAT:
        msg_dict2 = handle_ping(msg_dict)
      elif msg_dict["cmd"] == CMD_CV_TRIGGER:
        msg_dict2 = handle_trigger(msg_dict)

      print(f"Sending response on {args.response_topic} {msg_dict2}")
      client.publish(args.response_topic, payload=json.dumps(msg_dict2))

  client = mqtt.Client()
  client.on_connect = on_connect
  client.on_message = on_message

  print(f"Trying to connect to {args.broker}:{args.port}")
  client.connect(args.broker, args.port, args.keepalive)

  while True:
    try:
      client.loop()
      sch.run(blocking=False)
    except KeyboardInterrupt:
      break

  client.disconnect()
  print('Done!')

if __name__ == "__main__":
  main()
