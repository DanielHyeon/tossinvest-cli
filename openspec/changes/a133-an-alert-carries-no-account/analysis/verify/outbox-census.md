# alert_outbox 실측 — a133 (값 미열람, 개수만)

- 2026-10-13 KST, `python3 -I analysis/harness/outbox_census.py` — `~/.config/tossctl/journal.db` 를 `mode=ro` URI 로 열어 읽기만 함.
- 판정 규칙은 a131 `internal/redact` 의 것을 Python 으로 옮긴 사본(측정 전용 — 생산 규칙 아님): 등록 원문(attestation 의 `account_ref`, 메모리에서만) ·
  대시로 끊긴 숫자 덩어리의 숫자 = 등록 숫자열 · 마스크 꼴 · 유도 id 꼴. 계좌 값은 출력하지 않음.

```text
registered_ref_present True digits_len 11
total_rows 16
by_state {'DELIVERED': 16}
event_key raw {'DELIVERED': 2} mask {} derived {'DELIVERED': 12}
title raw {} mask {} derived {}
body raw {} mask {} derived {}
payload raw {'DELIVERED': 4} mask {} derived {'DELIVERED': 12}
last_error raw {} mask {} derived {}
payload_account_key {'DELIVERED': 4}
title_or_body_any(would_be_sent_if_pending) {}
pending_types_count 0 0
```

## 해석

- 16 행 전부 DELIVERED → PENDING 0 이라 **재전송 대기 중인 행이 없다**. 엔진 전달기(`alertDeliverer.cycle`)와 `Flush` 는 PENDING 만 나열한다.
- 정착 행은 재알림 창이 지나 같은 사건 키가 다시 관측될 때 **새 내용으로 덮여** 재무장된다(`recordAlertTx` 의 UPDATE 가 title · body · payload 를 새 값으로 씀) —
  옛 행의 저장 내용이 그대로 다시 나가는 길은 없다.
- title · body 에 계좌 0. 계좌는 `payload`(원문 4 · pos- 12)와 `event_key`(원문 2 · pos- 12)에만 있다. payload 와 event_key 는 **전송되지 않는다**
  (재전송 경로는 title · body 만 씀 — `alertdelivery.go` `deliverOne` · `notifier.go` `Flush`). 다만 실시간 경로 `deliver` 는 같은 사건의 `Fields` 를 본문에 펼쳐
  보냈으므로, payload 에 계좌가 있는 4 행과 pos- 가 있는 12 행의 그 값은 처음 전달 때 ntfy 로 나갔을 수 있다.
