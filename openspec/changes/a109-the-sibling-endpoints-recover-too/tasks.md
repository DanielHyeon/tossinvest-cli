# Tasks — a109-the-sibling-endpoints-recover-too

등록 2026-08-14, 착수 2026-08-15 (a108 land 3615f793 확인). 입력: a108 review §1 F3·F4
실측 + `analysis/function-logic/` AST 13개 + design.md D1–D5.

## 0. 착수 전 조건

- [x] 0.1 a108 land 확인(3615f793), base 재고정(016da624) + `make sdd-sync` (rc=0,
  GBrain은 advisory).
- [x] 0.2 대상 함수 FLM AST 산출물 **20개** 생성(발행·회수·Close·기동·소비자 13개와
  freeze P1-6 추가분 7개: descriptor 발행 3·보고 1·디렉터리 검증 3). 마크다운 맵 완성은
  각 Teammate가 자기 표면에서 **편집 전** 수행(§1a·§2a).
- [x] 0.3 endpoint별 강등 판정 — design D3 표(freeze P0-4 반영: 모든 논증의 기준선은
  fatal). 셋 다 강등 + 소비자 오귀속 문구 정직화(D3a-2).
- [x] 0.4 proposal-freeze 리뷰(Manager 4관점 + 적대 Eng 별도 컨텍스트) → review.md §0.
  P0 4건·P1 9건 수용 반영, P2 7건 처리 — 판결표 참조.

## 1. T1 — transport 발행·회수 (internal/*)

- [x] 1a FLM 맵 완성(자기 표면 **17개** — 등록문의 "9개"는 슬러그 수를 잘못 센 것이다:
  positionpolicyrpc 8 + app/engine 9). `check_analysis.py` 기준 T1 슬러그 오류 0
  (남은 3건은 T2 표면). Pre-Edit 선언은 `pre-edit-gate-t1.md`(T1-1~T1-4 + 선언된 무변경).
- [x] 1.1 RED: pre-chmod 0700 socket 잔재에서 policy runtime·alert control 기동이 영구
  거부됨을 고정(A1 F3 재현 — 잔재는 명시적 chmod 0700으로 제작). 관측:
  runtime = `preparing position policy runtime socket: ... stale endpoint is not an exact
  0600 Unix socket`, alert = `preparing the alert control socket: ... path is not an exact
  0600 Unix socket`. 완화 폭 핀(0640 거부)은 편집 전후 모두 GREEN이어야 하는 항목으로 동봉.
- [x] 1.2 RED: 수락 중인 socket 위 두 번째 Start가 기존 socket을 unlink하고 올라서는
  현재 동작을 고정(탈취 → 거부가 목표 동작). 관측: 두 endpoint 모두 두 번째 Start가
  **성공**했다(`주인이 수락 중인 socket 위에서 두 번째 기동이 성공했다`). 핀은 거부와
  "거부하면서 주인의 socket을 그대로 둔다"(inode 동일성 + 재수락) 둘 다 잰다.
- [ ] 1.3 RED: staging 잔재(**현행·구버전 공통** CreateTemp `.position-policy-*`/
  `.endpoint-*` + 신규 `.s-*`)와 낯선 엔트리 각각의 기동 결과 고정(잔재=회수,
  낯섦=거부 — socket endpoint 한정, command는 무시 유지 D2a).
- [ ] 1.4 GREEN: 이름-독립 staged listen(11자 staging — hex 절단, D1a)·전체성 회수
  (staging 소유 uid 검증 포함)·connect probe를 `internal/positionpolicyrpc`에 추가,
  두 socket transport가 사용(D1·D2). 아는-이름 집합은 호출자가 넘기고 완전성 테스트로
  고정(D1b). command endpoint는 staging 위생만(D2a). Close에 listener 직접 닫기(D2b —
  AlertControlServer listener 필드 추가). 클라이언트 검증은 정확-0600 유지(P1-3).
- [ ] 1.5 staging basename **길이 11 직접 측정** + 각 최종 socket basename과의 ≤
  관계식 상수 테스트(절대 경로 103 요구 금지 — D1a, Linux 실측 상한 107).
- [ ] 1.6 뮤테이션 원장(t1) — 회수 분기·probe 3갈래·완화 경계 각각을 죽여서 테스트가
  잡는지 기록. 원복은 심볼 수로 확인.
- [ ] 1.7 구현 후 FLM AST·맵 재최신화.

## 2. T2 — 기동 강등·소비자 재부착 (cmd/tossctl/*)

- [ ] 2a FLM 맵 완성(자기 표면: runEngineRun, strategyRuntimeReaderFor,
  reportStrategyProjectionDegraded) + Pre-Edit 선언.
- [ ] 2.1 RED: 세 Start 실패 시 엔진 기동이 죽는 현재 동작 고정(seam은
  `engineStrategyProjectionStart`(engine.go:416) 패키지 var 관례 — T1 완료 불요,
  P1-5).
- [ ] 2.2 GREEN: 세 fatal 강등(D3) + 보고 일반화(D3a — stderr + 등급표 미등재 obs
  Normal + WithoutCancel goroutine, 금지 3종 주석 유지).
- [ ] 2.3 RED→GREEN: httpapi lazy 재-dial(D4 개정) — wrapper는 **live client 포함**
  전체를 감싸고(P0-1), 시도는 백그라운드 single-flight(요청 goroutine에서 dial·probe
  금지, P0-2), 부재/unavailable 화면 구분 보존(nil 검사 2곳 —
  httpapi_reader.go:565·internal/httpapi/strategy_runtime.go:18 — 상태 신호로 교체,
  P1-4), 보고는 상태 전이 1회(P2-4). 냉부팅 순서·가동 중 재시작(새 socket·새 토큰)
  복귀 시나리오 모두.
- [ ] 2.4 defer 순서 정적 핀(D5a) — lock.Release() defer가 endpoint Close defer보다
  먼저 등록됨을 go/parser로 고정.
- [ ] 2.5 소비자 오귀속 문구 정직화(D3a-2): engine_alerts_client_unix.go:33–35,
  internal/console/exit_quarantine.go:227–229 — "엔진 부재 단정"을 "부재 또는 강등"
  안내로. 메시지 텍스트 핀 테스트 포함.
- [ ] 2.6 뮤테이션 원장(t2) — 강등 경로·rate-limit·재부착 교체·문구 핀 각각.
- [ ] 2.7 구현 후 FLM AST·맵 재최신화.

## 2b. 리뷰 라운드

- [ ] 2b.1 A1(T1 적대)·A2(T2 적대) 독립 리뷰 — 뮤테이션 재현 포함 → review.md §1·§2.
- [ ] 2b.2 Manager 판결 + 수정 라운드 → review.md §3·§4.
- [ ] 2b.3 gstack 전체 리뷰 + 수렴 → review.md §5.

## 3. 게이트

- [ ] 3.1 영향 패키지 `go test -race`(app/engine ≥ `-timeout 25m`) + `go vet`.
- [ ] 3.2 `openspec validate --all --strict` → `make sdd-sync` → `make sdd-check` →
  `make gate CHANGE=a109-the-sibling-endpoints-recover-too` (검증 명령은 파이프 없이).
- [ ] 3.3 롤백 절차(D5b): 코드 근거(구버전 회수는 ReadDir 미호출 — P2-3) + 구버전
  바이너리 실측 둘 다 → 아래 배포 절차에 결과 기록.
- [ ] 3.4 review.md 정산 + PM(STORY-TOS-a109) 동기화 + Manager 독립 검증 기록.

## 배포 절차 (게이트 밖 — 사람 승인 필요)

merge·push·이미지 빌드·컨테이너 재시작은 사용자 승인 후 별도 수행한다. engine lock을
잡는 재시작은 두 시장 폐장 창(KST 05:00–09:00)을 따른다. a109를 가로지르는 롤백
절차는 3.3의 측정 결과가 정본이다: 측정이 "구버전이 신버전 잔재를 무시한다"로 나오면
사전 wipe 불요를 명기하고, 아니면 wipe 대상 디렉터리(`.position-policy-control`,
`.position-policy-runtime`, `.alert-control`)를 a108 D5-3 형식으로 명기한다.
