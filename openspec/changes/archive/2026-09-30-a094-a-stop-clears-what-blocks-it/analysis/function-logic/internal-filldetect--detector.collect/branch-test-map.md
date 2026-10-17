# Branch Test Map: `Detector.collect`

- Source: `internal/filldetect/detect.go`

> **「진입 실측」은 측정값이다** — 패키지 시험 전체를 `-covermode=set`으로 돌린 프로파일에서 그 분기가 만든 블록의 count다. 어떤 **개별** 시험이 그 분기를 밟는지는 이 실행이 답하지 않는다. 따라서 「Test」 열은 **a094가 요구하는 시험**이며 현존 증명이 아니다.

| Branch | 조건 | 진입 실측 | Test (a094 요구) | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `:365` `if err != nil {` | 예 | **a094 3.E1** — 스냅샷을 얻지 못하면 청소가 `clear=false`로 떨어진다 | no | no |
| B2 | `:371` `if err != nil {` | 아니오 | 기존 | no | no |
| B3 | `:375` `if accountRef == "" {` | 아니오 | 기존 | no | no |
| B4 | `:382` `for _, order := range tracked {` | 예 | 기존 | no | no |
| B5 | `:384` `if keyErr != nil {` | 아니오 | 기존 | no | no |
| B6 | `:387` `if _, duplicate := trackedByKey[key]; duplicate {` | 예 | 기존 | no | no |
| B7 | `:399` `for _, raw := range raws {` | 예 | 기존 | no | no |
| B8 | `:401` `if err != nil {` | 예 | 기존 | no | no |
| B9 | `:409` `if complete {` | 예 | 기존 | no | no |
| B10 | `:414` `} else if len(trackedByID[snap.OrderID]) > 0 {` | 예 | 기존 | no | no |
| B11 | `:410` `if order, locallyTracked := trackedByKey[key]; locallyTracked {` | 예 | 기존 | no | no |
| B12 | `:414` `} else if len(trackedByID[snap.OrderID]) > 0 {` | 예 | 기존 | no | no |
| B13 | `:423` `for _, key := range trackedKeys {` | 예 | 기존 | no | no |
| B14 | `:424` `if seen[key] {` | 예 | 기존 | no | no |
| B15 | `:429` `if err != nil {` | 아니오 | 기존 | no | no |
| B16 | `:434` `if err != nil {` | 아니오 | 기존 | no | no |
| B17 | `:440` `if strings.TrimSpace(snap.OrderID) == "" {` | 예 | 기존 | no | no |
| B18 | `:445` `if !complete {` | 아니오 | 기존 | no | no |
| B19 | `:449` `if actual != key {` | 예 | 기존 | no | no |

**미진입 분기 6개**: B2, B3, B5, B15, B16, B18
**자체 블록 없는 분기 0개**: 없음 — 컴파일러가 별도 블록을 만들지 않는 형태이며 미커버와 다르다.
