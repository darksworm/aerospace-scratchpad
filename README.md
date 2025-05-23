# aerospace-scratchpad

A I3/Sway like scratchpad extension for [AeroSpace WM](https://github.com/nikitabobko/AeroSpace)

**Beta**: I use this daily, but it's still a work in progress, so expect breaking changes. 
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

**Future implementation** As of now show doesn't toggle the scratchpad, but I'll implement this behaviour

### Config Usage

```toml
# ~/.config/aerospace/config.toml
[mode.main.binding] 
# This toggle the scratchpad window show/hide
cmd-ctrl-1 = [
    """exec-and-forget aerospace-scratchpad show Finder || \
                        aerospace-scratchpad move Finder
    """
]
```

## Advanced Usage

See more in [documentation](docs/)

### Dinamic scratchpads mapping

By pairing with [aerospace-marks](https://github.com/cristianoliveira/aerospace-marks) you 
can turn any window into a scratchpad window and bind a toggle key on the fly!

```toml
# ~/.config/aerospace/config.toml
[mode.main.binding] 
# Mark the current window with a given identifier
cmd-shit-ctrl-1 = [
    "exec-and-forget aerospace-marks mark sp-1"
]

# Toggle show/hide the marked window as scratchpad
# If current window is not a scratchpad, move to scratchpad and mark with `sp-1`
# otherwise show/hide window marked as `sp-1`
cmd-ctrl-1 = [
    """exec-and-forget aerospace-scratchpad show "$(aerospace-marks get sp-1 -a)" || \
                       aerospace-scratchpad move "$(aerospace-marks mark sp-1 -s)"
    """
]
```

## Installation

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
[exec.env-vars]
    # Replace 'aerospace-scratchpad/install/path' with the actual path from the above command
    PATH = 'aerospace-scratchpad/install/path/bin:${PATH}'
```

## License

This project is licensed under the terms of the LICENSE file.
