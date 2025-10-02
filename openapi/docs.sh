#!/bin/bash

if [ -z "$1" ]; then
  echo "Usage: $0 <destination>"
  exit 1
fi

DEST="$1"

npx @redocly/cli build-docs openapi.yml --o "$DEST"