---
title: "OS Keychain"
weight: 4
---

# OS Keychain Integration

Typing your vault passphrase every time you start Vaulty gets old fast. Your operating system has a perfectly good keychain — let's use it.

## Supported Keychains

| OS | Keychain |
|----|----------|
| macOS | Keychain (the one with the lock icon) |
| Linux | GNOME Keyring / KWallet / Secret Service |
| Windows | Windows Credential Manager |

## Save Your Passphrase

```bash
vaulty keychain save
```

This prompts for your passphrase and stores it securely in your OS keychain. From now on, `vaulty start` and `vaulty mcp` will grab the passphrase automatically — no prompting, no env vars.

## Check Status

```bash
vaulty keychain status
```

Tells you whether a passphrase is currently stored in the keychain.

## Remove It

```bash
vaulty keychain delete
```

Changed your passphrase? Moving to team mode? Just want to feel the friction of typing it every time? This removes the stored passphrase.

## How Vaulty Finds Your Passphrase

Vaulty checks these sources in order:

1. `VAULTY_PASSPHRASE` environment variable (highest priority)
2. OS keychain
3. Interactive prompt (asks you to type it)

For MCP mode specifically, the interactive prompt doesn't work (stdin is used for JSON-RPC), so you need either the env var or the keychain. The keychain is the cleaner option — no secrets in config files.

## Security Notes

Your OS keychain is protected by your login credentials. On macOS, apps need explicit permission to access keychain items. On Linux, the keyring unlocks when you log in. On Windows, it's tied to your user account.

Is it *as* secure as typing the passphrase every time? No. Is it secure enough for most threat models while being dramatically more convenient? Yes. Pick your trade-off.
