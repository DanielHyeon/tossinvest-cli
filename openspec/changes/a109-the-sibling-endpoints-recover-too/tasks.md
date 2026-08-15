# Tasks — a109-the-sibling-endpoints-recover-too

등록 2026-08-14, 착수 2026-08-15 (a108 land 3615f793 확인). 입력: a108 review §1 F3·F4
실측 + `analysis/function-logic/` AST 13개 + design.md D1–D5.

## 0. 착수 전 조건

- [x] 0.1 a108 land 확인(3615f793), base 재고정(016da624) + `make sdd-sync` (rc=0,
  GBrain은 advisory).
- [x] 0.2 대상 함수 FLM AST 산출물 13개 생성(발행·회수·Close·기동·소비자). 마크다운
  맵 완성은 각 Teammate가 자기 표면에서 **편집 전** 수행(§1a·§2a).
- [x] 0.3 endpoint별 강등 판정 — design D3 표. 셋 다 강등, 근거는 endpoint별 보수성
  논증 + flock 싱글턴 + 소비자 부재 경로 기존재.
- [ ] 0.4 proposal-freeze 리뷰(gstack 4관점 + 적대 Eng) → review.md §0.

## 1. T1 — transport 발행·회수 (internal/*)

- [ ] 1a FLM 맵 완성(자기 표면 9개: prepare/stat/validate×2계열, Start×3, Close×3)
  + Pre-Edit 선언(High-risk).
- [ ] 1.1 RED: pre-chmod 0700 socket 잔재에서 policy runtime·alert control 기동이 영구
  거부됨을 고정(A1 F3 재현 — umask 077 실측 모양).
- [ ] 1.2 RED: 수락 중인 socket 위 두 번째 Start가 기존 socket을 unlink하고 올라서는
  현재 동작을 고정(탈취 → 거부가 목표 동작).
- [ ] 1.3 RED: staging 잔재(legacy `.position-policy-*`/`.endpoint-*` + 신규 `.s-*`)와
  낯선 엔트리 각각의 기동 결과 고정(잔재=회수, 낯섦=거부).
- [ ] 1.4 GREEN: 이름-독립 staged listen(11자 staging)·전체성 회수·connect probe를
  `internal/positionpolicyrpc`에 추가, 두 socket transport가 사용(D1·D2). command
  endpoint는 staging 위생만(D2a). Close에 listener 직접 닫기(D2b — AlertControlServer
  listener 필드 추가).
- [ ] 1.5 staging basename ≤ 최종 basename(alerts.sock 11자 포함) + sun_path 이식 경계
  103 상수 테스트.
- [ ] 1.6 뮤테이션 원장(t1) — 회수 분기·probe 3갈래·완화 경계 각각을 죽여서 테스트가
  잡는지 기록. 원복은 심볼 수로 확인.
- [ ] 1.7 구현 후 FLM AST·맵 재최신화.

## 2. T2 — 기동 강등·소비자 재부착 (cmd/tossctl/*)

- [ ] 2a FLM 맵 완성(자기 표면: runEngineRun, strategyRuntimeReaderFor) + Pre-Edit 선언.
- [ ] 2.1 RED: 세 Start 실패 시 엔진 기동이 죽는 현재 동작 고정(seam 도입,
  cli_testseams.go 관례 — T1 완료 불요).
- [ ] 2.2 GREEN: 세 fatal 강등(D3) + 보고 일반화(D3a — stderr + 등급표 미등재 obs
  Normal + WithoutCancel goroutine, 금지 3종 주석 유지).
- [ ] 2.3 RED→GREEN: httpapi lazy 재-dial(D4) — nil·sentinel reader의 rate-limit
  재부착, single-flight, 요청 비차단. 엔진 냉부팅 순서·가동 중 재시작 복귀 시나리오.
- [ ] 2.4 defer 순서 정적 핀(D5a) — lock.Release() defer가 endpoint Close defer보다
  먼저 등록됨을 go/parser로 고정.
- [ ] 2.5 뮤테이션 원장(t2) — 강등 경로·rate-limit·재부착 교체 각각.
- [ ] 2.6 구현 후 FLM AST·맵 재최신화.

## 2b. 리뷰 라운드

- [ ] 2b.1 A1(T1 적대)·A2(T2 적대) 독립 리뷰 — 뮤테이션 재현 포함 → review.md §1·§2.
- [ ] 2b.2 Manager 판결 + 수정 라운드 → review.md §3·§4.
- [ ] 2b.3 gstack 전체 리뷰 + 수렴 → review.md §5.

## 3. 게이트

- [ ] 3.1 영향 패키지 `go test -race`(app/engine ≥ `-timeout 25m`) + `go vet`.
- [ ] 3.2 `openspec validate --all --strict` → `make sdd-sync` → `make sdd-check` →
  `make gate CHANGE=a109-the-sibling-endpoints-recover-too` (검증 명령은 파이프 없이).
- [ ] 3.3 롤백 절차 측정(D5b): 구버전 바이너리에 신버전 잔재 모양별 관용 여부 실측 →
  아래 배포 절차에 결과 기록.
- [ ] 3.4 review.md 정산 + PM(STORY-TOS-a109) 동기화 + Manager 독립 검증 기록.

## 배포 절차 (게이트 밖 — 사람 승인 필요)

merge·push·이미지 빌드·컨테이너 재시작은 사용자 승인 후 별도 수행한다. engine lock을
잡는 재시작은 두 시장 폐장 창(KST 05:00–09:00)을 따른다. a109를 가로지르는 롤백
절차는 3.3의 측정 결과가 정본이다: 측정이 "구버전이 신버전 잔재를 무시한다"로 나오면
사전 wipe 불요를 명기하고, 아니면 wipe 대상 디렉터리(`.position-policy-control`,
`.position-policy-runtime`, `.alert-control`)를 a108 D5-3 형식으로 명기한다.
