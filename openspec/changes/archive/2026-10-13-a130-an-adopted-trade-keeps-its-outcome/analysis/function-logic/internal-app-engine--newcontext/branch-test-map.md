# Branch Test Map: `NewContext`

편집 뒤 좌표(4판 GREEN). 분기별 RED 는 따로 재지 않았음(no) — 4판 시험 전체가 base 에서 컴파일 RED(`analysis/red/red-4.txt`), 조작 행 모양은 행동 RED(red-2 · red-4). GREEN = 패키지 시험 통과(`analysis/verify/receipts.md`). Test 칸은 그 분기 몸체를 실행한 시험(`analysis/verify/btm-coverage.md` 측정, 시험 하나씩 커버리지).

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:439` `if err != nil {` | 패키지 전체 시험이 실행(후보 패턴 밖 — 개별 이름 미측정) | no | yes |
| B2 | `:446` `if err != nil {` | not-applicable: 오류 반환 — SQLite 읽기 · 쓰기 · 커밋 실패(또는 열기 실패)에서만 열림; 건강한 원장에서 도달 불가이고 결함 주입 시험 없음. a130 무편집 갈래(편집은 이 함수의 다른 줄) | no | n/a |
| B3 | `:450` `if clk == nil {` | 패키지 전체 시험이 실행(후보 패턴 밖 — 개별 이름 미측정) | no | yes |
| B4 | `:464` `if opts.Publisher != nil {` | 패키지 전체 시험이 실행(후보 패턴 밖 — 개별 이름 미측정) | no | yes |
| B5 | `:472` `if err := recordGateSettings(auditLog, gate, cfg.Engine.Adoption, notifications,` | not-applicable: 오류 반환 — SQLite 읽기 · 쓰기 · 커밋 실패(또는 열기 실패)에서만 열림; 건강한 원장에서 도달 불가이고 결함 주입 시험 없음. a130 무편집 갈래(편집은 이 함수의 다른 줄) | no | n/a |
| B6 | `:487` `if err != nil {` | 패키지 전체 시험이 실행(후보 패턴 밖 — 개별 이름 미측정) | no | yes |
| B7 | `:494` `if err != nil {` | 패키지 전체 시험이 실행(후보 패턴 밖 — 개별 이름 미측정) | no | yes |
| B8 | `:503` `if err := bindApplyHooks(jrn, tradeOutcomeRefusalSink(opts.Logger)); err != nil ` | not-applicable: bindApplyHooks 실패 — 새로 연 원장에 훅을 한 번 다는 자리라 「이미 달림」 거절이 생산에서 열리지 않음; a130 은 인자(싱크)만 더함, 거절 자체는 bindApplyHooks B1 · B2 가 잼 | no | n/a |
| B9 | `:522` `if err != nil {` | not-applicable: 오류 반환 — SQLite 읽기 · 쓰기 · 커밋 실패(또는 열기 실패)에서만 열림; 건강한 원장에서 도달 불가이고 결함 주입 시험 없음. a130 무편집 갈래(편집은 이 함수의 다른 줄) | no | n/a |
| B10 | `:537` `if gate.Enabled && guardian == nil && !opts.disableProductionGuardian {` | 패키지 전체 시험이 실행(후보 패턴 밖 — 개별 이름 미측정) | no | yes |
| B11 | `:539` `if factory == nil {` | 패키지 전체 시험이 실행(후보 패턴 밖 — 개별 이름 미측정) | no | yes |
| B12 | `:545` `if err != nil {` | 패키지 전체 시험이 실행(후보 패턴 밖 — 개별 이름 미측정) | no | yes |
| B13 | `:565` `if err != nil {` | 패키지 전체 시험이 실행(후보 패턴 밖 — 개별 이름 미측정) | no | yes |
| B14 | `:569` `if !automation.Verified {` | 패키지 전체 시험이 실행(후보 패턴 밖 — 개별 이름 미측정) | no | yes |
| B15 | `:574` `if err != nil {` | not-applicable: 오류 반환 — SQLite 읽기 · 쓰기 · 커밋 실패(또는 열기 실패)에서만 열림; 건강한 원장에서 도달 불가이고 결함 주입 시험 없음. a130 무편집 갈래(편집은 이 함수의 다른 줄) | no | n/a |
