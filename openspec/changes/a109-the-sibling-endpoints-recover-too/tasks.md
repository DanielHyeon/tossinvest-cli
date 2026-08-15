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
- [x] 1.3 RED: staging 잔재(**현행·구버전 공통** CreateTemp `.position-policy-*`/
  `.endpoint-*` + 신규 `.s-*`)와 낯선 엔트리 각각의 기동 결과 고정(잔재=회수,
  낯섦=거부 — socket endpoint 한정, command는 무시 유지 D2a). 관측: 두 socket endpoint
  모두 회수 후 디렉터리에 잔재 3개가 그대로 남고(`[.position-policy-runtime-… .s-e… .s-s…
  endpoint.json runtime.sock]`), 낯선 엔트리가 있어도 기동이 **성공**했다. command는
  자기 staging 잔재가 남았고(err=nil) 이물 무시는 이미 참이다.
- [x] 1.4 GREEN: 이름-독립 staged listen(11자 staging — hex 절단, D1a)·전체성 회수
  (staging 소유 uid 검증 포함)·connect probe를 `internal/positionpolicyrpc`에 추가,
  두 socket transport가 사용(D1·D2). 아는-이름 집합은 호출자가 넘기고 완전성 테스트로
  고정(D1b). command endpoint는 staging 위생만(D2a). Close에 listener 직접 닫기(D2b —
  AlertControlServer listener 필드 추가). 클라이언트 검증은 정확-0600 유지(P1-3).
  산출물: `internal/positionpolicyrpc/private_staging.go`(이식성 있는 이름·위생),
  `private_staging_unix.go`(staged listen·회수·probe). 세 transport 재배선 완료.
- [x] 1.5 staging basename **길이 11 직접 측정** + 각 최종 socket basename과의 ≤
  관계식 상수 테스트(절대 경로 103 요구 금지 — D1a, Linux 실측 상한 107).
  추가 경계: 클라이언트 정확-0600 핀(P1-3), probe 4갈래(생존·ECONNREFUSED·부재·
  owner-write 없음), 이름 집합 방어(빈 descriptor·빈 접두), staging 모양 거부,
  회수 후 디렉터리 부재, 위생의 무해성, 이름 집합 완전성(P2-5).
- [x] 1.6 뮤테이션 원장(t1) — 회수 분기·probe 3갈래·완화 경계 각각을 죽여서 테스트가
  잡는지 기록. 원복은 심볼 수로 확인. 결과: 적용 31건 = **사망 25 · 생존 6**(등가 2·
  경합 3·비root 측정 불가 1). 뮤테이션이 새 핀 3개를 열었다(hard link·디렉터리 위생·
  descriptor 모양) — §1.5 경계 테스트만으로는 그 네 뮤테이션이 전부 살아남았다.
  `mutation-ledger-t1.md`.
- [x] 1.7 구현 후 FLM AST·맵 재최신화. 편집한 3파일의 9개 슬러그 ast.json·risk report를
  재생성했고(`positionpolicyrpc`의 8개는 파일 무변경이라 그대로 유효), 분기 집합이 바뀐
  4개(Start×2·Close×2)의 맵을 편집-후 AST로 다시 썼다: alert Start 15→13, runtime Start
  16→14(`createdControlDir` 조건과 `os.Chmod` 절이 사라지고 회수·재-Mkdir이 들어왔다),
  두 Close 3→5(listener 직접 닫기). `check_analysis.py` 기준 T1 슬러그 오류 0
  (남은 3건은 T2 표면).

## 2. T2 — 기동 강등·소비자 재부착 (cmd/tossctl/*)

- [x] 2a FLM 맵 완성(자기 표면: runEngineRun, strategyRuntimeReaderFor,
  reportStrategyProjectionDegraded) + Pre-Edit 선언.
- [x] 2.1 RED: 세 Start 실패 시 엔진 기동이 죽는 현재 동작 고정(seam은
  `engineStrategyProjectionStart`(engine.go:416) 패키지 var 관례 — T1 완료 불요,
  P1-5).
- [x] 2.2 GREEN: 세 fatal 강등(D3) + 보고 일반화(D3a — stderr + 등급표 미등재 obs
  Normal + WithoutCancel goroutine, 금지 3종 주석 유지).
- [x] 2.3 RED→GREEN: httpapi lazy 재-dial(D4 개정) — wrapper는 **live client 포함**
  전체를 감싸고(P0-1), 시도는 백그라운드 single-flight(요청 goroutine에서 dial·probe
  금지, P0-2), 부재/unavailable 화면 구분 보존(nil 검사 2곳 —
  httpapi_reader.go:565·internal/httpapi/strategy_runtime.go:18 — 상태 신호로 교체,
  P1-4), 보고는 상태 전이 1회(P2-4). 냉부팅 순서·가동 중 재시작(새 socket·새 토큰)
  복귀 시나리오 모두.
- [x] 2.4 defer 순서 정적 핀(D5a) — lock.Release() defer가 endpoint Close defer보다
  먼저 등록됨을 go/parser로 고정.
- [x] 2.5 소비자 오귀속 문구 정직화(D3a-2): engine_alerts_client_unix.go:33–35,
  internal/console/exit_quarantine.go:227–229 — "엔진 부재 단정"을 "부재 또는 강등"
  안내로. 메시지 텍스트 핀 테스트 포함.
- [x] 2.6 뮤테이션 원장(t2) — 강등 경로·rate-limit·재부착 교체·문구 핀 각각.
- [x] 2.7 구현 후 FLM AST·맵 재최신화.

## 2b. 리뷰 라운드

- [x] 2b.1 A1(T1 적대)·A2(T2 적대) 독립 리뷰 — 뮤테이션 재현 포함 → review.md §1·§2.
  A1: P1 3·P2 6, 뮤테이션 7건 원장 대조(3건 반증), RED 3커밋 축자 재현.
  A2: P0 0·P1 2·P2 11, 뮤테이션 5건 재현 전부 일치, RED 2커밋 재현, D5b 독립 재현.
- [x] 2b.2 Manager 판결 + 수정 라운드 → review.md §3·§4. T1-fix F1~F6(8커밋)·
  T2-fix F1~F9(12커밋) — 전 항목 판결 후 위임, 전부 GREEN·원장 정산 완료.
- [ ] 2b.3 gstack 전체 리뷰 + 수렴 → review.md §5.

## 3. 게이트

- [ ] 3.1 영향 패키지 `go test -race`(app/engine ≥ `-timeout 25m`, **internal/console
  ≥ `-timeout 20m`** — 기본 10m 경계 582~693s 실측, A2 P1-2) + `go vet`. 병렬 실행
  금지(CPU 기아로 위양성 FAIL 이력).
- [ ] 3.2 `openspec validate --all --strict` → `make sdd-sync` → `make sdd-check` →
  `make gate CHANGE=a109-the-sibling-endpoints-recover-too` (검증 명령은 파이프 없이).
- [x] 3.3 롤백 절차(D5b): 코드 근거(구버전 production에 `os.ReadDir` 0건 — T2·A2 각각
  독립 확인) + 실측(T2: 3 Start × 6모양 = 18/18 관용·잔재 생존, A2: 표본 3모양 독립
  재현 일치) → 아래 배포 절차에 기록 완료.
- [ ] 3.4 review.md 정산 + PM(STORY-TOS-a109) 동기화 + Manager 독립 검증 기록.

## 배포 절차 (게이트 밖 — 사람 승인 필요)

merge·push·이미지 빌드·컨테이너 재시작은 사용자 승인 후 별도 수행한다. engine lock을
잡는 재시작은 두 시장 폐장 창(KST 05:00–09:00)을 따른다.

- **land 단위는 T1+T2 합본이다**(A1 P1-C). T1의 낯선-엔트리 거부는 T2의 강등이 받는
  것을 전제로 설계됐다 — T1만 land하면 이물 하나가 엔진 전체를 fatal 영구 기동
  루프(장중 손절 없음)로 되돌린다. `feat/a109-…` 브랜치는 T2 병합 후에만 main 후보다.
- **롤백(a109를 가로지르는): 사전 wipe 불요.** 근거 — 코드: 구버전 회수·발행 경로는
  `os.ReadDir`를 부르지 않아 신버전 `.s-*` 잔재를 보지 못한다(T2·A2 독립 확인).
  실측: 구버전 바이너리 기동 18/18 관용(T2), 표본 3모양 독립 재현 일치(A2).
  단 두 가지를 함께 기록한다(A2 P2-11): ① 신버전 잔재는 구버전 부팅을 가로질러
  **그대로 남고**, 구버전 Close는 그 잔재 때문에 control 디렉터리 rmdir을 조용히
  실패한다(무해, 로그 없음). ② a109 재배포 시 첫 부팅의 회수가 잔재를 치운다.
- **배포 후 확인**: 엔진 로그에서 세 endpoint의 강등 보고 유무 확인(운영 control
  디렉터리의 이물 현존 여부는 개발 세션에서 확인 불가 — freeze 잔존 리스크). 강등
  보고가 있으면 지목된 엔트리를 제거 후 재시작한다.
