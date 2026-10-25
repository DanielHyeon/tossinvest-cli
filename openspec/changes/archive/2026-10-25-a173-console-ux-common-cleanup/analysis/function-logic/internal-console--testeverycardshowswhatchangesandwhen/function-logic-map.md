# Function Logic Map: `TestEveryCardShowsWhatChangesAndWhen`

- Source: `internal/console/settings_card_test.go` (83-112) — 편집 뒤(a173 rebase 후 main).
- Qualified: `TestEveryCardShowsWhatChangesAndWhen`
- AST evidence: `ast.json` (`source_sha256` cf628bf01a4f118e…) — AST branches 6 · returns 0
- Risk scan: `risk-pattern-report.md`
- 작성 시점: 완료 게이트 직전(2026-10-25, 종결 에이전트). tasks 3.1 의 비례 원칙 면제는 required 7 이라 gate ⑤ 가 받지 않아 사후에 만든 산출물임(FLM 이 구현 판단에 쓰이지 않음).

**역할.** 설정 카드마다 `적용 후` 미리보기와 `반영 시점` 한 줄이 있는지 고정하는 시험.

a173 편집: B6 의 단언 대상 class 를 `notice` → `muted effect-timing` 으로 교체(알림 3단계 스펙: 반영 시점은 상자가 아니라 한 줄 본문). 약화 아님 — 마커 앞 60 자 안 해당 class 필수, 카드 둘 다(tasks 「고정 리터럴 시험 수정 기록」 1).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 설정 탭 경로 · 카드 | settingsTabPaths · cardsOn | `fullSettingsHarness` | 단언 실패 → t.Errorf |

## Branches and early returns

| Branch | Source line | AST kind | Condition / 효과 | a173 |
|---|---|---|---|---|
| B1 | `86:2` | range | 설정 탭 순회 | 무변경 |
| B2 | `87:3` | range | 카드 순회 | 무변경 |
| B3 | `88:4` | if | 미리보기 없음 → Errorf | 무변경(실패 갈래) |
| B4 | `100:2` | range | adoption · gate 순회 | 무변경 |
| B5 | `103:3` | if | 반영 시점 마커 없음 → Errorf · continue | 무변경(실패 갈래) |
| B6 | `107:3` | if | 마커 앞 class 없음 → Errorf | a173: class 리터럴 교체(실패 갈래) |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `fullSettingsHarness` · `h.authenticate` · `h.get` · `body` · `cardsOn` | 렌더 얻기 | 하네스 | AST calls 84 · 85 · 87 · 99 |
| `strings.Contains` · `strings.Index` | 단언 | 순수 | AST calls 88 · 102 · 107 |
| `t.Errorf` | 실패 보고 | — | AST calls 89 · 104 · 108 |

## State mutations and fallbacks

- 시험. 상태 변경 0.

## Safety conclusion

- Safe edit boundary: 시험 리터럴 교체. 생산 경로 0.
- High-risk impact: no — 콘솔 읽기 전용 렌더(주문 · 손절 · 사이징 · 원장 · 인증 경로 미접촉).
