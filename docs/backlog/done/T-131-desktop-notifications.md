# T-131: Desktop notifications for denied requests

**Epic:** 13 — Wayland / Arch Integration
**Status:** done
**Priority:** P2

## Description

Send a desktop notification (via notify-send / libnotify) when Vaulty denies a request due to policy violation. Critical for catching prompt injection attempts in real-time.

## Acceptance Criteria

- [ ] notify-send called on policy denial (domain or command mismatch)
- [ ] Notification shows secret name, target, and reason
- [ ] Configurable in vaulty.toml (`notifications = true`)
- [ ] Works with mako, dunst, and other libnotify-compatible daemons
