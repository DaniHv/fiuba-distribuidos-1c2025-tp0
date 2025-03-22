#!/bin/bash

output="$(docker container run \
            --rm \
            --network=tp0_testing_net \
            alpine:3.21.3 \
            sh -c 'echo "Hello, world!" | nc server 12345'
        )"

if [[ $output == "Hello, world!" ]]; then
    echo "action: test_echo_server | result: success"
    exit 0
else
    echo "action: test_echo_server | result: fail"
    exit 1
fi