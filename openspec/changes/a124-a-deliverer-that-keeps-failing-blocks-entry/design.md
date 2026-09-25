# a124 design — 판정은 원장의 시도 수에 선다

## Context

a098 이 critical outbox 를 비우는 보조 실행자를 세웠고(감독 밖, 결정 9-2), a099 가 배제를 원장 임차로 옮겼다.
둘 다 「지속 실패 → 진입 차단·모드 승격」을 실행자에 두지 않았다 — 오늘 그 일은 동기 경로 `Notifier.deliver` 가
한다(proposal 표). a092 가 그 동기 시도를 엔진 루프에서 빼면 그 일의 주인이 사라진다. 이 change 가 그 주인을
실행자로 정한다.

## D1 — 주인은 실행자, 판정 입력은 행의 `attempts`

| 후보 | 재시작 | 두 실행자(재시작 직후 겹침) | 채택 |
|---|---|---|---|
| 실행자 메모리의 연속 실패 사이클 수 | 지워진다 | 각자 센다 | ✗ |
| **원장 행의 `attempts`** (`outbox.go:58`, `MarkAlertAttemptFailed` 가 +1 :471) | 남는다 | 같은 값을 본다 | **✓** |

정본 시나리오가 「재시도 한도까지 실패하면」이라 적으므로 단위는 **행**이고 값은 **시도 수**다. 판정 자리는
`deliverOne` 의 실패 경로(B8 · B9) 끝 — `MarkAlertAttemptFailed` 가 돌려주는 갱신된 행(또는 그 반환값)의 `attempts`
가 한도 이상이면 잠그고 승격한다. 게이트는 멱등(`Block` 재호출은 같은 사유의 갱신), 승격도 멱등
(`EscalateOperatingMode` 는 이미 그 모드면 `changed=false`). 그래서 한도 이상의 행이 매 사이클 다시 시도돼도 폭풍이
아니라 같은 사실의 재확인이다.

## D2 — 한도 값 (열린 결정)

| 안 | 값 | 두절 → 승격까지 | 대가 |
|---|---|---|---|
| ㄱ 정본 「재시도 한도」= 동기 경로와 같은 횟수 | `attempts ≥ 3` | 약 2 사이클 ≈ 4~6 s | 짧은 transport 흔들림이 durable 승격 → 사람 완화 |
| ㄴ 동기 경로와 **등가 시간** | 34 s 에 해당하는 사이클 수 (실측) | ≈ 34 s | 한도가 시간 상수에 의존 — `alertDeliveryInterval` 이 바뀌면 같이 |
| ㄷ 별도 상수 | 실행자 전용 | 설계값 | 두 한도가 갈린다 |

Manager 권고는 **ㄱ** — 정본 문장에 가장 가깝고 상수가 하나다. 대가는 review 에 적고 freeze 리뷰가 답한다.
어느 안이든 값은 상수 하나로 두고 시험이 그 상수를 인용한다(기억: 계약 숫자는 영수증에서).

## D3 — publisher 부재는 실패 시도로 센다 (열린 결정)

동기 경로는 `Publisher == nil` 이면 시도 없이 루프를 끝내고 곧장 잠근다(`deliver` B3 → B27). 실행자의 B8 은 오늘
`attempts` 를 올리지 않고 반납만 한다. 이 change 는 B8 도 `MarkAlertAttemptFailed` 를 거치게 해 같은 한도를 타게
한다 — "설정되지 않은 publisher" 는 "응답 없는 publisher" 와 운영자에게 같은 사실이다. 반대 안(세지 않는다)은
무설정 엔진이 영구히 진입을 열어 두는 구멍이다.

## D4 — 굶주림 없는 선택

`PendingAlerts` 를 `ORDER BY (attempts >= ?) , id` 로 바꾸거나(한도 아래 먼저, 그 안에서 오래된 것 먼저), 두 번
조회(한도 아래 `LIMIT n`, 남은 자리만 한도 이상)한다. 전자가 한 문장이라 YAGNI 다. 한도는 D2 의 상수를 인자로
받는다 — SQL 에 값을 박지 않는다. 버리는 경로는 없다.

## D5 — 배선

`alertDeliverer` 에 `Gate *execgw.EntryGate` · `AccountRef string` 를 더하고 `auxiliary.go` 의 생성 자리에서 넘긴다.
`Journal` 은 이미 있다. 승격은 `Journal.EscalateOperatingMode(ctx, AccountRef, ModeTriggerCriticalAlertUndelivered, nil)`
— `Notifier.escalate` 와 같은 호출(:383)이라 announcer 는 nil 로 둔다(실행자는 `Notify` 를 재진입할 이유가 없다).

## D6 — 최악 래치 시간 (R3)

큐 대기 + 시도 = (선행 배치 서비스 시간 ≤ 10 × publish timeout) + (한도 × interval + 한도 × publish timeout).
값은 tasks 4.4 에서 실측해 review 에 적는다. 큐를 뺀 값을 상한이라 부르지 않는다.

## 안전 불변식 대조

- 손절 즉시성: exit 관측 루프를 만지지 않는다. `a098_the_backlog_does_not_delay_protection_test.go` 가 그대로 초록이어야 한다.
- 토글 OFF = upstream: 새 토글 없음. 알림 게이트 OFF 면 `Notifier` 자체가 없다(a092 20라운드 A 확인).
- 보수 방향: 이 change 가 여는 문은 없다 — 오늘 안 잠기던 자리가 잠긴다.

## Risks

- D2 ㄱ 은 durable 승격을 더 자주 만든다. 완화는 사람 승인 — 운영 마찰의 크기를 review 에 적는다.
- a092 가 먼저 착지하면 정본 SHALL 이 빈다. 순서를 a092 tasks 의 착수 조건으로 못 박는다.
