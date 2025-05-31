#!/bin/bash

set -euo pipefail

# Current benchmark using aerospace-scratchpad
# 2025/05/31 08:13:36 GetAllWindows took 31.3675ms
# 2025/05/31 08:13:36 GetFocusedWorkspace took 50.599167ms
# 2025/05/31 08:13:36 IsWindowInWorkspace took 50.858125ms
# 2025/05/31 08:13:36 MoveWindowToWorkspace took 70.47675ms
# 2025/05/31 08:13:36 SetFocusByWindowID took 83.191167ms
# Window '55 | Finder  | .scratchpad | com.apple.finder' is summoned
# aerospace-scratchpad show Finder  0.01s user 0.01s system 18% cpu 0.116 total
#
# Previous benchmark using aerospace-scratchpad
# aerospace-scratchpad show Finder  0.01s user 0.01s system 12% cpu 0.125 total
# aerospace-scratchpad show Finder  0.01s user 0.01s system 14% cpu 0.097 total
# aerospace-scratchpad show Finder  0.01s user 0.01s system 12% cpu 0.140 total

# Test cli `time aerospace-scratchpad show <window-name>`
# Test script `time scripts/benchmark.sh <window-id> <workspace-id>`

WINDOW_APP_NAME=${1?"Missing window app name"}

WINDOW_ID="$(aerospace list-windows --all | grep "Finder" | awk '{print $1}')"
WORKSPACE_DEST="$(aerospace list-workspaces --focused | awk '{print $1}')"
aerospace list-workspaces --focused --json > /dev/null
aerospace list-windows --focused --json > /dev/null
aerospace move-node-to-workspace "$WORKSPACE_DEST" --window-id "$WINDOW_ID" > /dev/null
aerospace focus --window-id "$WINDOW_ID" > /dev/null

# Latest output:
# scripts/benchmark.sh 55 2  0.11s user 0.07s system 69% cpu 0.263 total
# scripts/benchmark.sh 55 2  0.11s user 0.06s system 69% cpu 0.252 total
# scripts/benchmark.sh 55 2  0.11s user 0.06s system 72% cpu 0.246 total
