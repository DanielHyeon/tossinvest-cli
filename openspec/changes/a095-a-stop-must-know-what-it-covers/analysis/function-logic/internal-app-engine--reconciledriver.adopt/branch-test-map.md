# Branch Test Map: `ReconcileDriver.adopt`

- Source: `internal/app/engine/adoption.go`

> **「진입 실측」은 측정값이다** — 패키지 시험 전체를 `-covermode=set`으로 돌린 프로파일에서 그 분기가 만든 블록의 count다. 어떤 **개별** 시험이 그 분기를 밟는지는 이 실행이 답하지 않는다. 따라서 「Test」 열은 **a095 3판이 요구하는 시험**이며 현존 증명이 아니다. `[비움 — Qn]`은 사용자 결정이 덮지 않아 Manager 에게 올린 질문이다.

| Branch | 조건 | 진입 실측 | Test (a095 요구) | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `:175` `if len(candidates) == 0 {` | 예 | 이 분기 자체는 무변화 — 함수의 **결과 형태**만 편집 경계 안(r4 R4-1, 6판 r5 R5-4) | no | no |
| B2 | `:180` `if err != nil {` | 아니오 | **a095 2.6** — [비움 — Q2(c)] 연기된 후보의 등급 | no | no |
| B3 | `:182` `if cycle.Err == nil {` | 아니오 | 이 분기 자체는 무변화 — 함수의 **결과 형태**만 편집 경계 안(r4 R4-1, 6판 r5 R5-4) | no | no |
| B4 | `:189` `if bound <= 0 {` | 예 | 이 분기 자체는 무변화 — 함수의 **결과 형태**만 편집 경계 안(r4 R4-1, 6판 r5 R5-4) | no | no |
| B5 | `:192` `for _, c := range candidates {` | 예 | 이 분기 자체는 무변화 — 함수의 **결과 형태**만 편집 경계 안(r4 R4-1, 6판 r5 R5-4) | no | no |
| B6 | `:195` `if !ok {` | 예 | **a095 2.6** — [비움 — Q2(c)] 연기된 후보의 등급 | no | no |
| B7 | `:201` `if age := d.clk.Now().Sub(readAt); age > bound {` | 예 | **a095 2.6** — [비움 — Q2(c)] 연기된 후보의 등급 | no | no |
| B8 | `:210` `if d.adoptOne(ctx, c, observed) {` | 예 | 이 분기 자체는 무변화 — 함수의 **결과 형태**만 편집 경계 안(r4 R4-1, 6판 r5 R5-4) | no | no |

**미진입 분기 2개**: B2, B3
**자체 블록 없는 분기 0개**: 없음 — 컴파일러가 별도 블록을 만들지 않는 형태(빈 `switch {` 등)이며 미커버와 다르다.
