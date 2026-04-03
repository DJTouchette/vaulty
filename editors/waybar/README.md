# Vaulty Waybar Module

A custom waybar module that shows whether the Vaulty daemon is running.
The module polls the daemon every 5 seconds and displays a color-coded
status indicator in your bar.

Works with Hyprland, Sway, and other wlroots-based compositors.

## Installation

### 1. Copy the status script

```bash
mkdir -p ~/.config/vaulty
cp editors/waybar/vaulty-status.sh ~/.config/vaulty/vaulty-status.sh
chmod +x ~/.config/vaulty/vaulty-status.sh
```

### 2. Add the module to your waybar config

Open `~/.config/waybar/config.jsonc` (or `config`) and add `"custom/vaulty"`
to one of the module arrays:

```jsonc
"modules-right": ["custom/vaulty", "clock", "tray"]
```

Then add the module definition (see `config.jsonc` in this directory for the
full block):

```jsonc
"custom/vaulty": {
  "exec": "~/.config/vaulty/vaulty-status.sh",
  "return-type": "json",
  "interval": 5,
  "on-click": "vaulty start || vaulty stop",
  "format": "{}",
  "tooltip": true
}
```

### 3. Add the CSS

Append the contents of `style.css` to `~/.config/waybar/style.css`:

```css
#custom-vaulty.running { color: #a6e3a1; }
#custom-vaulty.stopped { color: #f38ba8; }
```

### 4. Reload waybar

```bash
killall -SIGUSR2 waybar
```

Or restart it manually.

## Click behavior

Clicking the module toggles the daemon: it tries `vaulty start` first, and
if the daemon is already running it falls back to `vaulty stop`. Make sure
the `vaulty` CLI is on your `$PATH`.

## Requirements

- `curl` on your `$PATH` (used for the health check).
- The `vaulty` CLI on your `$PATH`.
- Waybar (any recent version with custom module support).
