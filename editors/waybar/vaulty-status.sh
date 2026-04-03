#!/usr/bin/env bash
# vaulty-status.sh — Waybar custom module script for Vaulty daemon status.
# Outputs JSON in the waybar custom module format.

VAULTY_URL="http://127.0.0.1:19876/"

if curl -s -o /dev/null -w '' --connect-timeout 1 "$VAULTY_URL" 2>/dev/null; then
    echo '{"text": " Vaulty", "tooltip": "Vaulty daemon running", "class": "running"}'
else
    echo '{"text": " Vaulty", "tooltip": "Vaulty daemon stopped", "class": "stopped"}'
fi
