# T-132: Systemd user service for Vaulty daemon

**Epic:** 13 — Wayland / Arch Integration
**Status:** done
**Priority:** P1

## Description

Create a systemd user service file so the daemon can be managed with `systemctl --user start vaulty` and auto-started on login.

## Acceptance Criteria

- [ ] `vaulty.service` systemd user unit file
- [ ] Starts daemon in foreground mode
- [ ] Reads passphrase from keychain (no interactive prompt)
- [ ] `systemctl --user enable vaulty` for auto-start
- [ ] Included in AUR package
