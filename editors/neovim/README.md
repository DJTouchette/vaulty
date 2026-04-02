# Vaulty Neovim Plugin

Shows Vaulty daemon status inside Neovim and provides commands to
start/stop the daemon. Includes an optional lualine component.

## Installation

### lazy.nvim

```lua
{
  "djtouchette/vaulty",
  config = function()
    require("vaulty").setup()
  end,
  submodules = false,
  -- Only load the Neovim plugin, not the whole repo.
  dir = vim.fn.stdpath("data") .. "/lazy/vaulty/editors/neovim",
}
```

Or point directly at the subdirectory if you have the repo cloned locally:

```lua
{
  dir = "~/path/to/vaulty/editors/neovim",
  config = function()
    require("vaulty").setup()
  end,
}
```

### packer

```lua
use {
  "djtouchette/vaulty",
  rtp = "editors/neovim",
  config = function()
    require("vaulty").setup()
  end,
}
```

## Commands

| Command         | Description                          |
| --------------- | ------------------------------------ |
| `:VaultyStatus` | Show daemon status in a notification |
| `:VaultyStart`  | Start the daemon in a terminal split |
| `:VaultyStop`   | Stop the daemon in a terminal split  |

## Configuration

```lua
require("vaulty").setup({
  host = "127.0.0.1",  -- daemon address
  port = 19876,        -- daemon HTTP port
  poll_interval_ms = 5000, -- background poll interval (0 to disable)
})
```

## Lualine integration

Add the Vaulty component to your lualine config:

```lua
require("lualine").setup({
  sections = {
    lualine_x = {
      {
        require("vaulty").lualine,
        color = require("vaulty").lualine_color,
      },
    },
  },
})
```

The component updates asynchronously and never blocks the UI.

## Requirements

- Neovim 0.9 or later.
- `curl` on your `$PATH` (used for the health check).
- The `vaulty` CLI must be on your `$PATH`.
