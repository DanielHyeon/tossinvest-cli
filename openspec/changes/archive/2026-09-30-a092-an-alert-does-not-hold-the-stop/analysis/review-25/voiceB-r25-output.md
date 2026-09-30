# 보이스 B (code-reviewer, Claude) — 25라운드 원문 (2026-09-29, 트리 git archive 22db26e7)

**판정: APPROVE**. P0는 없고 안전 불변식 위반도 찾지 못했음. 발견은 전부 P2이며, 문서 정직성 1건 · 핀 공백 3건 · 이미 있던 잔여 3건임.

| # | 등급 | 파일:줄 | 무엇 | 제안 |
|---|---|---|---|---|
| 1 | P2 | tasks.md 25.7 vs 21.4 · 21.5 | 25.7 [x] 가 「21.4 · 21.5 닫음」이라 적지만 21.4 · 21.5 체크박스는 [ ] 그대로. 21.4 의 「§6·§8 옛 task 를 하나씩 대체/철회/유지로 판정해 표지」는 이 로트에서 하지 않음 | 21.4 를 「GREEN 코드 부분 완료 / 옛 task 표지 미완」으로 나누고 표지 작업을 25.9/25.10 에 배정 |
| 2 | P2 | tasks.md 23.3 K18 · 25.x | K18 「입구 도달 경로 전수 구조 핀」이 닫힘/미룸 어디에도 없음. 지금 핀은 알려진 주입 지점 4개만 잼. Guardian 은 동기 announcer(`guardian_wiring.go:81`)를 들고 있고 오늘 `IssueReduction` 은 `escalateFor` 를 안 부르지만 누가 붙이면 핀이 없음 | K18 을 25.10 에 명시, exit 주입 부품(Issuer · Submit · Floor · Prices)의 announcer 도달 판정 핀 |
| 3 | P2 | a092_exit_cycle_records_only_test.go · exitwiring.go:207,231 | k3 는 합성으로만 섬(합성 자체는 참). `ConfirmedFloor` 가 Retrier 를 우회하는 편집 · floor 에 다른 announcer 를 넣는 편집을 잡는 핀이 없음. 동결된 22.3 C1(k3)은 행동 RED 를 요구 | `ConfirmedFloor` 행동 시험(exit Retrier + RecordOnly + 막힌 publisher) 또는 모든 `f.official.*` 호출이 `f.retrier.Query` 클로저 안이라는 AST 핀 |
| 4 | P2 | exitwiring.go:346-354 · a092_exit_options_pin_test.go | 기록 전용 보장이 「유일한 생산 호출자가 Alerts · Announcer · Floor 를 넘기지 않는다」에 기댐. 구조 핀은 리터럴 키만 봄 — 리터럴 뒤 대입은 통과(존재 검사이지 역할 검사가 아님) | 생산 경로에서 무조건 기록 전용으로 덮거나 RecordOnly 아닌 값 거절, 또는 조립된 observer 옵션을 읽는 행동 핀 |
| 5 | P2(기존 · C8) | exitloop.go:1537 → :1357 IssueReduction → :1387 Place | 일반 등급 capped 알림이 같은 청산 주문 제출 **앞**에서 동기 best-effort 발행(`record_only.go:49-51`) — 전송이 멈추면 그 손절 주문이 최대 publish 상한(ntfy 10s)만큼 늦음(불변식 4 자리). 회귀 아님 | 단위 ⑤ 에서 먼저 다루거나 적어도 `Place` 뒤로 옮길 것, 25.10 에 순서 사실 기록 |
| 6 | P2 | alertdelivery.go:325-330 | 배달 실행자는 `alert.Body` 만 발행(필드 블록 없음) — 예전 동기 경로(Body + fields)보다 운영자에게 가는 내용이 줄었음 | 의도면 design 에 적고 아니면 payload 렌더 결정 |
| 7 | P2 | retry.go:413-416 · exitloop.go:846-852 | 기록 전용 announcer 의 기록 실패는 `ErrModeAnnouncementFailed` 를 `changed=true` 와 함께 반환 → 호출자 둘이 「did not reach the operating mode, so a restart would lift the block」으로 로그, `checkOutage` 는 `cycle.Escalated` 미설정. 실제로는 커밋됨 | `errors.Is(err, journal.ErrModeAnnouncementFailed)` 따로 처리 |
| 8 | P2 | record_only.go:79-90,131-135 · risk_relaxation_command.go:178 | 25.6 은 부수 효과도 바꿈(실패 → 래치 + 승격). 승격 쪽 시험 없음, logEvent 가 완화 필드(계좌 포함 target)를 구조화 로그에 씀 | 승격 시험 또는 design 기록, 「행 모양 불변 · 실패 부수 효과 변경」 명시 |

저자 주장 1~6: 전부 **참**(1 합성 성립 — 빈틈은 핀 부재, 2 L19 원장 읽기만, 3 코드 판정, 4 참, 5 행 모양 참 · 부수 효과 변경, 6 critical · 모드 통지 범위 참, 일반 등급 둘은 동기 — 하나는 주문 앞).

실행: obs · journal · app/engine · cmd 관련 시험 ok, `-race` obs(A092|A096|A099|A097) · app/engine A092 ok. 변이 재실행 없음.

Recommendation: APPROVE because P0와 안전 불변식 위반이 없고 저자 주장 1~6이 코드로 성립함. 다만 다음 단위 전에 25.7의 21.4 과대 주장과 추적되지 않은 K18을 바로잡고, k3 · 주입 지점의 핀 공백과 주문 앞 일반 등급 동기 발행을 단위 ⑤의 명시 항목으로 올릴 것.
