# Vaulty VS Code Extension

Shows the Vaulty daemon status in the VS Code status bar and provides
commands to start and stop the daemon.

## Features

- **Status bar indicator** -- displays whether the Vaulty daemon is running
  (green) or stopped (red). Updates automatically every 5 seconds.
- **Commands** (open the Command Palette with `Ctrl+Shift+P` / `Cmd+Shift+P`):
  - `Vaulty: Show Daemon Status` -- show current status with an action button.
  - `Vaulty: Start Daemon` -- runs `vaulty start` in an integrated terminal.
  - `Vaulty: Stop Daemon` -- runs `vaulty stop` in an integrated terminal.

## Installation (development)

```bash
cd editors/vscode
npm install
npm run compile
```

Then press **F5** in VS Code to launch an Extension Development Host with the
extension loaded.

## Packaging

Install the `vsce` tool and build a `.vsix`:

```bash
npm install -g @vscode/vsce
vsce package
```

Install the resulting `.vsix` via:

```
code --install-extension vaulty-0.1.0.vsix
```

## Requirements

- VS Code 1.85.0 or later.
- The `vaulty` CLI must be on your `$PATH`.
- The daemon listens on `127.0.0.1:19876` (the default HTTP port).
