#!/bin/bash
# This script requires jq (https://jqlang.org/)
QUEUE=$HOME/.local/state/scratchpad-queue
touch "$QUEUE" 2>/dev/null || true

COMMAND=$1

if [ -z "$COMMAND" ]; then
  echo "Usage: $0 <push|pop|next>"
  exit 1
fi

function push_into() {
  aerospace-scratchpad move --output=json \
    | jq -r 'select(.result=="ok") | .window_id' \
    >> "$QUEUE"
}

function pop_from() {
  wid=$(head -n1 "$QUEUE")
  tail -n +2 "$QUEUE" > "$QUEUE.tmp" && mv "$QUEUE.tmp" "$QUEUE"
  aerospace-scratchpad show ".+" --filter window-id="^${wid}$" --output=json
}

if [ "$COMMAND" = "push" ]; then
  # Push IDs of moved windows
  push_into
  exit 0
fi

if [ "$COMMAND" = "pop" ]; then
  # Pop oldest window ID and summon it back
  if [ -s "$QUEUE" ]; then
    pop_from
  else
    echo "Queue empty"
  fi
fi

# This is implements a similar behavior as i3/SwayWM:
# `bindsym $mod+minus scratchpad show` 
if [ "$COMMAND" = "next" ]; then
  if [ -s "$QUEUE" ]; then
    # First send all floating windows back
    aerospace-scratchpad move --all-floating -o json |
      jq -r 'select(.result=="ok") | .window_id' >> "$QUEUE"

    # Then pop the next window
    pop_from
  else
    echo "Queue empty"
  fi
fi
