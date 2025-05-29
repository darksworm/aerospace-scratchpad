#!/bin/bash

set -euo pipefail

# Current benchmark using aerospace-scratchpad
# 2025/05/29 15:36:03 GetAllWindows took 54.792917ms
# 2025/05/29 15:36:03 GetFocusedWorkspace took 109.821833ms
# 2025/05/29 15:36:03 IsWindowInWorkspace took 109.871125ms
# 2025/05/29 15:36:03 Bef MoveWindowToWorkspace took 109.872917ms
# 2025/05/29 15:36:03 MoveWindowToWorkspace took 159.813167ms
# 2025/05/29 15:36:03 SetFocusByWindowID took 346.021625ms
# Window '10539 | Finder  | .scratchpad | com.apple.finder' is summoned
# 2025/05/29 15:36:03 Finished in 346.048042ms

# Test with `time scripts/benchmark.sh <window-id> <workspace-id>`

WINDOW_ID=${1?"Missing window id"}
WORKSPACE_DEST=${2?"Missing workspace"}

aerospace list-windows --all --json > /dev/null
aerospace list-workspaces --focused --json > /dev/null
aerospace list-windows --workspace 2 --json > /dev/null
aerospace list-windows --focused --json > /dev/null
aerospace move-node-to-workspace "$WORKSPACE_DEST" --window-id "$WINDOW_ID" > /dev/null
aerospace focus --window-id "$WINDOW_ID" > /dev/null
