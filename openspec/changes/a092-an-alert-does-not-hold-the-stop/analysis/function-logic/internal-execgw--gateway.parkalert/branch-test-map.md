# Branch Test Map: `Gateway.parkAlert`

Source: `internal/execgw/replay.go` (534-559). AST 기준 분기 2 / 이탈 0 /
defers 0 / go_statements 0.

`return` 문이 0개인 void 함수다. 관측은 게이트 상태와 outbox 행으로 한다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:535` 참 — 게이트가 배선돼 있다 → `ReasonUnresolvedInDoubt` 래치 · **거짓 — `g.entry == nil`이면 래치를 건너뛰고 outbox만 쓴다** | 참: `internal/execgw/replay_test.go TestReplay422KeyConflictParksAndNeverFails:592`, `TestReplayEchoingADifferentKeyParks:662` · **거짓: 없음** — 게이트 없이 배선한 replay 테스트가 없다 | no | 참만 yes |
| B2 | `:548` `json.Marshal` 실패 → 빈 payload로 계속 | **없음** — marshal 불가한 값을 넣는 경로가 없다(모두 문자열) | no | no |
| — `:551` | outbox에 critical 행이 들어간다 | `replay_test.go TestReplayKeyConflictEnqueuesTheCriticalAlert:628` | no | yes |

B1의 참·거짓 갈래를 한 행에 담았다. `check_analysis.py`가 분기 ID 중복을
거부하므로 한 분기는 한 행이고, 갈래는 행 안에서 나눈다. **거짓 갈래는
미테스트다.**

## a092가 이 함수에 대해 지는 것

편집하지 않으므로 이 함수에 대한 새 RED는 없다. **그러나 이 함수의 결과에
대해서는 진다.**

`TestReplayKeyConflictEnqueuesTheCriticalAlert:628`은 **행이 들어갔다**까지만
관측한다. 그 행이 **나간다**는 관측은 이 저장소 어디에도 없다 — 그것이
`order.unresolved_in_doubt`가 조용히 쌓이고 있는 이유다.

- **§6.0 R17-7** — 이 함수가 넣은 행이 배달 루프를 통해 실제로 발행되는지.
  a092가 만드는 회귀 테스트이고, **현재 HEAD에서 RED다.**

B2는 도달 불가에 가깝다(`map[string]any`의 값이 전부 문자열·상수) —
`not-applicable`: 이 change는 이 함수를 편집하지 않고, 도달 경로를 만들지도 않는다.
