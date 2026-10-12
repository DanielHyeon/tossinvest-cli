# Branch Test Map: `NewContext`

편집 전 좌표(base `233d5d9c`). RED/GREEN 은 구현 뒤 갱신(아래 「a130 시나리오」).

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:439` resolveEnginePaths 오류 | 무편집 | no | no |
| B2 | `:446` 감사 로그 열기 오류 | 무편집 | no | no |
| B3 | `:450` clk nil | 무편집 | no | no |
| B4 | `:464` Publisher 주입 | 무편집 | no | no |
| B5 | `:472` 감사 기록 오류 | 무편집 | no | no |
| B6 | `:487` 계좌 해석 오류 | 무편집 | no | no |
| B7 | `:494` 저널 열기 오류 | 무편집 | no | no |
| B8 | `:503` bindApplyHooks 오류 | a130 편집 자리 — 호출 하나를 `bindApplyHooksLogged(jrn, opts.Logger)` 로 | no | no |
| B9 | `:522` 게이트웨이 오류 | 무편집 | no | no |
| B10 | `:537` Guardian 생성 조건 | 무편집 | no | no |
| B11 | `:539` factory nil | 무편집 | no | no |
| B12 | `:545` Guardian 오류 | 무편집 | no | no |
| B13 | `:565` 인터록 오류 | 무편집 | no | no |
| B14 | `:569` 미검증 | 무편집 | no | no |
| B15 | `:574` 오류 | 무편집 | no | no |

## a130 시나리오(편집 계획에 묶인 시험)

- B1~B7 · B9~B15 — 무편집 갈래 — 기존 엔진 기동 시험(a129 와 같은 관례로 무편집 갈래는 인용만)
- B8 — 바인딩 호출이 로거를 넘김 / 바인딩 실패 거절 유지 — a130 구조 핀 · 기존 interlock 시험
