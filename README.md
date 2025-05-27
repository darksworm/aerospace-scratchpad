# aerospace-scratchpad

A I3/Sway like scratchpad extension for [AeroSpace WM](https://github.com/nikitabobko/AeroSpace)

**Beta**: I use this daily, but it's still a work in progress, so expect breaking changes till the 1.0.0 release. 
Please report any issues or ideas in the [issues](https://github.com/cristianoliveira/aerospace-marks/issues) section.

## Demo

https://github.com/user-attachments/assets/108149b7-772e-4972-bc3f-7b0199b8ef6a

From [I3 User Guide](https://i3wm.org/docs/userguide.html#_scratchpad):
> 6.21. Scratchpad
> 
> There are two commands to use any existing window as scratchpad window. move scratchpad will move a window to the scratchpad workspace. This will make it invisible until you show it again. There is no way to open that workspace. Instead, when using scratchpad show, the window will be shown again, as a floating window, centered on your current workspace (using scratchpad show on a visible scratchpad window will make it hidden again, so you can have a keybinding to toggle). Note that this is just a normal floating window, so if you want to "remove it from scratchpad", you can simple make it tiling again (floating toggle).
>
> As the name indicates, this is useful for having a window with your favorite editor always at hand. However, you can also use this for other permanently running applications which you don’t want to see all the time: Your music player, alsamixer, maybe even your mail client…?

[What Are Scratchpads and WHY Are They Good?](https://youtu.be/72ccdlOWe20?si=tyYhdW6_BRJSCSHr)

## Basic Usage

Move the current focused window to the scratchpad workspace.
```text
aerospace-scratchpad move <pattern>
```
Summon or move a window from the scratchpad workspace to the current workspace.
```text
aerospace-scratchpad show <pattern>
```

To find the correct `pattern` run `aerospace list-windows --all --json | grep app-name`

### Config Usage

```toml
# ~/.config/aerospace/config.toml
[mode.main.binding] 
# This toggle the scratchpad window show/hide
cmd-ctrl-1 = "exec-and-forget aerospace-scratchpad show Finder"

# Or using summon instead
cmd-ctrl-2 = [
    """exec-and-forget aerospace-scratchpad summon Finder || \
                        aerospace-scratchpad move Finder
    """
]

# Bring windows one by one to current workspace
ctrl-minus = "exec-and-forget aerospace-scratchpad next"
```

## Advanced Usage

See more in [documentation](docs/)

### Dinamic scratchpads mapping

By pairing with [aerospace-marks](https://github.com/cristianoliveira/aerospace-marks) you 
can turn any window into a scratchpad window and bind a toggle key on the fly!

```toml
# ~/.config/aerospace/config.toml
[mode.main.binding]
# Toggle show/hide the window marked as 'sp-1' as scratchpad
# If window is not marked does nothing
cmd-ctrl-1 = "exec-and-forget aerospace-scratchpad show \"$(aerospace-marks get sp-1 -a)\""

# Mark current window with sp-1 so you can use the shortcut
cmd-shit-ctrl-1 = [
    "exec-and-forget aerospace-marks mark sp-1"
]
```

## Installation

### Using Homebrew

If you have Homebrew installed, you can install `aerospace-scratchpad` using the following command:

```bash
brew install cristianoliveira/tap/aerospace-scratchpad
```

### Nix

If you have Nix installed, you can build and install `aerospace-scratchpad` using the following command:

```bash
nix profile install github:cristianoliveira/aerospace-scratchpad
```

You can also run without installing it by using:

```bash
nix run github:cristianoliveira/aerospace-scratchpad
```

This will build the default package defined in `flake.nix`.

### Go

If you have Go installed, you can install `aerospace-scratchpad` directly using:

```bash
go install github.com/cristianoliveira/aerospace-scratchpad@latest
```

This will download and install the latest version of `aerospace-scratchpad` to your `$GOPATH/bin`.

#### Post installation

After installing, you may need to include aerospace-scratchpad in aerospace context.

Check where the binary is installed, run:
```bash
echo $(which aerospace-scratchpad) | sed 's/\/aerospace-scratchpad//g'
```

And in your config add:
```toml
[exec]
    inherit-env-vars = true
# OR
[exec.env-vars]
    # Replace 'aerospace-scratchpad/install/path' with the actual path from the command above
    PATH = 'aerospace-scratchpad/install/path/bin:${PATH}'
```
## How does it work?

This extension uses IPC (Inter-Process Communication) to communicate directly with the AeroSpace Unix socket, just like the built-in AeroSpace CLI. By avoiding repeated process spawning, this approach offers lower latency and better efficiency, specially when one have to query AeroSpace many time.

See: https://github.com/cristianoliveira/aerospace-ipc

## License

This project is licensed under the terms of the LICENSE file.
