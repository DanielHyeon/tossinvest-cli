# a133 — 시스템 밖으로 나가는 알림에는 계좌 정보가 실리지 않는다 (an-alert-carries-no-account)

## Why

사용자 결정(2026-10-13): **계좌 정보는 로그에 남기지 않는다 + 알림 유출을 고친다.** 안전 불변식 8의 집행이다.

a131 이 범위 밖에서 찾았고 Manager 가 확인했다(a131 design D0 마지막 행, 잔재 ③). 알림 본문은 `obs.Event.Fields` 를 펼친다
(`notificationFor` → `renderFields`, `internal/obs/notifier.go:1003-1040`). 그래서 계좌 원문(`account`)과 계좌 유도 식별자
(`position_id` = `pos-` + 24 hex — 소금 없는 해시라 대입으로 되돌릴 수 있음, a131 D0)가 **외부 서비스(ntfy)** 로 나간다.
a131 은 로그·기록 파일 싱크만 막았고 알림 payload 는 원문 그대로 뒀다. 외부 전송이므로 로컬 로그보다 심각하다.

운영 원장 실측(2026-10-13, 읽기 전용 · 값은 열람·출력하지 않고 개수만 셈 — `analysis/verify/outbox-census.md`): `alert_outbox` 16 행 전부
DELIVERED. `payload`(= 그 알림의 `Fields`)에 계좌 원문 4 행 · `pos-` 유도 id 12 행. 실시간 경로(`deliver`)는 이 `Fields` 를 본문에
펼쳐 보낸다 — 이 행들의 계좌·포지션 id 는 이미 ntfy 로 나갔을 수 있다(서버 쪽은 이 저장소에서 잴 수 없음).

## What Changes

1. **경계 한 곳** — 유일한 생산 발행자 `obs.(*Ntfy).Publish` 가 요청을 만들기 **전에** 알림을 a131 의 규칙(`internal/redact`)에
   통과시킨다: Title · Body · Tags(사건 유형 태그 포함) 각각 `redact.String`. 규칙 사본 없음. 요청 URL 은 설정의 서버 + 토픽뿐임을 시험이 고정한다.
2. **필드 펼침에서 계좌 키 버림** — `renderFields` 가 `redact.IsAccountKey` 인 필드를 본문에 쓰지 않는다(이름은 구조가 있을 때만 알 수 있음).
   나머지 필드 값은 1 의 경계에서 치환된다.
3. **구조 핀** — 비시험 코드에서 `obs.Publisher` 를 구현하는 타입은 `*obs.Ntfy` 하나뿐이고 그 `Publish` 가 치환 함수를 부른다(go/parser).
   새 전송 수단이 경계를 우회하면 시험이 실패한다.
4. 회귀 방지 시험 — 계좌 센티널과 계좌 유도 id 를 Fields · Title · Body 에 심고 **httptest ntfy 가 받은 바이트**(본문 · 헤더 · URL)에
   없음을 단언한다: 실시간 critical · best-effort · normal relay · outbox 재전송(엔진 전달기 · Flush) · 알림 시험 명령.

## What Does NOT Change

- 알림의 운영 의미: 종목 · 사유 코드 · attempt id · 수량 · 시장 · 사건 유형은 그대로 남는다. 포지션 id 는 a131 의 로그와 같이 `pos-[redacted]`.
- critical 전달 · 재시도 · 전달 실패 판정(a124 미전달 3회 → 진입 래치 · 운영 모드 승격) · 임차 · 승인 해제 — `Publish` 의 반환은 같은 입력에 같다.
- 알림 경로가 청산 · 손절을 막거나 늦추지 않는 구조(a092 기록 전용 입구 · 잠금 밖 전송). 치환은 순수 문자열 연산이며 전송 goroutine 안에서만 돈다.
- outbox 저장 내용(`event_key` · `title` · `body` · `payload`)과 원장 스키마 — design D4 의 결정.
- 주문 · 청산 · 대사 · 체결 감지 · 인터록. 브로커 호출 수(0 추가).

## Impact

- Affected specs: `engine-safety`(요구 1개 추가).
- Affected code: `internal/obs/ntfy.go`(`Ntfy.Publish`), `internal/obs/notifier.go`(`renderFields`), 새 `internal/obs/a133_transport_redact.go`(치환 leaf).
  시험: 새 `internal/obs/a133_*_test.go` · `internal/app/engine/a133_*_test.go` · `cmd/tossctl/a133_*_test.go`.
- High-risk 인접(alert · operating mode 승격의 입력인 전달 결과): 편집하는 기존 함수 2개 모두 Function Logic Map + Branch Test Map 을 편집 전에 만든다.
- 배포: a129 · a131 · a130 과 한 번에(사용자 결정 「다 완료 후 배포」). 형식 · 스키마 변경이 없어 순서 의존을 더하지 않는다(design 「배포」).
