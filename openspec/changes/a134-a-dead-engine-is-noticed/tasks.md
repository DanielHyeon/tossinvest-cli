# a134 Tasks

Story: `STORY-TOS-a134` · 알림 · 생존 관측 경로(주문 · 손절 · 사이징 경로 아님) — 비례 원칙: 시험 + lint + 경량 독립 리뷰.

## 0. 계약 · 증거

- [x] 0.1 Story 등록(`docs/pm/portfolio/stories/STORY-TOS-a134.yaml`, `_registry.yaml`, `FEAT-TOS-009` 역링크) + tracker 재생성 + proposal — `421bf53c`
- [x] 0.2 design(D0 ntfy 전제 반증 · D1 기존 마커 채택 · D2 판정표 · D3 상태 · D4 전송 · D7 dead-man 설계만) · spec 델타 · tasks,
      `openspec validate a134-a-dead-engine-is-noticed --strict --no-interactive` rc=0
- [x] 0.3 `capture_change_base.py` 로 base 고정 — `421bf53c`
- [x] 0.4 CodeGraph hard evidence(`analysis/code-context/`) — codegraph 1.6.0, `affected --filter '*_test.go'` 1070 · 불일치 R5(문서 전제) 하나
- [x] 0.5 Function Logic Map — 편집하는 기존 함수 `newEngineCmd`(분기 0) 번들. 그 밖은 새 파일 · 새 leaf: `Function Logic Map: not-applicable` 사유는 review.md
      — `analysis/function-logic/cmd-tossctl--newenginecmd/`(편집 전 sha 446e27a3… · 편집 뒤 f311f5ad…, 분기 0 → 0) · Pre-Edit 선언 review.md

## 1. 구현 (RED → GREEN → REFACTOR)

- [x] 1.1 RED — 판정표 · 중복 억제 · 복구 · 전송 실패 · 알림 OFF · 엔진 무요구 · 계좌 칸 없음 · 명령 등록 시험이 base 에서 실패(`analysis/red/`)
      — 새 패키지라 RED = 빌드 실패(기호 미정의): `red-enginewatch.txt` · `red-cmd.txt`
- [x] 1.2 GREEN — `internal/enginewatch`(판정 · 상태 · 한 번 전송) · `engine.NotificationPublisher` 입구 · `tossctl engine heartbeat-check` · `newEngineCmd` 등록
      — enginewatch 14/14 · cmd a134 4 + 규약 3 PASS
- [x] 1.3 REFACTOR — `internal/obs/heartbeat.go` 주석 정정(함수 무편집) · systemd 단위 파일 · 편집 뒤 AST 재추출 · `check_analysis.py --change` 통과
      — 주석 · 단위 파일 · 번들 완료. check_analysis 는 착지 기록(커밋 뒤) 후 영수증 — 그 전 판정은 base 뒤 a128 커밋(7c16a776)의 형제 함수까지 창에 잡음

## 2. 검증

- [x] 2.1 `gofmt -l`(빈 출력) · `go vet ./...` · `rtk proxy go test -count=1` (`internal/enginewatch` · `internal/obs` · `internal/app/engine` · `cmd/tossctl`) · `openspec validate --strict`
      — `analysis/verify/receipts.md`: 전부 rc 0, 단 cmd/tossctl 전체의 실패 1 은 병행 a135 의 미커밋 시험(a134 무관). 경량 변이 4/4 CAUGHT
- [x] 2.2 경량 독립 리뷰(분리된 컨텍스트) → `review.md`
      — APPROVE-WITH-FIXES(P0 0 · P1 1 · P2 3 · P3 5) — P1-1 · P2-1 · P3-3 · P3-4 수리(시험 6 추가, 변이 M5~M7 CAUGHT), P2-2 · P2-3 · P3-1 · P3-2 · P3-5 는 design 잔재로 명시

## 3. 사람 · Manager

- [ ] 3.1 Manager `make sdd-sync` · `make sdd-check` · `make gate CHANGE=a134-a-dead-engine-is-noticed` + 독립 검증

### 게이트 뒤 절차 — 사람 게이트 (체크박스 아님 — a108 선례)

감시자 설치는 **사람이 한다**. 에이전트는 단위를 설치 · 활성화하지 않는다.

1. 이 change 를 담은 빌드로 호스트 `~/.local/bin/tossctl` 을 교체한다(엔진 이미지 교체 불요 — 엔진 코드 무편집).
2. 보호된 토픽(토큰)이거나 토픽을 환경 변수로만 준 배포면 drop-in 을 만든다:
   `systemctl --user edit tossos-heartbeat.service` → `[Service]` 아래 `EnvironmentFile=%h/.config/tossctl/ntfy.env`
   (파일 안 `TOSSCTL_NTFY_TOKEN=…` · 필요하면 `TOSSCTL_NTFY_TOPIC=…`, 권한 0600, 저장소에 넣지 않음).
3. 단위 파일을 놓는다: `install -m 0644 deploy/systemd/tossos-heartbeat.service deploy/systemd/tossos-heartbeat.timer ~/.config/systemd/user/`
   → `systemctl --user daemon-reload`.
4. **전달 실측(필수, 리뷰 P1-1)** — 단위와 같은 환경으로 시험 알림 한 건:
   `systemd-run --user --wait --pipe -p EnvironmentFile=-$HOME/.config/tossctl/ntfy.env $HOME/.local/bin/tossctl --config-dir ~/.config/tossctl engine heartbeat-check --send-test`
   (drop-in 이 없으면 그냥 `~/.local/bin/tossctl --config-dir ~/.config/tossctl engine heartbeat-check --send-test`).
   - 기대: 휴대폰에 「TossOS 엔진 생존 감시 시험」 도착 + 종료 코드 0. 실패(비정상 종료)면 토큰 · 토픽부터 고친다 — 고치기 전에는 타이머를 켜지 않는다.
   - `알림을 보낼 수 없음 …` 오류면 알림 설정(콘솔 [알림 켜기])부터 한다. 감시자는 알릴 수 없으면 매번 실패한다(D4).
   - 이어서 판정 한 번: `… engine heartbeat-check` → `engine heartbeat  up · …` 한 줄, 종료 코드 0, 알림 없음.
5. 활성화:
   ```bash
   systemctl --user enable --now tossos-heartbeat.timer
   systemctl --user list-timers tossos-heartbeat.timer
   loginctl show-user "$USER" -p Linger   # Linger=yes 여야 로그아웃 뒤에도 돈다(tossos-attest 와 같음)
   ```
6. 확인: `journalctl --user -u tossos-heartbeat.service -n 5` 에 1분마다 `engine heartbeat  up` 줄.
7. 계획된 정지(배포 · KST 05:00~09:00 교체) 동안 autostart ON 이면 「끊김」 1건 + 「복구」 1건이 온다 — 정상. 원치 않으면 그 동안
   `systemctl --user stop tossos-heartbeat.timer` 후 다시 `start`.
8. (선택, 기본 OFF — design D7) 기계 전체 정지까지 잡으려면 외부 박동 서비스 URL 을 drop-in 의 `ExecStartPost=/usr/bin/curl -fsS -m 10 --retry 2 <url>` 로 붙인다.
