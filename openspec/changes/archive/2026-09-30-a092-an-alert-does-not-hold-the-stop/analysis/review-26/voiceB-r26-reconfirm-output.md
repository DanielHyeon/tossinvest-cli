# 26라운드 보이스 B 재확인 (55b435ea, R26B-FIX.diff) — APPROVE-WITH-FIX

판정: 표시면 누설은 닫힘, 엔진 로그 한 줄이 가림 없이 남음(부분). 수리 필요, BLOCK 사유 아님.

사본 `/tmp/claude-1000/a092-r26-B2-*`(`git -C` rev-parse 실패 단언, 끝난 뒤 삭제)에서 `-run A092` 로 obs · execgw · journal · app/engine · cmd/tossctl 전부 ok, 프로브 1건.

## 1. #1 · #2 — 부분(표시면 닫힘, 엔진 로그 한 줄 남음)

- **닫힘 · 기록 전용 입구 로그 필드**: 세 메서드 모두 `n.logEvent(withoutFields(e), …)`. 정상 경로 로그에 계좌 원문 0건.
- **닫힘 · 게이트 설명**: 고정 문구. 원장을 닫아 기록 실패시켜도 계좌 0건.
- **닫힘 · modeops 결과 칸**: 전부 고정 상수. 재읽기 · 목록 실패 원문은 `logFailure` 로 가려짐.
- **닫힘 · 500 경로**: `ops.logFailure(...)` 뒤 고정 문구. 400 경로는 `err.Error()` 그대로이나, Release 에서 도달하는 journal `ErrInvalidRequest` · `ErrModeApprovalRequired` 문구를 전수로 읽었고 계좌를 담은 것이 없음.
- **닫힘 · CLI**: 서버가 준 고정 문구만 출력.
- **안 닫힘 · 엔진 로그 한 줄**: `recordCritical` 의 `n.Log.Error(EventAlertUndelivered, err, …)` — a092 새 함수이고 원문 오류를 가리지 않음.
  - 프로브: `{"level":"ERROR","msg":"engine.alert_undelivered",…,"error":"journal: recording alert operating_mode:99887766554:NORMAL:t2: sql: database is closed"}`
  - Manager 판정 (a) 와 어긋남 → P2.
- 표시면으로 가는 a092 새 경로: 남은 것 없음.

## 2~4. 나머지 항목 — 닫힘

- #9 — 닫힘(`runtime.go` · `riskguardian.go` 모두 `errors.Is` 갈래, guardian 시험 있음).
- #10 — 닫힘(`grep "투영기 미배선" internal` 0건).
- #6 — 닫힘.
  - 잔여(P2, 생산 무관): `Notifier.Journal == nil` 조립에서 행이 없는데도 「대기 목록에 없음」 문구.

## 5. 이번 diff 의 새 스펙 위반 — 없음

- `detachedAnnouncer` · `WithoutCancel` 은 커밋 뒤에만 적용되고, 델타 순서(audit → commit → 투영 → 통지)는 유지.
- 선점 로그 · 모르는 결과 Error 갈래는 기록만 더함.

## 새 발견

| # | 등급 | 파일:줄 | 무엇 | 근거 | 제안 |
|---|---|---|---|---|---|
| N1 | P2 | record_only.go `recordCritical` Log.Error | 기록 실패 원문(키에 계좌) 무가림 — a092 새 함수 | 프로브 원문 | `[account]` 가림 |
| N2 | P2(재확인 범위 밖) | normal_relay.go `Run` | 1차 #3(패닉 정지 뒤 무기록 유실)이 이번 diff 에 없음 | diff | 처분 기록 |

## base 관행 잔여 — 목록만, 발견 아님 (판정 (b))

- `notifier.escalate` 의 `FieldAccount` · journal 오류 원문.
- `notifier.claimAndDeliver` 게이트 설명 `%v err`(base 721d0338:275).
- `runtime.log` 의 계좌 필드와 새 escalate 줄의 `err.Error()`.
- `riskguardian.escalateFor` 반환 오류의 원문 err.
- 동기 통지자 `logEvent` 필드.
- `reconcile notifierAlerter` 의 `FieldAccount`.

Recommendation: APPROVE-WITH-FIX — 표시면 누설은 닫혔고, 남은 것은 recordCritical 로그 한 줄의 가림 누락(P2).
