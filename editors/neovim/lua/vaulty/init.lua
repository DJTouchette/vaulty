local M = {}

--- Default configuration.
M.config = {
  host = "127.0.0.1",
  port = 19876,
  -- Polling interval in milliseconds for the lualine component (0 = no poll).
  poll_interval_ms = 5000,
}

-- Cached status used by the lualine component so it never blocks the UI.
local cached_running = false
local poll_timer = nil

-----------------------------------------------------------------------------
-- Helpers
-----------------------------------------------------------------------------

--- Check whether the Vaulty daemon is reachable.
--- Shells out to curl with a short timeout.  Returns true/false.
---@return boolean
local function is_daemon_running()
  local result = vim.fn.system(
    "curl -s -o /dev/null -w '%{http_code}' --connect-timeout 1 http://"
      .. M.config.host
      .. ":"
      .. M.config.port
      .. "/ 2>/dev/null"
  )
  -- Any HTTP response (including 404) means the daemon is up.
  local code = tonumber(result) or 0
  return code > 0
end

--- Async status check that updates the cache and optionally calls a callback.
---@param callback? fun(running: boolean)
local function check_status_async(callback)
  vim.fn.jobstart(
    "curl -s -o /dev/null -w '%{http_code}' --connect-timeout 1 http://"
      .. M.config.host
      .. ":"
      .. M.config.port
      .. "/ 2>/dev/null",
    {
      stdout_buffered = true,
      on_stdout = function(_, data)
        local raw = table.concat(data, "")
        local code = tonumber(raw) or 0
        cached_running = code > 0
        if callback then
          callback(cached_running)
        end
      end,
    }
  )
end

-----------------------------------------------------------------------------
-- Commands
-----------------------------------------------------------------------------

--- Show daemon status in a notification.
function M.show_status()
  check_status_async(function(running)
    vim.schedule(function()
      if running then
        vim.notify(
          "Vaulty daemon is running on " .. M.config.host .. ":" .. M.config.port,
          vim.log.levels.INFO
        )
      else
        vim.notify("Vaulty daemon is not running.", vim.log.levels.WARN)
      end
    end)
  end)
end

--- Start the daemon in a terminal buffer.
function M.start_daemon()
  vim.cmd("split | terminal vaulty start")
end

--- Stop the daemon in a terminal buffer.
function M.stop_daemon()
  vim.cmd("split | terminal vaulty stop")
end

-----------------------------------------------------------------------------
-- Lualine component
-----------------------------------------------------------------------------

--- Returns a string suitable for use as a lualine component.
--- The value is updated asynchronously so it never blocks rendering.
---
--- Example lualine config:
---   lualine_x = { require("vaulty").lualine }
---
---@return string
function M.lualine()
  if cached_running then
    return "Vaulty: Running"
  end
  return "Vaulty: Stopped"
end

--- Color function for lualine that returns green/red based on status.
---@return table
function M.lualine_color()
  if cached_running then
    return { fg = "#a6e3a1" } -- green
  end
  return { fg = "#f38ba8" } -- red
end

-----------------------------------------------------------------------------
-- Setup
-----------------------------------------------------------------------------

--- Plugin setup.  Call from your init.lua:
---   require("vaulty").setup({ port = 19876 })
---@param opts? table
function M.setup(opts)
  M.config = vim.tbl_deep_extend("force", M.config, opts or {})

  -- Register user commands.
  vim.api.nvim_create_user_command("VaultyStatus", function()
    M.show_status()
  end, { desc = "Show Vaulty daemon status" })

  vim.api.nvim_create_user_command("VaultyStart", function()
    M.start_daemon()
  end, { desc = "Start the Vaulty daemon" })

  vim.api.nvim_create_user_command("VaultyStop", function()
    M.stop_daemon()
  end, { desc = "Stop the Vaulty daemon" })

  -- Start background polling for the lualine component.
  if M.config.poll_interval_ms > 0 then
    -- Do an initial check immediately.
    check_status_async()

    poll_timer = vim.loop.new_timer()
    if poll_timer then
      poll_timer:start(
        M.config.poll_interval_ms,
        M.config.poll_interval_ms,
        vim.schedule_wrap(function()
          check_status_async()
        end)
      )
    end
  end
end

return M
