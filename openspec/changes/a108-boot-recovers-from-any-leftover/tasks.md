# Tasks — a108-boot-recovers-from-any-leftover

역할: Manager(Fable)는 openspec·검토·완료 검증만. 구현·테스트는 T1·T2(Opus Teammate),
각각 적대 리뷰(A1·A2) 후 gstack 리뷰, 마지막에 Manager 독립 검증.

**0절이 끝나기 전에 1절을 시작하지 않는다. T1(1~2절)과 T2(3절)는 파일 소유가 겹치지
않으므로 병행 가능하되, 각자 자기 소유 파일만 편집·스테이징한다.**

- T1 소유: `internal/strategyprojectionrpc/**`, (3.4가 결함을 드러내면)
  `internal/app/engine/{alert_control*,position_policy*}` — 단 그 경우 착수 전 Manager 보고.
- T2 소유: `cmd/tossctl/engine.go`, `cmd/tossctl/httpapi.go`와 각 테스트 파일.

## 0. 착수 전 조건 (Manager — 완료)

- [x] 0.1 사고 실측 기록: 호스트 재부팅 23:35:19 KST, `.strategy-runtime-read/`에
  `endpoint.json`만 잔존(mtime 14:33Z), 엔진 exit 1 루프, httpapi crash loop,
  다른 세 endpoint는 같은 잔재를 통과. proposal.md에 정리.
- [x] 0.2 FLM 7건 생성(`analysis/function-logic/`): start·reclaimstalecontroldirectory·
  server.close·dial·processalive·runenginerun·runhttpapi. 문서의 분기 주장은 AST 열거 기준.
- [x] 0.3 base 고정(`base-commit.txt`) — 브랜치 `feat/a108-boot-recovers-from-any-leftover`.
- [ ] 0.4 **High-risk Pre-Edit 선언** — T1·T2가 각자 첫 편집 전에 `pre-edit-gate.md`에
  대상 함수·실패 테스트 선행 계획을 남긴다. 엔진 기동 경로는 면제 불가.

## 1. 겹1 RED — 사고 상태의 재현 (T1)

전부 tmpdir fixture. 실계좌·실컨테이너 접촉 없음.

- [ ] 1.1 **S1 재현(이번 사고)**: dir + descriptor만 → 현재 `Start`가
  "stale endpoint is incomplete"로 실패함을 고정하는 RED. 사고 당일 `endpoint.json`
  형태(schemaVersion v1·pid·token)를 fixture로 쓴다.
- [ ] 1.2 S0(빈 디렉터리) RED, S2(socket만·죽은 주인) RED — 각각 현재 코드가 거부함을 먼저
  고정한다.
- [ ] 1.3 **PID 재사용 RED(D2 핀)**: descriptor PID = 현재 테스트 프로세스 PID(확실히 생존),
  socket은 존재하되 listener 없음(사망) → 현재 코드가 "projection owner is still alive"로
  영구 거부함을 고정.
- [ ] 1.4 거부 유지 핀: S4(낯선 엔트리) 거부, S2/S3에서 **수락 중인 socket**(테스트가 실제
  listener를 연다) 거부, 소유권·perm·symlink 검사 거부 — 새 코드에서도 그대로 통과해야
  하는 보안 속성.

## 2. 겹1 GREEN — 회수의 전체성 (T1)

- [ ] 2.1 D1 표대로 `reclaimStaleControlDirectory` 재작성: S0·S1·S2-사망·S3-사망 회수,
  S2/S3-생존·S4·검증 실패 거부. 표의 행마다 테스트 하나.
- [ ] 2.2 D2: connect probe 생존 판정 구현, `processAlive` 판정 제거. 1.3의 핀이 GREEN이
  되고, `processAlive`로 되돌리는 뮤테이션에 실패하는 것을 원장에 기록.
- [ ] 2.3 Close·reclaim의 제거 경합 양성 확인 테스트(ENOENT 용인) — D2 경합 창 문단의 근거.
- [ ] 2.4 뮤테이션 원장: 각 상태 행의 판정을 하나씩 뒤집어(회수→거부, 거부→회수) 해당
  테스트만 죽는 것을 확인·기록한다. 원복은 심볼 대조로 확인.
- [ ] 2.5 **네 endpoint crash-shape 핀(D5)**: policy command(TCP·descriptor만)·policy
  runtime·alert control 각각의 transport에 "부분 잔재 후 기동 성공" 테스트. 핀이 결함을
  드러내면 **작업을 멈추고 Manager에 보고**(같은 패턴 수정의 스코프 승인).

## 3. 겹2·겹3 — 소비자는 잔재로 죽지 않는다 (T2)

- [ ] 3.1 겹2 RED: projection `Start` 실패 주입(seam) 시 현재 `runEngineRun`이 종료함을
  고정 → GREEN: 강등 + critical 알림 + 루프 계속. 알림은 기존 alert 배관 관행(D3)을 따르고
  기동 1회 유실형이 아니어야 한다.
- [ ] 3.2 겹2 안전 핀: 강등 기동에서도 ① journal flock 싱글턴 불변(둘째 엔진 거부),
  ② a102 ready 신호(marker ready_at) 발행 시점 불변, ③ automation interlock 평가 순서
  불변임을 테스트로 고정.
- [ ] 3.3 겹3 RED: descriptor 존재 + dial 실패 fixture에서 현재 `runHTTPAPI`가 종료함을
  고정 → GREEN: `strategyRuntime = nil` 강등 기동, descriptor-부재 경로와 관측 동작 동일.
  비-NotExist stat 오류는 fatal 유지(그 구분 테스트 포함).
- [ ] 3.4 뮤테이션 원장: 강등을 fatal로 되돌리는 뮤테이션에 3.1·3.3이 죽는 것을 확인·기록.

## 4. 리뷰·게이트·검증

- [ ] 4.1 A1(적대 리뷰, T1 산출물)·A2(적대 리뷰, T2 산출물) — 반증 지향: 회수가 치우면
  안 되는 것을 치우는 입력, 강등이 가리는 실패, 경합 창. 리뷰 결과는 `review.md`에 기록.
- [ ] 4.2 편집한 기존 함수 FLM **편집 후 재생성**(SHA-256 일치),
  `python3 tools/logic-map/check_analysis.py --change a108-boot-recovers-from-any-leftover`.
- [ ] 4.3 영향 패키지 `go test -race` + `go vet`
  (`strategyprojectionrpc`, `strategyprojection`, `app/engine`, `cmd/tossctl`).
- [ ] 4.4 gstack 독립 리뷰(High-risk — adversarial Eng voice 필수), Fix-First 처리.
- [ ] 4.5 `openspec validate --all --strict` → `make sdd-sync` → `make sdd-check` →
  `make gate CHANGE=a108-boot-recovers-from-any-leftover`.
- [ ] 4.6 Manager 독립 검증: 뮤테이션 스팟체크 재현 ≥2건(팀메이트당 1), 사고 상태 fixture
  재실행, tasks↔design 대조, STORY-TOS-a108 measured 기입.
- [ ] 4.7 `review.md`에 생략 항목 `not-applicable` 사유 명시 — 침묵한 생략 금지.

## 5. 배포와 사후

- [ ] 5.1 배포 전 main `SchemaVersion` 대조(이 change는 journal 무접촉 — 불일치 시 중단).
- [ ] 5.2 `make image CHANGE=a108-…` → 두 시장 닫힌 창(KST 05:00~09:00) 원칙. 단, **지금
  엔진이 다운이라면 창 규칙보다 보호 복원이 우선**(사람 판단·승인으로 즉시 배포 가능 —
  사고 당일 0.10b 선례).
- [ ] 5.3 배포 후 실측: 잔재 없는 정상 부팅 + (가능하면) S1 잔재를 남긴 재기동에서 자가
  회수 확인, httpapi healthy, 콘솔 전략 화면 정상.
- [ ] 5.4 memory retain + 운영 문서에 "잔재 수동 제거 불필요" 반영.
