# Branch Test Map: `_walk_floor`

짝은 변이(`77_mut.py`)가 **실제로 빨갛게 만든** 시험이다. 부르는 쪽이 둘이라 두 경로의 시험이 같이 잡는지 본다.

| 갈래 (소스) | 변이 | 계산 경로에서 잡는 시험 | 조언 · 기록 명령 경로에서 잡는 시험 |
|---|---|---|---|
| `if not _pinning_bundles(root, analysis)` | R17 | `test_the_computation_names_what_stops_it_before_the_walk` | `test_it_refuses_when_no_evidence_pins_a_landing` · `test_a_refusal_that_outlives_a_commit_is_named_before_one_that_does_not` |
| `if not floor` | R18 | `test_the_computation_names_what_stops_it_before_the_walk` | `test_uncommitted_evidence_is_named_as_the_reason_nothing_was_recorded` · `test_the_advice_never_names_a_command_that_would_refuse` |
| `return (floor, '')` | **변이 안 걸었다** | 이 갈래를 **지나가는** 시험: `test_the_recorded_value_is_one_the_gate_then_accepts` (죽이는지는 재지 않았다) | 지나가는 시험: `test_where_the_command_would_record_the_advice_still_names_it` |

`compute_landing` 의 `if why:` 를 지우는 R16 은 `test_the_computation_names_what_stops_it_before_the_walk` **하나만** 잡는다.
기록 명령은 판정 함수가 같은 사유로 먼저 멈추므로 그 경로로는 안 닿는다 — 그 시험이 이 함수를 직접 부르는 이유다.
