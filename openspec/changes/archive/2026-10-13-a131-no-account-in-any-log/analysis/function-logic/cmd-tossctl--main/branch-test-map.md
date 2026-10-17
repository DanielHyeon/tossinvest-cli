# Branch Test Map: `main`

- Source: `cmd/tossctl/main.go` (11-22) — 편집 뒤

| Branch | Anchor | Scenario | Test | a131 |
|---|---|---|---|---|
| B1 | if at 18:2 | `if err := root.Execute(); err != nil {` | `TestA131TheSinksCallTheRule` (구조 핀 — SetErr · redact.Writer · os.Stderr 직접 Fprint 없음) · `TestA131CobraErrorLineIsRedacted` | 편집 영향 |

## a131 시험

- `TestA131TheSinksCallTheRule`
- `TestA131CobraErrorLineIsRedacted`
