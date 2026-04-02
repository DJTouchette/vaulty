---
title: "Installation"
weight: 1
---

# Installation

Vaulty is a single Go binary. No Docker, no containers, no cloud accounts, no runtime dependencies, no calling your DevOps team. Just a binary that does one thing well.

## Go Install (Recommended)

If you have Go installed (and if you're reading docs for a Go tool, you probably do):

```bash
go install github.com/djtouchette/vaulty/cmd/vaulty@latest
```

That's it. You now have a `vaulty` binary in your `$GOPATH/bin`. Make sure that's in your `$PATH` — but you already knew that.

## From Source

For those who like to read the code before trusting it with their secrets (respect):

```bash
git clone https://github.com/djtouchette/vaulty.git
cd vaulty
go build ./cmd/vaulty
```

The binary lands in the current directory. Move it wherever you like.

## Verify It Works

```bash
vaulty --help
```

If you see a help message, you're golden. If you see `command not found`, your `$PATH` is having a bad day — go check on it.

## Next Up

Time to [create your first vault]({{< relref "/docs/getting-started/first-vault" >}}).
