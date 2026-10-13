# a134 — 죽은 엔진은 엔진 밖에서 알아챈다 (a-dead-engine-is-noticed)

## Why

보호 주문이 UNWIRED 인 지금 손절 · 청산을 실행하는 것은 **엔진 프로세스 하나뿐**이다(엔진 정지 = 손절 없음). 엔진은 콘솔 컨테이너 안의
자식 프로세스(`console` · `engine run` · `soak run` 이 각각 따로)라서, 엔진이 OOM · 크래시로 죽어도 콘솔이 건강하면 컨테이너는 살아 있고
`restart: unless-stopped` 는 반응하지 않는다(컨테이너 전체가 끝날 때만 반응). 지금 그때 **아무도 통지받지 못한다.**

Manager 실측(2026-10-14)과 이 change 의 HEAD 대조:

- `obs.Heartbeat`(`internal/obs/heartbeat.go`)는 생산 생성자가 **0** 이다(`grep -rn "obs.Heartbeat{" cmd internal` 비시험 0).
- 그 파일 주석과 `engine-safety` 「등급화된 알림」의 문장 「heartbeat(예상 주기 초과 시 **ntfy 측에서** 경보)」의 전제 — ntfy 가 박동이
  끊기면 스스로 경보한다 — 는 **근거가 없다.** 저장소 안의 ntfy 문서(`docs/operations.md` · `AGENTS.md` · `README.md`)는 발행 · 구독
  필터만 말하고, 수신 측 침묵 경보 기능은 어디에도 없다. ntfy 는 받은 메시지를 전달하는 서버이고 「안 온 메시지」를 판정하지 않는다.
  그러므로 침묵을 판정하는 쪽은 **엔진 밖에 살아 있는 무언가**여야 한다.
- 엔진 쪽 생존 신호는 이미 있다: 엔진 활성 마커 `engine-run.json`(`internal/enginelock`, 부팅 6단계)이 저널 디렉터리에 1분마다
  **원자적으로**(임시 파일 + rename) 다시 쓰이고, 정상 정지 때 지워지며, 내용은 pid · 시작 시각 · 갱신 시각 · 빌드 도장 · 준비 시각 ·
  프로세스 인스턴스뿐이다(계좌 칸 없음). 빠진 것은 **그 침묵을 읽는 쪽**이다.

## What Changes

1. **엔진 밖 감시 명령** — `tossctl engine heartbeat-check`(읽기 전용 판정 + 자기 상태 파일 쓰기 + 알림 1건). 엔진 저널 디렉터리의
   마커 갱신 시각을 읽어 판정한다:
   - 마커가 기준(기본 3분)보다 오래됨 → **끊김**(죽은 프로세스는 마커를 못 지움 — OOM · SIGKILL · 호스트 정지).
   - 마커 없음 + `engine.autostart` ON(엔진이 돌아야 한다는 사람의 승인) → **끊김**.
   - 마커 없음 + autostart OFF → 감시 대상 아님(정상 정지 · 수동 운용).
   - 끊김으로 **바뀐 순간 한 번** critical ntfy 알림, 끊김에서 **벗어난 순간 한 번** 복구 알림. 상태 파일(`engine-heartbeat-watch.json`)이
     중복을 막는다. 전송에 실패하면 상태를 옮기지 않아 다음 실행이 다시 보낸다.
   - 전송 수단은 엔진과 **같은 규칙**(`resolveNotificationPublisher`, 새 내보내기 입구 하나)으로 얻은 `obs.Ntfy` — a133 전송 경계를 지난다.
     알림이 꺼져 있거나 토픽이 없으면 **실패 종료 + 이유 문구**(조용한 성공 없음).
   - 엔진 · 저널 · 브로커 · 세션을 요구하지 않고 열지도 않는다.
2. **호스트 단위 파일** — `deploy/systemd/tossos-heartbeat.service` · `.timer`(1분마다, `tossos-attest` 선례). 설치 · 활성화는 사람이 한다(tasks 3.x).
3. **오해 주석 정정** — `internal/obs/heartbeat.go` 의 파일 · 타입 주석에서 「ntfy 가 스스로 경보」를 지우고, 생산 배선 0 과 실제 경로(마커 + 감시 명령)를 적는다.
   함수 본문 무편집.

## What Does NOT Change

- 엔진 런타임 · 감독 루프 · 보조 실행자 · 마커의 갱신 주기(1분) · stale 창(5분) · flock 배타 · 기동 순서. **엔진 프로세스 코드 무편집.**
- 주문 · 손절 · 청산 · 대사 · 체결 감지 · Guardian · 인터록 · outbox · 진입 래치 · 운영 모드. 브로커 호출 0 추가.
- 알림 등급 체계와 a133 전송 경계. 감시자의 알림은 outbox 를 거치지 않는다(엔진 밖 · 저널 비소유).
- 설정 스키마(새 키 0). 원장 스키마(변경 0).

## Impact

- Affected specs: `engine-safety`(요구 1개 추가 — 「엔진의 침묵은 엔진 밖에서 판정된다」).
- Affected code: 새 패키지 `internal/enginewatch`(판정 · 상태 · 전송 한 번), 새 `cmd/tossctl/engine_heartbeat_check.go`(명령),
  새 `internal/app/engine/a134_notification_publisher.go`(기존 해석 규칙의 내보내기 입구 — 규칙 사본 없음),
  `cmd/tossctl/engine.go` 의 `newEngineCmd` 한 줄(하위 명령 등록), `internal/obs/heartbeat.go` 주석.
  새 `deploy/systemd/tossos-heartbeat.{service,timer}`.
- 위험 등급: 알림 · 생존 관측 경로. 주문 · 손절 · 사이징 경로가 아니다. 비례 원칙(docs/WORKFLOW.md)상 무거운 규율(변이 원장 · 다중 적대 리뷰)은
  의무가 아니다. 게이트가 요구하는 수정된 기존 함수의 FLM 은 `newEngineCmd` 하나(분기 0)뿐이다.
- 배포: 호스트 `tossctl` 교체 + 사람이 단위 파일 설치. 이미지 교체 불요(엔진 코드 무편집).
