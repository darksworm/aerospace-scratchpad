# aerospace-scratchpad

Here you find a extensive documentation about the CLI.

## Command: `move`

Move the current focused window to the scratchpad workspace. The window will be hidden until you show it again.
You actually can see this in your workspace list, but it can be ignored, it just to store the windows that are "hidden".

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

Similar to Sway's `show` this command will:

 - Show a window that was previously moved to the scratchpad workspace.
 - Move the window to scratchpad if it is focused and matches the <pattern>
 - If a scratchpad window is in another workspace, it will move it to the current workspace.
 - If a scratchpad window is already in the current workspace, it will focus it.
 - If multiple windows match a pattern, it will bring all of them to the current workspace.
 - If no window matches a pattern, it will do nothing.

The `pattern` is a regex pattern to match the "App Name"

USAGE: `aerospace-scratchpad show <pattern>`

For more details:
```bash
aerospace-scratchpad show --help
```

## Command: `move`

Move the current focused window to the scratchpad workspace if matches the given pattern.

### USAGE

The `pattern` is a regex pattern to match the "App Name"

```bash
aerospace-scratchpad move <pattern>
```

## Command: `summon`

Different from command `show`, this command will only summon the window to the current workspace and focus it.

### USAGE

The `pattern` is a regex pattern to match the "App Name"

```bash
aerospace-scratchpad summon <pattern>
```

## Implementation details

### Scratchpad workspace

It will send the window to a "special" workspace that is called `.scratchpad`. This workspace is like any other workspace, but can be ignored. The window will be hidden until you show it again.

### Comunication with AeroSpaceWM

The communication with AeroSpaceWM is done through an ipc socket client
See: https://github.com/cristianoliveira/aerospace-ipc
