# Verification: console-adoption-controls

## 2026-07-31 — 외부 종목 자동관리 메뉴 개정

### 범위

- 기존 `engine.adoption` 설정·대사·정책 적용 경로는 수정하지 않았다.
- navigation label/href, adoption section anchor, 제목·설명과 렌더 회귀 테스트만
  변경했다.
- config 값, journal, automation gate, Guardian, 엔진 기동, 주문·LIVE 경로를
  변경하거나 호출하지 않았다.
- 기존 함수 내부 로직 수정이 없으므로 이번 개정의 Function Logic Map과 Branch Test
  Map은 not-applicable이다.

### RED → GREEN

- RED: `TestExternalPositionAutomaticManagementHasADiscoverableMenu` 추가 직후
  메뉴 href/label, section anchor, 제목, 수동 매수·기존 정책·저장/실행 경계의 7개
  단언 실패를 관측했다.
- GREEN: 기존 `/settings` route와 저장 seam을 유지한 채 template만 최소 수정했다.
- `go test -race ./internal/console` — PASS (`34.060s`, 최종 배포 소스).

### 회귀·정적 검증

- `make test` — PASS.
- `make vet` — PASS.
- `make validate` — PASS (`44/44`).
- `openspec validate console-adoption-controls --strict --no-interactive` — PASS.
- PM generator/check — PASS; `STORY-TOS-025` ↔ `console-adoption-controls` 1:1 유지.
- code-reviewer quality checker (`internal/console`, Go) — PASS, issue 0.
- `git diff --check` — PASS.

### 배포 스모크

- `docker compose build` — PASS, image
  `sha256:b203b98efa4c1f21f47d7eaabdb3bd0bc985339b81228c3df4a6c4b5369842d5`.
- `docker compose up -d --no-build` — PASS.
- container `tossos-tossos-1` — `healthy`, restart policy `unless-stopped`,
  `127.0.0.1:37085->37085/tcp`.
- `GET https://127.0.0.1:37085/settings` — 200; menu label/href, `#adoption`,
  제목·설명, 기존 enabled/default stop controls를 렌더 HTML에서 확인했다.
- `GET /` — 200, `GET /login` — 404: 기존 trusted-network 무인증 동작 유지.
- 배포 직후 engine marker 0개. `engine.autostart`는 config에 없어서 기본 OFF다.
- 기존 사용자 config의 `engine.adoption.enabled=true`와
  `engine.automation_gate.enabled=true`는 읽기 확인만 했고 수정하지 않았다.

### SDD 계층 상태

- CodeGraph hard index는 메뉴 개정 파일을 동기화했다.
- CodeGraphContext advisory update는 응답 없이 정지해 중단했다. advisory 계층이므로
  코드·완료 판정 근거로 사용하지 않았다.
- `make sdd-check` — PASS. CodeGraphContext·GBrain stale은 advisory warning이며
  CodeGraph hard-evidence fingerprint, PM 1:1, memory, 도구·서비스, SDD test는 통과했다.
- `make gate CHANGE=console-adoption-controls` — **BLOCKED at 4/8 Function Logic Map**.
  tasks 0개, review 존재까지는 통과했다. change의 고정 base
  `fc9cf51d498185f42e6ce6ea2a4a64bc17d0bdfb` 이후 공유 worktree에 병행 change의
  기존 함수 수정이 대량 포함되어, 이 change의 기존 map도 stale이고 다른 change의
  함수들은 missing evidence로 판정됐다.
- 이번 메뉴 개정은 template 상수와 신규 test 함수뿐이므로 자체 Function Logic Map은
  명시적으로 not-applicable이다. 병행 change의 수백 개 map을 이 change 산출물로
  복제하거나 base를 현재 dirty worktree로 재고정하면 증거를 왜곡하므로 하지 않았다.
- 따라서 구현·배포는 검증됐지만 이 기록 시점에는 OpenSpec archive와 Full SDD 완료
  보고를 하지 않는다. 병행 change를 각각 정리·archive하고 격리된 base/diff에서
  `make gate CHANGE=console-adoption-controls`를 다시 통과시켜야 한다.
