# a114 리뷰 기록

## §0 proposal-freeze (2026-09-25)

위험 등급: Normal(콘솔 lifecycle — 단 Preview/Apply/격리 해제 **명령 client** 의 획득 경로라 적대 Eng 포함).
보이스: Teammate 셀프 4관점 + **독립 적대 Eng 리뷰어 1**(Opus 서브에이전트, 별도 컨텍스트, 읽기 전용).
`autoplan` 은 대화형 파이프라인이라 자율 세션에서 독립 서브에이전트로 대체했다(Manager 지시의 대체 경로).

### 셀프 4관점

- **CEO/Product** — 범위: a109 선언된 생략 P2-7 한 건. 합본 안 함(design 0.1). UI 마찰 추가 없음.
- **Eng** — AST 로 lifecycle client 가 격리 해제도 싣는다는 것(`quarantineClient` B1)을 먼저 찾았다 — wrapper 가
  세 메서드를 빠뜨리면 격리 해제 버튼이 사라진다.
- **Security** — client 권한 불변(같은 descriptor 검증·loopback·bearer). 명령 재전송 금지를 요구로.
- **QA** — 진짜 엔진 endpoint 로 늦은 기동·재시작, 가짜 client 로 기전, 콘솔 실측.

### 독립 적대 Eng 리뷰 — P0 0 · P1 3 · P2 7 + 시험 공백

| id | 발견 | 판결 → 반영 |
| --- | --- | --- |
| P1-1 | 여섯 메서드가 자리 하나를 공유 — 엔진 `internal`(500, untyped remote failure)을 탈착으로 읽으면 List 성공과 번갈아 전이 로그가 깜빡이고 멀쩡한 client 가 매 렌더 교체 | 수용 → 「코드가 붙은 모든 rpcError = 답」, 단 토큰 거절 문구는 탈착. positionpolicyrpc 타입 추가는 표면 밖이라 문구 상수 셋 + 원본 리터럴 핀(issues S1). 시험: 깜빡임·토큰 거절·멈춘 엔진 timeout |
| P1-2 | 렌더 말고는 아무도 깨우지 않는다(콘솔엔 publisher 없음) · 부팅이 lastTry 를 찍는지 미정 | 수용 → 간격 주기 펌프(탈착 동안만 wake), 부팅 해석은 lastTry 안 찍음. 시험: 렌더 없이 붙기·첫 wake 자유. a081 캐시 지연은 issues R2 |
| P1-3 | spec delta 가 읽기만 다룬다 | 수용 → 명령 최대 1회·답한 거절 비탈착·미배선 표시 조건·격리 해제 발견·부착 전 ErrUnwired 금지를 SHALL/시나리오로 |
| P2 | 밀려난 client 해제가 운영에서 no-op | 수용(기록) → design·issues S2 에 숨기지 않고 적음 |
| P2 | 거절 화면의 「아무것도 변경되지 않았다」 | 기록 → issues R1(표면 밖) |
| P2 | a079 이전 엔진 client → wrapper 안에서 ErrUnwired | 이미 구현 · 시험(`a114PreA079`) |
| P2 | client timeout = DeadlineExceeded, ctx 확인이 취소와 가른다 | 수용 → `TestAHungEngineTimeoutIsADetachment` |
| P2 | runtime descriptor reader 가 탈착 중에도 렌더마다 불림 | 수용 → spec 에서 명시 제외, design 비목표에 비용 기록 |
| P2 | FLM B34/B37 둘 다 :408 | 확인 — AST 가 else-if 사슬을 같은 줄의 else(B34)·if(B37) 두 분기로 연다. 문구 유지 |
| P2 | 규칙 13 부착 절반 | 수용 → issues R3 not-applicable 사유 |
| 시험 공백 | M35 누락, M31/33/34/37/38 사유 없음, 재전송 합산, 변이 4종 추가 | 수용 → 원장에 M35·C1–C8, not-applicable 사유 명기, 재전송 시험은 옛·새 client 양쪽을 센다 |

생존한 공격(리뷰어): Wired 뒤집힘의 fail-closed 방향, 엔진 단일 write connection 부하 없음(detached 는 무네트워크,
시도는 DB 안 쓰는 health GET), nil-포인터 함정 회피, 이중 전송 경로 없음(Go transport 는 nothingWritten 일 때만
POST 재시도), 401 재부착 정당, a079 이전 404→ErrUnwired 를 답으로 보는 것 정당, a115 분리 정당, 줄 인용 일치.

freeze 종결 — 구현 착수.

## Pre-Edit Gate (2026-09-25)

- change id / task id: a114-the-console-reattaches-its-lifecycle / 1.1–1.2
- 대상 심볼: `main.runConsole`(lifecycle dial 블록 → 한 줄) · 신규 `positionPolicyLifecycleAttachment`·
  `consolePositionPolicyCommanderFor`(새 파일) · 테스트 핀 2개(`console_test.go`, 읽는 소스 확장)
- CodeGraph: `analysis/code-context/codegraph-baseline.md` — runConsole 호출자 newConsoleCmd 1, commander 생성
  호출자 runConsole + 테스트 1, affected `--filter` 860(기본 0)
- CodeGraphContext/reconciliation: CGC not-applicable, impact 52 는 이름 해소 과대 추정 — `evidence-reconciliation.md`
- 기존 동작 근거: base `634cf3c5` AST 4 번들(runConsole·quarantineClient·handlePositionManagement·decoratePositionRows)
- FLM / Branch Test Map: `analysis/function-logic/`(runConsole · 소비자 3 · 테스트 핀 2)
- upstream 상속 테스트 영향: no — 콘솔은 TossOS 신규
- 실패 테스트 선행: yes — `TestRunConsoleNeverDialsTheLifecycleItself` 가 base 에서 FAIL("positionpolicyrpc.Dial 을
  1번 직접 부른다", 격리 프로브 파일로 관측), 나머지는 컴파일 RED(undefined 심볼)
- 설정·DB·journal 변경과 rollback: 없음 — rollback = 커밋 revert
- 안전 불변식 §0: 통과 — 주문·손절·원장 무접촉, 명령은 재전송하지 않음(시험·변이 C6), UI 마찰 무추가

## 콘솔 실측 (규칙 13, 2026-09-25)

격리 config(`mktemp -d /tmp/a114-console-XXXX`, 0700, 엔진 없음) · `tossctl --config-dir <그것> console --port
18471` · 세션 링크로 로그인(303) · `GET /position-management` → **200**. 본문:

- a114 바이너리: `불러오기 실패: 엔진 포지션 정책 control plane 에 붙어 있지 않다 — 엔진이 내려갔거나 아직 기동
  중이거나 이 표면 없이 강등 부팅했을 수 있다(엔진 로그의 강등 보고를 확인하라). 엔진이 endpoint 를 발행하면
  콘솔 재시작 없이 다시 붙는다` · 「배선되지 않아」 notice **없음**.
- base 바이너리(`634cf3c5`, 같은 config): `engine-owned PositionPolicyCommander가 배선되지 않아 조회만 가능하다.`
- 버튼·폼은 누르지 않았다(GET 두 번뿐). 부착 절반은 issues R3.

## §1 구현 후 리뷰 (2026-09-25 — 648df8ef)

구성: 독립 서브에이전트(Opus, 별도 컨텍스트, 읽기 전용 — gstack `review` pre-landing 관점 대체; codex 적대 패스는
이 자율 세션에서 쓰지 않았다). `-race -count=3` 통과·잠금 중첩 없음·펌프 종료·부팅 시간 불변·재전송 경로 없음을 확인했고,
새 변이 7개 중 2개 생존을 찾았다.

| id | 발견 | 판결 → 반영 |
| --- | --- | --- |
| P2-1 | 사유 없는 JSON 거절도 「remote failure」 문구라 답으로 읽혀 자리가 죽은 client 에 묶일 수 있다 | 수용 → 사유가 비면 탈착(`engineRemoteFailure`) + 시험 + 변이 P21. 타입 도입은 issues S1 후속 |
| P2-2 | 펌프 틱 = 간격이라 지터로 15~19% 틱이 건너뛰어져 60s 공백 | 수용 → 틱을 간격의 절반으로. 결정적 시험 없음(원장 명기) |
| P2-3 | 하네스가 rc 로만 판정 — TMPDIR 정리 실패가 대조군을 rc=1 로 만듦 | 수용 → FAIL/ok 줄로 판정(CAUGHT/SURVIVED/BROKEN) |
| P2-4 | 밀려난 client 를 강제로 닫으면 진행 중인 명령을 끊는다(오늘 no-op) | 수용 → 코드 주석·issues S2 에 「유휴 연결만」 |
| N2 생존 | 성공이 `failed` 를 안 지워도 초록 | 수용 → `TestARecoveredSeatStopsAsking` |
| N7 생존 | position policy 문구 제거가 초록 | 수용 → `TestAnInternalListErrorDoesNotFlapTheSeat` |
| N4 생존 | wake 의 ctx 검사 제거 | 수용(기록) — 종료 뒤 시도는 즉시 실패, 무해 |

화면 영향(리뷰어 확인): 설정 요약의 정책 행이 「읽지 못함 — <detached 문장>」(의도, 다소 김), 포트폴리오는 탈착 동안
「관리 여부 불명」(issues R2), 격리 배지는 List 성공 뒤에만 조회되어 새 배너 없음.

## 검증 기록 (2026-09-25)

- `go test -race -count=1 ./cmd/tossctl/` → ok 173.9s (1244be95 워킹트리) · a114 대상 `-race -count=3` ok
- `make vet` rc 0 · `make lint` rc 0
- `make test`(`go test -timeout 30m ./...`) → **ok 99 · no test files 9 · FAIL 0** (GOCACHE 를 /mnt/D 로 — 루트
  파일시스템이 한때 100% 였다; a113 판의 `make test` 에서는 그 때문에 `internal/app/engine`
  `TestTheDriverFoldsAndAdoptsInOneCycle` 가 "database or disk is full" 로, `tools/a112-mb-us-source` 가 빌드 신원
  불일치로 실패했고 이번 판에서는 둘 다 ok)
- `check_analysis.py --change a114-…` rc 0 · 착지 `1244be95`(병행 세션의 더러운 워킹트리 때문에 같은 커밋의 shared
  clone 에서 도구가 계산한 값을 옮겼다)
