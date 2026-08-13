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
- [x] 0.4 **High-risk Pre-Edit 선언** — T1·T2가 각자 첫 편집 전에 `pre-edit-gate.md`에
  대상 함수·실패 테스트 선행 계획을 남긴다. 엔진 기동 경로는 면제 불가.

## 1. 겹1 RED — 사고 상태의 재현 (T1)

전부 tmpdir fixture. 실계좌·실컨테이너 접촉 없음.

- [x] 1.1 **S1 재현(이번 사고)**: dir + descriptor만 → 현재 `Start`가
  "stale endpoint is incomplete"로 실패함을 고정하는 RED. 사고 당일 `endpoint.json`
  형태(schemaVersion v1·pid·token)를 fixture로 쓴다.
- [x] 1.2 S0(빈 디렉터리) RED, S2(socket만·죽은 주인) RED — 각각 현재 코드가 거부함을 먼저
  고정한다.
- [x] 1.3 **PID 재사용 RED(D2 핀)**: descriptor PID = 현재 테스트 프로세스 PID(확실히 생존),
  socket은 존재하되 listener 없음(사망) → 현재 코드가 "projection owner is still alive"로
  영구 거부함을 고정.
- [x] 1.4 거부 유지 핀: S4(낯선 엔트리) 거부, S2/S3에서 **수락 중인 socket**(테스트가 실제
  listener를 연다) 거부, 소유권·perm·symlink 검사 거부 — 새 코드에서도 그대로 통과해야
  하는 보안 속성.

## 2. 겹1 GREEN — 회수의 전체성 (T1)

- [x] 2.1 D1 표대로 `reclaimStaleControlDirectory` 재작성: S0·S1·S2-사망·S3-사망 회수,
  S2/S3-생존·S4·검증 실패 거부. 표의 행마다 테스트 하나.
- [x] 2.2 D2: connect probe 생존 판정 구현, `processAlive` 판정 제거. 1.3의 핀이 GREEN이
  되고, `processAlive`로 되돌리는 뮤테이션에 실패하는 것을 원장에 기록.
- [x] 2.3 Close·reclaim의 제거 경합 양성 확인 테스트(ENOENT 용인) — D2 경합 창 문단의 근거.
- [x] 2.4 뮤테이션 원장: 각 상태 행의 판정을 하나씩 뒤집어(회수→거부, 거부→회수) 해당
  테스트만 죽는 것을 확인·기록한다. 원복은 심볼 대조로 확인.
- [ ] 2.5 **네 endpoint crash-shape 핀(D5)**: policy command(TCP·descriptor만)·policy
  runtime·alert control 각각의 transport에 "부분 잔재 후 기동 성공" 테스트. 핀이 결함을
  드러내면 **작업을 멈추고 Manager에 보고**(같은 패턴 수정의 스코프 승인).
  — 2026-08-14 A1 판정: 이 핀들은 관용이 자명한 모양만 만들어 실패할 수 없었다(review.md §1 F4).
  **미완으로 남긴다** — 사고급 모양 핀은 a109 tasks 1절이 소유한다(design D5-2).

## 3. 겹2·겹3 — 소비자는 잔재로 죽지 않는다 (T2)

- [x] 3.1 겹2 RED: projection `Start` 실패 주입(seam) 시 현재 `runEngineRun`이 종료함을
  고정 → GREEN: 강등 + critical 알림 + 루프 계속. 알림은 기존 alert 배관 관행(D3)을 따르고
  기동 1회 유실형이 아니어야 한다.
  — 완료했으나 알림 형태는 A2 리뷰로 **반전**됨(D3-2): outbox 철회는 6.7이 수행한다.
- [x] 3.2 겹2 안전 핀: 강등 기동에서도 ① journal flock 싱글턴 불변(둘째 엔진 거부),
  ② a102 ready 신호(marker ready_at) 발행 시점 불변, ③ automation interlock 평가 순서
  불변임을 테스트로 고정.
- [x] 3.3 겹3 RED: descriptor 존재 + dial 실패 fixture에서 현재 `runHTTPAPI`가 종료함을
  고정 → GREEN: `strategyRuntime = nil` 강등 기동, descriptor-부재 경로와 관측 동작 동일.
  비-NotExist stat 오류는 fatal 유지(그 구분 테스트 포함).
- [x] 3.4 뮤테이션 원장: 강등을 fatal로 되돌리는 뮤테이션에 3.1·3.3이 죽는 것을 확인·기록.

## 4. 리뷰·게이트·검증

- [x] 4.1 A1(적대 리뷰, T1 산출물)·A2(적대 리뷰, T2 산출물) — 반증 지향: 회수가 치우면
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

## 6. Fix 라운드 (A1·A2 적대 리뷰 판정, 2026-08-14 — design D1-2~D5-2가 계약)

**4.4(gstack) 전에 완료한다.** T1-fix·T2-fix는 원래 소유 경계를 유지한다.

### 6a. T1-fix (`internal/strategyprojectionrpc`)

- [ ] 6.1 RED: pre-chmod socket 잔재(umask 077 → 0700)의 영구 거부와 0바이트·잘린
  descriptor의 영구 거부를 A1 절차대로 재현·고정한다.
- [ ] 6.2 GREEN(D1-2): descriptor·socket 모두 stage+rename 발행(socket은 임시 이름
  bind→chmod 0600→rename), staging 잔재를 자기 잔재로 회수, 검증-사망 socket perm
  검사를 `perm&0o077==0`으로 좁게 완화, descriptor 내용 파싱 실패는 사망 입증 시 회수
  (형식 검사 유지). 6.1의 RED가 GREEN이 된다.
- [ ] 6.3 GREEN(D2-2): `SetUnlinkOnClose(false)` + `Close`가 `s.listener`를 명시적으로
  닫고 최종 경로 제거를 소유한다. A1의 늦은-unlink 재현을 회귀 핀으로 결정화한다
  (후계자 socket 보존).
- [ ] 6.4 GREEN(D4-2 T1 몫): `Dial`에 connect probe — S3(socket 잔존 사망)에서 즉시
  오류. S3 fixture 테스트.
- [ ] 6.5 기존 핀 재분류: `TestStartRefusesUnsafeLeftoverShapes`의 socket-0600·반쯤-descriptor
  행을 회수 테스트로 이동하고 M7 해석을 정정한다. 뮤테이션 원장을 새 판정 뒤집기로
  갱신한다(stage+rename 제거·perm 완화 폭 넓힘·probe 제거 각각이 어떤 테스트를 죽이는지).
- [ ] 6.6 A1 P3 소거: `os.Remove(controlDir)`의 `ErrNotExist` 용인 비대칭, 새-주인 경합의
  flock 인용 주석, 절 단위 미핀 3건(symlink·uid·SameFile) 핀 추가.

### 6b. T2-fix (`cmd/tossctl`)

- [ ] 6.7 D3-2: `EnqueueAlert` 제거 — 강등 보고는 stderr + obs Normal 이벤트 로그.
  proc-instance dedup 토큰 기계 제거. 핀 2종 추가: ① 강등 기동이 outbox 미전달 행을
  만들지 않는다, ② 강등 기동 후 재기동에서 entry gate가 그것 때문에 잠기지 않는다
  (A2가 요구한 다음-부팅 측정 — 전달 루프가 있는 harness로).
- [ ] 6.8 D4-2 T2 몫: 비-NotExist stat 오류를 경고+강등으로(콘솔 패리티, 기존 fatal 핀
  반전), `httpapi_reader.go`의 strategy Read 실패를 dormant 흡수로(집계 스냅샷 생존
  테스트), S3 fixture(6.4의 Dial probe 위에서).
- [ ] 6.9 원장·문서 정정: execgw 선례 오독 정정, **비-unix 위험 항목 삭제**(A2 실측 —
  `flock_other.go:16` `ErrLockUnsupported`가 1단계에서 기동을 막아 7단계 도달 불가),
  ready 테스트의 실측 범위를 이름/주석으로 정직화, RED 재구성(커밋 순서 부재) 기록,
  뮤테이션 원장 갱신(M2·M3 등 outbox 계열 재작성).

### 6c. Manager

- [x] 6.10 design D1-2~D5-2 개정, spec delta 재작성(요구 1 좁힘 + 보고의 게이트 비연결
  SHALL NOT + httpapi S3·집계 생존), tasks §6 편성 — 이 커밋.
- [x] 6.11 a109 등록(형제 endpoint 셋의 동일 결함 — A1 실측 승계) + 2.5 핀 재라벨은
  a109 tasks로 승계 명시.
- [x] 6.12 A1·A2 라운드 테이블과 판정을 review.md에 기록.
