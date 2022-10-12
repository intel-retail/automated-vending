#!/bin/bash
# Copyright © 2020 Intel Corporation. All rights reserved.
# SPDX-License-Identifier: BSD-3-Clause
cards=("0003293374" "0003278380" "0003292371")

declare -i timeoutCounter=0


while true ; do 
    echo "Working..."
    result=$(curl -X GET http://localhost:48094/status | grep -n '{\\"lock1_status\\":1,\\"lock2_status\\":1,\\"door_closed\\":true') # -n shows line number
    echo "timeout value = $timeoutCounter"
    if [ ! -z $result ] ; then
        echo "Found correct state!"
        break

    fi
    if [[ $timeoutCounter -eq 5 ]]; then
        echo "reached timeout"
        exit 1
    else 
        echo "Waiting for correct state"
        timeoutCounter=$(( timeoutCounter + 1 ))
    fi
    sleep 1
done


for i in "${cards[@]}";
do
    curl -X GET http://localhost:48094/status | jq .
    # make sure locks are 1 and door is true

    echo $i
    echo 
    
    curl -X PUT -H "Content-Type: application/json" -d "{\"card-number\":\"$i\"}" http://localhost:48098/api/v2/device/name/card-reader/card-number | jq .
    echo
    echo "card read!!!!!!!!!!!!!"
    sleep 5
    curl -X GET http://localhost:48094/status | jq .
    # should show lock1_status: 0 (false)
    echo
    echo "Card status!!!!!!!!!!!!!"
    # open door
    curl -X PUT -H "Content-Type: application/json" -d '{"setDoorClosed":"0"}' http://localhost:48097/api/v2/device/name/controller-board/setDoorClosed | jq .
    echo
    echo "open door!!!!!!!!!!!!!"
    sleep 4
    curl -X GET http://localhost:48094/status | jq .
    # should show door:false
    echo
    echo "open door status!!!!!!!!!!!!!"
    curl -X PUT -H "Content-Type: application/json" -d '{"setDoorClosed":"1"}' http://localhost:48097/api/v2/device/name/controller-board/setDoorClosed | jq .
    echo
    echo "close door !!!!!!!!!!!!!"
    sleep 4
    curl -X GET http://localhost:48094/status | jq .
    # should show door:true
    echo
    echo "close door status!!!!!!!!!!!!!"
    curl -X GET http://localhost:48095/inventory | jq .
    echo
    echo "Get inventory !!!!!!!!!!!!!"
    curl -X GET http://localhost:48095/auditlog | jq .
    echo
    echo "Get Audit log!!!!!!!!!!!!!"
    curl -X GET http://localhost:48093/ledger | jq .
    echo
    echo "Get ledger!!!!!!!!!!!!"
    sleep 30
    echo "GOING TO NEXT CARD NUMBER"
done