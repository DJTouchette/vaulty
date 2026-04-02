# Vaulty systemd User Service

This directory contains a systemd user service file for running Vaulty as a background daemon.

## Installation

Copy the service file to your systemd user directory:

```bash
cp vaulty.service ~/.config/systemd/user/
```

Reload systemd to pick up the new service file:

```bash
systemctl --user daemon-reload
```

## Usage

Start the service:

```bash
systemctl --user start vaulty
```

Check the service status:

```bash
systemctl --user status vaulty
```

Stop the service:

```bash
systemctl --user stop vaulty
```

## Auto-Start on Login

To start Vaulty automatically when you log in:

```bash
systemctl --user enable vaulty
```

## Passphrase Configuration

Since the systemd service runs non-interactively, there is no terminal prompt for
your passphrase. You must configure the passphrase before starting the service
using one of the following methods.

### Option 1: OS Keychain (Recommended)

Save your passphrase to the OS keychain:

```bash
vaulty keychain save
```

Vaulty will retrieve it automatically on startup.

### Option 2: Environment Variable

Set the `VAULTY_PASSPHRASE` environment variable in the systemd user session:

```bash
systemctl --user set-environment VAULTY_PASSPHRASE="your-passphrase"
```

Then start or restart the service. Note that the variable persists only until the
user session ends. To make it permanent, add an `Environment=` directive to the
service file or use a drop-in override.

## Viewing Logs

View Vaulty service logs with journalctl:

```bash
journalctl --user -u vaulty
```

Follow logs in real time:

```bash
journalctl --user -u vaulty -f
```
