- source_spec: `_bmad-output/implementation-artifacts/spec-fix-tui-first-use-ssh-host-key-trust.md`
  summary: Support first-use host-key onboarding for ProxyJump bastions during the non-interactive TUI preflight.
  evidence: OpenSSH launches the jump host as a separate SSH process that does not inherit the target preflight's `StrictHostKeyChecking=accept-new` argument, so an unknown bastion still cannot be trusted by this direct-host fix.
- source_spec: `_bmad-output/implementation-artifacts/spec-fix-tui-first-use-ssh-host-key-trust.md`
  summary: Define compatibility behavior for systems whose SSH client does not support `StrictHostKeyChecking=accept-new`.
  evidence: The implemented policy requires a sufficiently recent OpenSSH client; the reported macOS OpenSSH 9.9 client supports it, but cross-platform support requires an explicit minimum-version or fallback design.
