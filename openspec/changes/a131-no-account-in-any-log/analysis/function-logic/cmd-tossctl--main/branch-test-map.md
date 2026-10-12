# Branch Test Map: `main`

- Source: `cmd/tossctl/main.go` (10-17) — 편집 전(base)

| Branch | Anchor | Scenario | Test | a131 |
|---|---|---|---|---|
| B1 | if at 13:2 | `if err := newRootCmd().Execute(); err != nil {` | `TestA131TheSinksCallTheRule` (구조 핀 — redact.Writer 호출 · os.Stderr 직접 Fprint 없음) | 편집 영향 |

## a131 시험

- `TestA131TheSinksCallTheRule`
