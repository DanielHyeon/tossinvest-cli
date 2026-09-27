# Branch Test Map: `Ingestor.IngestExternalPositions`

- Source: `internal/reconcile/external.go`

> **「진입 실측」은 측정값이다** — 패키지 시험 전체를 `-covermode=set`으로 돌린 프로파일에서 그 분기가 만든 블록의 count다. 어떤 **개별** 시험이 그 분기를 밟는지는 이 실행이 답하지 않는다. 따라서 「Test」 열은 **a095 3판이 요구하는 시험**이며 현존 증명이 아니다. `[비움 — Qn]`은 사용자 결정이 덮지 않아 Manager 에게 올린 질문이다.

| Branch | 조건 | 진입 실측 | Test (a095 요구) | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `:185` `if len(diff.ExternalPos) == 0 {` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B2 | `:188` `if in == nil \|\| in.Journal == nil {` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B3 | `:193` `if account == "" {` | 아니오 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B4 | `:197` `if asOf == "" {` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B5 | `:204` `for _, external := range diff.ExternalPos {` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B6 | `:207` `if market == "" {` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B7 | `:210` `if symbol == "" \|\| market == "" {` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B8 | `:218` `if err != nil {` | 아니오 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B9 | `:240` `if errors.Is(err, journal.ErrAdjustmentStale) {` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B10 | `:245` `if err != nil {` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B11 | `:258` `if strings.TrimSpace(result.Position.EntryDecisionID) != "" {` | 아니오 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B12 | `:278` `if !folded.Applied \|\| in.Alert == nil {` | 예 | **a095 2.9** — 생산 배선에서 알림 어댑터가 nil임을 고정 | no | no |
| B13 | `:281` `if err := in.Alert.ExternalPositionFound(ctx, ExternalPositionAlert{` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |

**미진입 분기 3개**: B3, B8, B11
**자체 블록 없는 분기 0개**: 없음 — 컴파일러가 별도 블록을 만들지 않는 형태(빈 `switch {` 등)이며 미커버와 다르다.
