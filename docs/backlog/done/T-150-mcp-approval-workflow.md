# T-150: MCP Approval Workflow

**Epic:** 15 — MCP Enhancements
**Status:** done
**Priority:** P1

## Description

Add an interactive approval step before secret injection when invoked via MCP. Agents must get explicit approval before Vaulty injects secrets into requests or commands, unless the secret's policy has `auto_approve = true`.

## Acceptance Criteria

- [x] PendingApproval struct with ID, SecretName, Target, Action, Status, CreatedAt, ExpiresAt
- [x] ApprovalStore with Create, Get, Approve, Deny, Cleanup, ListPending
- [x] AutoApprove field on SecretPolicy (TOML: auto_approve)
- [x] handleRequest and handleExec create pending approvals when auto_approve is false
- [x] vaulty_approve tool to approve or deny pending requests
- [x] vaulty_pending tool to list pending approvals
- [x] LogApproval method on audit logger
- [x] Tests for approval create, approve, deny, expire, list pending, auto-approve bypass
