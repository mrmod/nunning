#!/bin/bash

PINOT_INTERACTIVE="true"


if [ $PINOT_INTERACTIVE = "true" ]; then
    echo "Starting pinot in interactive mode"
    
    docker-compose -f apache-pinot/docker-compose.yml up
    exit 0
fi  

if [ $PINOT_INTERACTIVE = "false" ]; then
    echo "Starting pinot in detached mode"
    docker-compose -f apache-pinot/docker-compose.yml up -d
    exit 0
fi