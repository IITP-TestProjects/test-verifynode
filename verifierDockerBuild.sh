#!/bin/bash
echo "version: $1"

docker build --tag verifier:$1 .
