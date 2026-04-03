import * as vscode from "vscode";
import * as http from "http";

const DAEMON_URL = "http://127.0.0.1:19876";
const POLL_INTERVAL_MS = 5000;

let statusBarItem: vscode.StatusBarItem;
let pollTimer: ReturnType<typeof setInterval> | undefined;

// ---------------------------------------------------------------------------
// Activation
// ---------------------------------------------------------------------------

export function activate(context: vscode.ExtensionContext): void {
  // Create a status-bar item on the left side.
  statusBarItem = vscode.window.createStatusBarItem(
    vscode.StatusBarAlignment.Left,
    50
  );
  statusBarItem.command = "vaulty.showStatus";
  context.subscriptions.push(statusBarItem);

  // Register commands.
  context.subscriptions.push(
    vscode.commands.registerCommand("vaulty.showStatus", cmdShowStatus),
    vscode.commands.registerCommand("vaulty.startDaemon", cmdStartDaemon),
    vscode.commands.registerCommand("vaulty.stopDaemon", cmdStopDaemon)
  );

  // Initial check then start polling.
  refreshStatus();
  pollTimer = setInterval(refreshStatus, POLL_INTERVAL_MS);
  context.subscriptions.push({
    dispose: () => {
      if (pollTimer) {
        clearInterval(pollTimer);
      }
    },
  });
}

export function deactivate(): void {
  if (pollTimer) {
    clearInterval(pollTimer);
    pollTimer = undefined;
  }
}

// ---------------------------------------------------------------------------
// Daemon health check
// ---------------------------------------------------------------------------

function isDaemonRunning(): Promise<boolean> {
  return new Promise((resolve) => {
    const req = http.get(DAEMON_URL, { timeout: 2000 }, (res) => {
      // Any response (even 404) means the daemon process is listening.
      res.resume();
      resolve(true);
    });
    req.on("error", () => resolve(false));
    req.on("timeout", () => {
      req.destroy();
      resolve(false);
    });
  });
}

async function refreshStatus(): Promise<void> {
  const running = await isDaemonRunning();
  if (running) {
    statusBarItem.text = "$(check) Vaulty: Running";
    statusBarItem.color = new vscode.ThemeColor(
      "statusBarItem.foreground"
    );
    statusBarItem.backgroundColor = new vscode.ThemeColor(
      "statusBarItem.prominentBackground"
    );
    statusBarItem.tooltip = "Vaulty daemon is running on port 19876";
  } else {
    statusBarItem.text = "$(error) Vaulty: Stopped";
    statusBarItem.color = new vscode.ThemeColor(
      "statusBarItem.errorForeground"
    );
    statusBarItem.backgroundColor = new vscode.ThemeColor(
      "statusBarItem.errorBackground"
    );
    statusBarItem.tooltip =
      "Vaulty daemon is not running. Click to see options.";
  }
  statusBarItem.show();
}

// ---------------------------------------------------------------------------
// Commands
// ---------------------------------------------------------------------------

async function cmdShowStatus(): Promise<void> {
  const running = await isDaemonRunning();
  if (running) {
    const choice = await vscode.window.showInformationMessage(
      "Vaulty daemon is running on 127.0.0.1:19876.",
      "Stop Daemon"
    );
    if (choice === "Stop Daemon") {
      await cmdStopDaemon();
    }
  } else {
    const choice = await vscode.window.showWarningMessage(
      "Vaulty daemon is not running.",
      "Start Daemon"
    );
    if (choice === "Start Daemon") {
      await cmdStartDaemon();
    }
  }
}

function cmdStartDaemon(): void {
  const terminal = getOrCreateTerminal();
  terminal.show();
  terminal.sendText("vaulty start");
  // Give the daemon a moment to start, then refresh.
  setTimeout(refreshStatus, 2000);
}

function cmdStopDaemon(): void {
  const terminal = getOrCreateTerminal();
  terminal.show();
  terminal.sendText("vaulty stop");
  setTimeout(refreshStatus, 2000);
}

function getOrCreateTerminal(): vscode.Terminal {
  const existing = vscode.window.terminals.find(
    (t) => t.name === "Vaulty"
  );
  if (existing) {
    return existing;
  }
  return vscode.window.createTerminal("Vaulty");
}
