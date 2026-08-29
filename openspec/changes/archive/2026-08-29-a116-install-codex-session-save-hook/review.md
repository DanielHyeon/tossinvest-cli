# Review: a116-install-codex-session-save-hook

## 2026-08-17 Proposal Freeze

- **Review class:** Lightweight tooling change review permitted by `docs/WORKFLOW.md`; no trading runtime or high-risk path is affected.
- **Voices:** Manager CEO, design, engineering, and developer-experience perspectives.
- **User-confirmed premises:** Codex receives a dedicated implementation; the existing Claude configuration and script are untouched; generated storage is separate.
- **Validation:** `openspec validate a116-install-codex-session-save-hook --strict --no-interactive` passed.

### Evidence

- Memory recall returned no prior episode for this hook; project GBrain was busy behind its live MCP owner, so it was treated as advisory-only.
- `make sdd-sync` completed and `codegraph status .` reported the index up to date with 1,964 files and 34,665 nodes.
- CodeGraph resolved the existing observation path to `tools/sdd-history/agent_save_hook.py` and its test import. The current `.codex/hooks.json` has exactly one edit matcher invoking that script.
- CodeGraphContext reported zero indexed files/functions for this repository and supplied no usable supporting evidence. Current HEAD, direct file reads, CodeGraph, and tests remain authoritative.
- Official OpenAI documentation confirms project-local `.codex/hooks.json`, concurrent matching hooks, `PostToolUse` coverage for `Bash` and `apply_patch`, hook JSON on stdin, `transcript_path`, async handlers, Git-root path resolution, and explicit hook trust.

### Four-Perspective Findings

| Perspective | Finding | Decision |
|---|---|---|
| CEO | The operator's requested outcome is narrow: reliable handoff without cross-agent contamination. | Accept a separate saver and store; reject a shared-script refactor. |
| Design | A single stable latest file plus five timestamped backups is easier to resume from than per-tool files. | Keep `.codex-context/session-summary.md` as the entry point and bound backups. |
| Engineering | Async hooks may overlap and the transcript format is unstable. | Require a non-blocking lock, atomic replace, defensive parsing, and fail-open behavior. |
| DX | Project hooks do not execute until the exact definition is trusted. | Preserve `/hooks` as an explicit final manual activation step rather than bypassing trust. |

### Scope and Safety Decisions

- No third-party dependency is introduced.
- No network call, account state, trading command, or live-order side effect exists.
- Tool payloads, tool responses, reasoning, developer instructions, and transcript paths are excluded from the summary.
- Recognized credential values in retained user/assistant text are redacted.
- Existing `.codex/hooks.json` observation behavior remains present; the new saver is an additional handler.

### Function Logic Map

Function Logic Map: not-applicable — a116 adds a new Python leaf script, tests, JSON hook configuration, documentation, and an ignore rule. It does not modify an existing Go function or any high-risk function.

### Proposal-Freeze Verdict

**ACCEPTED.** The contract is implementation-ready. RED tests must demonstrate storage isolation and preservation of both Claude-owned files before the new hook is installed.

## 2026-08-17 Implementation Verification

### RED / GREEN Evidence

- **RED:** `python3 -m unittest tools/sdd-history/test_codex_session_save.py -v` ran 7 tests and failed all 7 before the saver, matcher, and ignore rule existed.
- **GREEN:** The same command ran 7 tests and passed all 7 after implementation.
- **Regression:** `python3 -m unittest discover -s tools/sdd-history -p 'test_*.py' -v` ran 22 tests and passed all 22, including the 15 pre-existing SDD history tests.
- **Syntax/config:** `python3 -m py_compile .codex/hooks/save_session.py tools/sdd-history/test_codex_session_save.py` and `python3 -m json.tool .codex/hooks.json` passed.
- **Synthetic hook:** A synthetic `PostToolUse` envelope produced `.codex-context/session-summary.md` with the expected session/model metadata; `git check-ignore -v` resolved the file to the root `.codex-context/` ignore rule.

### Isolation Evidence

- Before and after a116 implementation, `.claude/settings.json` SHA-256 remained `41e995815c0e72d2c240fa71dac8f40e30e00350e047c56619914c6db3b1b765`.
- Before and after a116 implementation, root `save-session.sh` SHA-256 remained `557d21ca136190c7473d85a58e465ebe8c430b1a5c914488b92e1111e20575be`.
- Tests reject a transcript under `.claude/`, preserve Claude fixture bytes, and verify no Claude marker reaches the Codex summary.
- The generated store is local and ignored. Activation still requires the operator to review and trust the exact project hook definition using `/hooks`; this change does not bypass Codex hook trust.

### Security Hardening and Independent Review

- A follow-up RED run exposed two redaction gaps: prefixed environment names such as `OPENAI_API_KEY` and quoted JSON/YAML keys such as `"api_key"`. Regression cases were added before the matcher was expanded; the focused 7-test suite and the 22-test SDD-history suite then passed.
- The saver requests `0700` for `.codex-context/` and `backups/`, and `0600` for latest, backup, temporary, and lock files. Those modes are asserted on a native Linux temporary filesystem.
- The actual repository is mounted at `/mnt/D` as `fuseblk` without POSIX metadata semantics, so both `.codex-context/` and the pre-existing Claude `.ai-context/` surface as `0755` even after explicit `chmod`. The store/path isolation contract remains satisfied; confidentiality between OS users must instead be enforced with Windows ACLs or a metadata-enabled Linux filesystem.
- Independent review reproduced the fixed redaction cases, verified the DrvFs limitation, and reported no remaining correctness blocker.
