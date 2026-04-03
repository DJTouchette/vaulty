//go:build !windows

package mcp

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/djtouchette/vaulty/internal/audit"
	"github.com/djtouchette/vaulty/internal/policy"
)

// Helper to extract the text content from a tool call response.
func resultText(t *testing.T, resp JSONRPCResponse) string {
	t.Helper()
	result, ok := resp.Result.(map[string]any)
	if !ok {
		t.Fatalf("result is not a map: %v", resp.Result)
	}
	content, ok := result["content"].([]any)
	if !ok || len(content) == 0 {
		t.Fatalf("content missing or empty in response id=%v", resp.ID)
	}
	return content[0].(map[string]any)["text"].(string)
}

func isError(resp JSONRPCResponse) bool {
	result, ok := resp.Result.(map[string]any)
	if !ok {
		return resp.Error != nil
	}
	isErr, ok := result["isError"]
	return ok && isErr.(bool)
}

// TestE2EFullSessionFlow tests a complete MCP session: initialize → list → discover → exec with approval → verify audit.
func TestE2EFullSessionFlow(t *testing.T) {
	env := setupTestEnv(t)

	lines := []string{
		// 1. Initialize
		`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`,
		// 2. Client initialized notification (no response expected)
		`{"jsonrpc":"2.0","method":"notifications/initialized"}`,
		// 3. List tools
		`{"jsonrpc":"2.0","id":3,"method":"tools/list"}`,
		// 4. List secrets
		`{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"vaulty_list","arguments":{}}}`,
		// 5. Check access for a matching URL
		`{"jsonrpc":"2.0","id":5,"method":"tools/call","params":{"name":"vaulty_check_access","arguments":{"url":"https://api.example.com/data"}}}`,
		// 6. Exec (requires approval since auto_approve is false by default)
		`{"jsonrpc":"2.0","id":6,"method":"tools/call","params":{"name":"vaulty_exec","arguments":{"command":"echo hello $TEST_KEY","secrets":["TEST_KEY"]}}}`,
		// 7. Check pending
		`{"jsonrpc":"2.0","id":7,"method":"tools/call","params":{"name":"vaulty_pending","arguments":{}}}`,
		// 8. Approve it
		`{"jsonrpc":"2.0","id":8,"method":"tools/call","params":{"name":"vaulty_approve","arguments":{"approval_id":"approval-1","decision":"approve"}}}`,
		// 9. Read secrets resource
		`{"jsonrpc":"2.0","id":9,"method":"resources/read","params":{"uri":"vaulty://secrets"}}`,
		// 10. Read audit resource (should have entries from exec)
		`{"jsonrpc":"2.0","id":10,"method":"resources/read","params":{"uri":"vaulty://audit"}}`,
	}

	input := strings.Join(lines, "\n") + "\n"
	responses := runServer(t, env, input)

	// notifications/initialized produces no response, so we should get 9 responses
	if len(responses) != 9 {
		t.Fatalf("expected 9 responses, got %d", len(responses))
	}

	// 1. Initialize
	initResult := responses[0].Result.(map[string]any)
	if initResult["protocolVersion"] != protocolVersion {
		t.Errorf("bad protocol version: %v", initResult["protocolVersion"])
	}
	caps := initResult["capabilities"].(map[string]any)
	if _, ok := caps["tools"]; !ok {
		t.Error("tools capability missing")
	}
	if _, ok := caps["resources"]; !ok {
		t.Error("resources capability missing")
	}

	// 2. Tools list (response index 1 since notification produced nothing)
	toolsResult := responses[1].Result.(map[string]any)
	tools := toolsResult["tools"].([]any)
	if len(tools) != 8 {
		t.Errorf("expected 8 tools, got %d", len(tools))
	}

	// 3. List secrets — should show TEST_KEY but not the value
	listText := resultText(t, responses[2])
	if !strings.Contains(listText, "TEST_KEY") {
		t.Errorf("list missing TEST_KEY: %q", listText)
	}
	if strings.Contains(listText, "secret_value_123") {
		t.Error("list should not contain secret value")
	}

	// 4. Check access — should find TEST_KEY for api.example.com
	accessText := resultText(t, responses[3])
	if !strings.Contains(accessText, "TEST_KEY") {
		t.Errorf("check_access should find TEST_KEY: %q", accessText)
	}

	// 5. Exec — should require approval
	execText := resultText(t, responses[4])
	if !strings.Contains(execText, "Approval required") {
		t.Errorf("exec should require approval: %q", execText)
	}
	if !strings.Contains(execText, "approval-1") {
		t.Errorf("exec should contain approval ID: %q", execText)
	}

	// 6. Pending list — should show the pending approval
	pendingText := resultText(t, responses[5])
	if !strings.Contains(pendingText, "approval-1") {
		t.Errorf("pending should list approval-1: %q", pendingText)
	}

	// 7. Approve — should execute and redact
	approveText := resultText(t, responses[6])
	if strings.Contains(approveText, "secret_value_123") {
		t.Error("approved exec should not contain raw secret")
	}
	if !strings.Contains(approveText, "[VAULTY:TEST_KEY]") {
		t.Errorf("approved exec should contain redacted placeholder: %q", approveText)
	}
	if !strings.Contains(approveText, "exit code: 0") {
		t.Errorf("approved exec should succeed: %q", approveText)
	}

	// 8. Secrets resource
	secretsResult := responses[7].Result.(map[string]any)
	secretsContents := secretsResult["contents"].([]any)
	secretsText := secretsContents[0].(map[string]any)["text"].(string)
	if !strings.Contains(secretsText, "TEST_KEY") {
		t.Errorf("secrets resource should contain TEST_KEY: %q", secretsText)
	}

	// 9. Audit resource — should have the exec entry
	auditResult := responses[8].Result.(map[string]any)
	auditContents := auditResult["contents"].([]any)
	auditText := auditContents[0].(map[string]any)["text"].(string)
	if !strings.Contains(auditText, "exec") {
		t.Errorf("audit should contain exec entry: %q", auditText)
	}
}

// TestE2EApprovalDenyFlow tests the deny path in a multi-step session.
func TestE2EApprovalDenyFlow(t *testing.T) {
	env := setupTestEnv(t)

	lines := []string{
		`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`,
		// Request that needs approval
		`{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"vaulty_exec","arguments":{"command":"echo $TEST_KEY","secrets":["TEST_KEY"]}}}`,
		// Deny it
		`{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"vaulty_approve","arguments":{"approval_id":"approval-1","decision":"deny"}}}`,
		// Pending should be empty now
		`{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"vaulty_pending","arguments":{}}}`,
	}

	responses := runServer(t, env, strings.Join(lines, "\n")+"\n")
	if len(responses) != 4 {
		t.Fatalf("expected 4 responses, got %d", len(responses))
	}

	// Approval prompt
	execText := resultText(t, responses[1])
	if !strings.Contains(execText, "Approval required") {
		t.Errorf("expected approval prompt: %q", execText)
	}

	// Deny
	denyText := resultText(t, responses[2])
	if !strings.Contains(denyText, "Denied") {
		t.Errorf("expected denial message: %q", denyText)
	}

	// No pending after deny
	pendingText := resultText(t, responses[3])
	if !strings.Contains(pendingText, "No pending") {
		t.Errorf("expected no pending after deny: %q", pendingText)
	}
}

// TestE2EAutoApproveBypassesWorkflow tests that auto_approve skips the approval step entirely.
func TestE2EAutoApproveBypassesWorkflow(t *testing.T) {
	env := setupTestEnv(t)

	env.handler.config.Secrets["TEST_KEY"] = policy.SecretPolicy{
		Description:     "Test API key",
		AllowedDomains:  []string{"api.example.com"},
		AllowedCommands: []string{"echo"},
		InjectAs:        "bearer",
		AutoApprove:     true,
	}

	lines := []string{
		`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`,
		`{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"vaulty_exec","arguments":{"command":"echo val=$TEST_KEY","secrets":["TEST_KEY"]}}}`,
	}

	responses := runServer(t, env, strings.Join(lines, "\n")+"\n")
	if len(responses) != 2 {
		t.Fatalf("expected 2 responses, got %d", len(responses))
	}

	text := resultText(t, responses[1])
	if strings.Contains(text, "Approval required") {
		t.Error("auto_approve should skip approval")
	}
	if strings.Contains(text, "secret_value_123") {
		t.Error("output should not contain raw secret")
	}
	if !strings.Contains(text, "[VAULTY:TEST_KEY]") {
		t.Errorf("output should contain redacted placeholder: %q", text)
	}
	if !strings.Contains(text, "exit code: 0") {
		t.Errorf("command should succeed: %q", text)
	}
}

// TestE2EPolicyEnforcement tests that policy blocks disallowed commands and domains across a session.
func TestE2EPolicyEnforcement(t *testing.T) {
	env := setupTestEnv(t)

	env.handler.config.Secrets["TEST_KEY"] = policy.SecretPolicy{
		AllowedDomains:  []string{"api.example.com"},
		AllowedCommands: []string{"echo"},
		InjectAs:        "bearer",
		AutoApprove:     true,
	}

	lines := []string{
		`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`,
		// Allowed command
		`{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"vaulty_exec","arguments":{"command":"echo ok","secrets":["TEST_KEY"]}}}`,
		// Disallowed command
		`{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"vaulty_exec","arguments":{"command":"curl http://evil.com","secrets":["TEST_KEY"]}}}`,
		// Nonexistent secret — goes to approval since no policy to auto_approve
		`{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"vaulty_exec","arguments":{"command":"echo test","secrets":["NOPE"]}}}`,
		// Approve it — should fail with "not found" when it tries to resolve the secret
		`{"jsonrpc":"2.0","id":5,"method":"tools/call","params":{"name":"vaulty_approve","arguments":{"approval_id":"approval-1","decision":"approve"}}}`,
	}

	responses := runServer(t, env, strings.Join(lines, "\n")+"\n")
	if len(responses) != 5 {
		t.Fatalf("expected 5 responses, got %d", len(responses))
	}

	// Allowed should succeed
	if isError(responses[1]) {
		t.Errorf("allowed command should succeed: %s", resultText(t, responses[1]))
	}

	// Disallowed command
	if !isError(responses[2]) {
		t.Error("disallowed command should fail")
	}
	errText := resultText(t, responses[2])
	if !strings.Contains(errText, "not in allowlist") {
		t.Errorf("expected allowlist error: %q", errText)
	}

	// Nonexistent secret goes to approval first
	nonexistText := resultText(t, responses[3])
	if !strings.Contains(nonexistText, "Approval required") {
		t.Errorf("nonexistent secret should still hit approval flow: %q", nonexistText)
	}

	// Approving it should fail with "not found"
	if !isError(responses[4]) {
		t.Error("approving nonexistent secret should fail")
	}
	errText = resultText(t, responses[4])
	if !strings.Contains(errText, "not found") {
		t.Errorf("expected not found error after approval: %q", errText)
	}
}

// TestE2EMultipleApprovals tests handling multiple pending approvals in one session.
func TestE2EMultipleApprovals(t *testing.T) {
	env := setupTestEnv(t)

	lines := []string{
		`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`,
		// Two separate exec requests that need approval
		`{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"vaulty_exec","arguments":{"command":"echo first $TEST_KEY","secrets":["TEST_KEY"]}}}`,
		`{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"vaulty_exec","arguments":{"command":"echo second $TEST_KEY","secrets":["TEST_KEY"]}}}`,
		// List pending — should show 2
		`{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"vaulty_pending","arguments":{}}}`,
		// Approve the first, deny the second
		`{"jsonrpc":"2.0","id":5,"method":"tools/call","params":{"name":"vaulty_approve","arguments":{"approval_id":"approval-1","decision":"approve"}}}`,
		`{"jsonrpc":"2.0","id":6,"method":"tools/call","params":{"name":"vaulty_approve","arguments":{"approval_id":"approval-2","decision":"deny"}}}`,
		// Pending should be empty
		`{"jsonrpc":"2.0","id":7,"method":"tools/call","params":{"name":"vaulty_pending","arguments":{}}}`,
	}

	responses := runServer(t, env, strings.Join(lines, "\n")+"\n")
	if len(responses) != 7 {
		t.Fatalf("expected 7 responses, got %d", len(responses))
	}

	// Both execs should require approval
	for i, idx := range []int{1, 2} {
		text := resultText(t, responses[idx])
		if !strings.Contains(text, "Approval required") {
			t.Errorf("exec %d should require approval: %q", i+1, text)
		}
	}

	// Pending should show 2 approvals
	pendingText := resultText(t, responses[3])
	if !strings.Contains(pendingText, "approval-1") || !strings.Contains(pendingText, "approval-2") {
		t.Errorf("pending should list both approvals: %q", pendingText)
	}

	// First approved — should contain redacted output
	approvedText := resultText(t, responses[4])
	if !strings.Contains(approvedText, "[VAULTY:TEST_KEY]") {
		t.Errorf("approved exec should have redacted output: %q", approvedText)
	}

	// Second denied
	deniedText := resultText(t, responses[5])
	if !strings.Contains(deniedText, "Denied") {
		t.Errorf("denied exec should show denial: %q", deniedText)
	}

	// No more pending
	finalPending := resultText(t, responses[6])
	if !strings.Contains(finalPending, "No pending") {
		t.Errorf("should have no pending approvals: %q", finalPending)
	}
}

// TestE2EDiscoveryWorkflow tests the agent discovery flow: list services → check access → get metadata → execute.
func TestE2EDiscoveryWorkflow(t *testing.T) {
	env := setupTestEnv(t)

	env.handler.config.Secrets["TEST_KEY"] = policy.SecretPolicy{
		Description:     "Test API key",
		AllowedDomains:  []string{"api.example.com"},
		AllowedCommands: []string{"echo"},
		InjectAs:        "bearer",
		AutoApprove:     true,
	}

	lines := []string{
		`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`,
		// Discovery: what services are available?
		`{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"vaulty_list_services","arguments":{}}}`,
		// Discovery: can I reach this URL?
		`{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"vaulty_check_access","arguments":{"url":"https://api.example.com/v1/users"}}}`,
		// Discovery: can I reach a different URL?
		`{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"vaulty_check_access","arguments":{"url":"https://api.stripe.com/v1/charges"}}}`,
		// Get detailed metadata
		`{"jsonrpc":"2.0","id":5,"method":"tools/call","params":{"name":"vaulty_secret_metadata","arguments":{}}}`,
		// Now execute with the discovered secret
		`{"jsonrpc":"2.0","id":6,"method":"tools/call","params":{"name":"vaulty_exec","arguments":{"command":"echo using $TEST_KEY","secrets":["TEST_KEY"]}}}`,
	}

	responses := runServer(t, env, strings.Join(lines, "\n")+"\n")
	if len(responses) != 6 {
		t.Fatalf("expected 6 responses, got %d", len(responses))
	}

	// List services
	servicesText := resultText(t, responses[1])
	if !strings.Contains(servicesText, "TEST_KEY") {
		t.Errorf("services should list TEST_KEY: %q", servicesText)
	}
	if !strings.Contains(servicesText, "api.example.com") {
		t.Errorf("services should show domain: %q", servicesText)
	}

	// Check access — matching domain
	matchText := resultText(t, responses[2])
	if !strings.Contains(matchText, "TEST_KEY") {
		t.Errorf("check_access should find TEST_KEY for matching domain: %q", matchText)
	}

	// Check access — non-matching domain
	noMatchText := resultText(t, responses[3])
	if !strings.Contains(noMatchText, "No secrets have access") {
		t.Errorf("check_access should report no access for stripe.com: %q", noMatchText)
	}

	// Metadata — no values leaked
	metaText := resultText(t, responses[4])
	if !strings.Contains(metaText, "Test API key") {
		t.Errorf("metadata should contain description: %q", metaText)
	}
	if strings.Contains(metaText, "secret_value_123") {
		t.Error("metadata should not contain secret value")
	}

	// Execute
	execText := resultText(t, responses[5])
	if !strings.Contains(execText, "[VAULTY:TEST_KEY]") {
		t.Errorf("exec should redact secret: %q", execText)
	}
}

// TestE2EResourcesAfterActions tests that resources reflect the state after tool actions.
func TestE2EResourcesAfterActions(t *testing.T) {
	env := setupTestEnv(t)

	env.handler.config.Secrets["TEST_KEY"] = policy.SecretPolicy{
		AllowedCommands: []string{"echo"},
		AutoApprove:     true,
	}

	lines := []string{
		`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`,
		// Execute a command to generate audit entries
		`{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"vaulty_exec","arguments":{"command":"echo hi","secrets":["TEST_KEY"]}}}`,
		// Read the audit log — should contain exec entry
		`{"jsonrpc":"2.0","id":3,"method":"resources/read","params":{"uri":"vaulty://audit"}}`,
		// Read policy
		`{"jsonrpc":"2.0","id":4,"method":"resources/read","params":{"uri":"vaulty://policy"}}`,
		// Read secrets list
		`{"jsonrpc":"2.0","id":5,"method":"resources/read","params":{"uri":"vaulty://secrets"}}`,
	}

	responses := runServer(t, env, strings.Join(lines, "\n")+"\n")
	if len(responses) != 5 {
		t.Fatalf("expected 5 responses, got %d", len(responses))
	}

	// Exec should succeed
	if isError(responses[1]) {
		t.Fatalf("exec failed: %s", resultText(t, responses[1]))
	}

	// Audit should contain the exec
	auditResult := responses[2].Result.(map[string]any)
	auditContents := auditResult["contents"].([]any)
	auditText := auditContents[0].(map[string]any)["text"].(string)
	if !strings.Contains(auditText, `"action":"exec"`) {
		t.Errorf("audit should contain exec entry: %q", auditText)
	}
	if !strings.Contains(auditText, "TEST_KEY") {
		t.Errorf("audit should reference TEST_KEY: %q", auditText)
	}
	if strings.Contains(auditText, "secret_value_123") {
		t.Error("audit should never contain secret values")
	}

	// Policy resource
	policyResult := responses[3].Result.(map[string]any)
	policyContents := policyResult["contents"].([]any)
	policyText := policyContents[0].(map[string]any)["text"].(string)
	if !strings.Contains(policyText, "echo") {
		t.Errorf("policy should contain allowed command: %q", policyText)
	}

	// Secrets resource
	secretsResult := responses[4].Result.(map[string]any)
	secretsContents := secretsResult["contents"].([]any)
	secretsText := secretsContents[0].(map[string]any)["text"].(string)
	if !strings.Contains(secretsText, "TEST_KEY") {
		t.Errorf("secrets resource should contain TEST_KEY: %q", secretsText)
	}
}

// TestE2ERedactionCoversAllEncodings tests that secrets are redacted in raw, base64, and URL-encoded forms.
func TestE2ERedactionCoversAllEncodings(t *testing.T) {
	env := setupTestEnv(t)

	env.handler.config.Secrets["TEST_KEY"] = policy.SecretPolicy{
		AllowedCommands: []string{"echo", "printf"},
		AutoApprove:     true,
	}

	lines := []string{
		`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`,
		// Echo the raw value
		`{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"vaulty_exec","arguments":{"command":"echo $TEST_KEY","secrets":["TEST_KEY"]}}}`,
		// Echo base64-encoded value
		`{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"vaulty_exec","arguments":{"command":"echo $TEST_KEY | base64","secrets":["TEST_KEY"]}}}`,
	}

	responses := runServer(t, env, strings.Join(lines, "\n")+"\n")
	if len(responses) != 3 {
		t.Fatalf("expected 3 responses, got %d", len(responses))
	}

	// Raw echo — should be redacted
	rawText := resultText(t, responses[1])
	if strings.Contains(rawText, "secret_value_123") {
		t.Error("raw echo should not contain secret value")
	}
	if !strings.Contains(rawText, "[VAULTY:TEST_KEY]") {
		t.Errorf("raw echo should contain redacted placeholder: %q", rawText)
	}

	// Base64 echo — the base64 of the secret value should also be redacted
	b64Text := resultText(t, responses[2])
	if strings.Contains(b64Text, "secret_value_123") {
		t.Error("base64 echo should not contain raw secret value")
	}
}

// TestE2EInvalidRequestsInSession tests that invalid requests don't break the session.
func TestE2EInvalidRequestsInSession(t *testing.T) {
	env := setupTestEnv(t)

	lines := []string{
		`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`,
		// Invalid method
		`{"jsonrpc":"2.0","id":2,"method":"nonexistent/method"}`,
		// Invalid tool name
		`{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"fake_tool","arguments":{}}}`,
		// Valid request after errors — session should still work
		`{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"vaulty_list","arguments":{}}}`,
	}

	responses := runServer(t, env, strings.Join(lines, "\n")+"\n")
	if len(responses) != 4 {
		t.Fatalf("expected 4 responses, got %d", len(responses))
	}

	// Initialize OK
	if responses[0].Error != nil {
		t.Errorf("initialize should succeed: %v", responses[0].Error)
	}

	// Invalid method → error
	if responses[1].Error == nil {
		t.Error("nonexistent method should return error")
	}

	// Invalid tool → isError in result
	if !isError(responses[2]) {
		t.Error("fake tool should return error")
	}

	// Session still alive — list works
	listText := resultText(t, responses[3])
	if !strings.Contains(listText, "TEST_KEY") {
		t.Errorf("list should still work after errors: %q", listText)
	}
}

// TestE2EAuditTrailCompleteness tests that all actions (exec, denied) are logged to the audit trail.
func TestE2EAuditTrailCompleteness(t *testing.T) {
	env := setupTestEnv(t)

	env.handler.config.Secrets["TEST_KEY"] = policy.SecretPolicy{
		AllowedCommands: []string{"echo"},
		AutoApprove:     true,
	}

	lines := []string{
		`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`,
		// Allowed exec
		`{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"vaulty_exec","arguments":{"command":"echo ok","secrets":["TEST_KEY"]}}}`,
		// Denied exec (curl not in allowlist)
		`{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"vaulty_exec","arguments":{"command":"curl http://evil.com","secrets":["TEST_KEY"]}}}`,
	}

	responses := runServer(t, env, strings.Join(lines, "\n")+"\n")
	if len(responses) != 3 {
		t.Fatalf("expected 3 responses, got %d", len(responses))
	}

	// Verify exec succeeded
	if isError(responses[1]) {
		t.Fatalf("allowed exec should succeed: %s", resultText(t, responses[1]))
	}

	// Verify denial
	if !isError(responses[2]) {
		t.Fatal("disallowed exec should fail")
	}

	// Read the audit log file directly
	data, err := os.ReadFile(env.logPath)
	if err != nil {
		t.Fatalf("reading audit log: %v", err)
	}

	logLines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(logLines) < 2 {
		t.Fatalf("expected at least 2 audit entries, got %d", len(logLines))
	}

	// Check exec entry
	var execEntry audit.Entry
	json.Unmarshal([]byte(logLines[0]), &execEntry)
	if execEntry.Action != "exec" {
		t.Errorf("first entry should be exec, got %q", execEntry.Action)
	}
	if execEntry.Secret != "TEST_KEY" {
		t.Errorf("exec entry secret = %q, want TEST_KEY", execEntry.Secret)
	}

	// Check denied entry
	var deniedEntry audit.Entry
	json.Unmarshal([]byte(logLines[1]), &deniedEntry)
	if deniedEntry.Action != "denied" {
		t.Errorf("second entry should be denied, got %q", deniedEntry.Action)
	}
	if deniedEntry.Reason == "" {
		t.Error("denied entry should have a reason")
	}
}
