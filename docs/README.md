# aerospace-scratchpad

Here you find a extensive documentation about the CLI.

## Command: `move`

Move the current focused window to the scratchpad workspace. The scratchpad workspace is a special workspace that is not visible to the user. The window will be hidden until you show it again.
You actually can see this in your workspace list, but it can be ignored, it just store the windows that are hidden.

### USAGE

`pattern` is a regex pattern to match the window name.

```bash
aerospace-scratchpad move <pattern>
```

For more details:
```bash
aerospace-scratchpad move --help
```

## Command: `show`

Summon a window from the scratchpad workspace to the current workspace. Automatically toggles the window's layout to floating, but it can be configured.

`pattern` is a regex pattern to match the window name.

USAGE: `aerospace-scratchpad show <pattern>`

For more details:
```bash
aerospace-scratchpad show --help
```
