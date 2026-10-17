# Branch Test Map: `newEngineRunCmd`

- Source: `cmd/tossctl/engine.go` (134-176) — 편집 뒤

| Branch | Anchor | Scenario | Test | a131 |
|---|---|---|---|---|
| B1 | branchless happy path at 134:1 | 본문 | `TestA131TheSinksCallTheRule` (구조 핀) · `TestA131WriterScrubsEveryWrite` | 편집 |

## a131 시험

- `TestA131TheSinksCallTheRule` — newEngineRunCmd 가 redact.Writer 를 두 번 부름(go/parser)
