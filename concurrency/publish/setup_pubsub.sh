#!/bin/sh

# Wait for pubsub-emulator to come up
bash -c 'while [[ "$(curl -s -o /dev/null -w ''%{http_code}'' 'localhost:8539')" != "200" ]]; do sleep 1; done'

curl -X PUT http://localhost:8539/v1/projects/project/topics/eq-submission-topic
curl -X PUT http://localhost:8539/v1/projects/project/subscriptions/rm-receipt-subscription -H 'Content-Type: application/json' -d '{"topic": "projects/project/topics/eq-submission-topic"}'