#!/bin/bash
# This script requires jq (https://jqlang.org/)
QUEUE=$HOME/.local/state/scratchpad-queue
touch "$QUEUE" 2>/dev/null || true

COMMAND=$1

if [ -z "$COMMAND" ]; then
  echo "Usage: $0 <push|pop|cycle>"
  exit 1
fi

function push_into_stack() {
  aerospace-scratchpad move --output=json \
    | jq -r 'select(.result=="ok") | .window_id' \
    >> "$QUEUE"
}

function pop_from_stack() {
  wid=$(head -n1 "$QUEUE")
  tail -n +2 "$QUEUE" > "$QUEUE.tmp" && mv "$QUEUE.tmp" "$QUEUE"
  aerospace-scratchpad show ".+" --filter window-id="^${wid}$" --output=json
}

function push_all_scratchpads_into_stack() {
  aerospace-scratchpad move --all-floating -o json |
    jq -r 'select(.result=="ok") | .window_id' >> "$QUEUE"
}

if [ "$COMMAND" = "push" ]; then
  push_into_stack
  exit 0
fi

if [ "$COMMAND" = "pop" ]; then
  if [ -s "$QUEUE" ]; then
    pop_from_stack
  else
    echo "Queue empty"
  fi
fi

# This implements a similar behavior as i3/SwayWM
# `bindsym $mod+minus scratchpad show` which cycles through the stack
if [ "$COMMAND" = "cycle" ]; then
  if [ -s "$QUEUE" ]; then
    push_all_scratchpads_into_stack
    pop_from_stack
  else
    echo "Queue empty"
  fi
fi
