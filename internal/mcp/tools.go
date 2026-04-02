package mcp

// toolDefinitions returns the MCP tool definitions for Vaulty.
func toolDefinitions() []map[string]any {
	return []map[string]any{
		{
			"name":        "vaulty_request",
			"description": "Make an authenticated HTTP request through Vaulty. The credential is injected by Vaulty — you never see the raw value. Use this instead of curl or fetch when you need to call an API that requires authentication.",
			"inputSchema": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"method": map[string]any{
						"type":        "string",
						"enum":        []string{"GET", "POST", "PUT", "PATCH", "DELETE"},
						"description": "HTTP method",
					},
					"url": map[string]any{
						"type":        "string",
						"description": "Full URL to request",
					},
					"secret_name": map[string]any{
						"type":        "string",
						"description": "Name of the Vaulty secret to use for auth",
					},
					"headers": map[string]any{
						"type":        "object",
						"description": "Additional headers (auth header is injected by Vaulty)",
					},
					"body": map[string]any{
						"type":        "string",
						"description": "Request body",
					},
				},
				"required": []string{"method", "url", "secret_name"},
			},
		},
		{
			"name":        "vaulty_exec",
			"description": "Execute a shell command with secrets injected as environment variables. Output is redacted to prevent secret leakage. Use this instead of running commands that need credentials directly.",
			"inputSchema": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"command": map[string]any{
						"type":        "string",
						"description": "Shell command to execute",
					},
					"secrets": map[string]any{
						"type":        "array",
						"items":       map[string]any{"type": "string"},
						"description": "List of Vaulty secret names to inject as env vars",
					},
					"working_dir": map[string]any{
						"type":        "string",
						"description": "Working directory (optional)",
					},
				},
				"required": []string{"command", "secrets"},
			},
		},
		{
			"name":        "vaulty_list",
			"description": "List available secrets and their access policies. Returns secret names, allowed domains, and allowed commands — never the actual secret values.",
			"inputSchema": map[string]any{
				"type":       "object",
				"properties": map[string]any{},
			},
		},
		{
			"name":        "vaulty_approve",
			"description": "Approve or deny a pending secret injection request. When Vaulty requires approval before using a secret, use this tool to proceed or reject the action.",
			"inputSchema": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"approval_id": map[string]any{
						"type":        "string",
						"description": "The approval ID returned by the pending request",
					},
					"decision": map[string]any{
						"type":        "string",
						"enum":        []string{"approve", "deny"},
						"description": "Whether to approve or deny the secret injection",
					},
				},
				"required": []string{"approval_id", "decision"},
			},
		},
		{
			"name":        "vaulty_pending",
			"description": "List pending approval requests that are waiting for a decision.",
			"inputSchema": map[string]any{
				"type":       "object",
				"properties": map[string]any{},
			},
		},
		{
			"name":        "vaulty_list_services",
			"description": "List configured services and their access policies. Shows domains, injection modes, and command restrictions — never secret values.",
			"inputSchema": map[string]any{
				"type":       "object",
				"properties": map[string]any{},
			},
		},
		{
			"name":        "vaulty_check_access",
			"description": "Check if a URL is accessible through Vaulty and which secret would be used. Useful for discovering which credentials to use before making a request.",
			"inputSchema": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"url": map[string]any{
						"type":        "string",
						"description": "URL to check access for",
					},
				},
				"required": []string{"url"},
			},
		},
		{
			"name":        "vaulty_secret_metadata",
			"description": "Get metadata about stored secrets including names, descriptions, and policies. Never returns actual secret values.",
			"inputSchema": map[string]any{
				"type":       "object",
				"properties": map[string]any{},
			},
		},
	}
}
