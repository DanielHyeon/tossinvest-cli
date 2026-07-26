# issues.md: add-engine-runtime

구현 중 발견·판단 기록. WORKFLOW.md "예외 경로" 분류를 따른다 — ① blocking ②
safe local ③ editorial. 아래는 전부 ②(스펙 의도가 명백한 보완) 또는 기록이며,
blocking 은 없었다.

## §0 검토 기록 (task 3.1)

구현 완료 시점에 최상위 안전 불변식 9개를 대조했다.

| §0 | 항목 | 이 change의 상태 |
|---|---|---|
| 0.1 | 승인 없는 LIVE 주문 side-effect 금지 | **새 mutation 경로 0.** 이 change가 추가한 주문 경로는 없다. 루프 3종은 전부 landed 코드이고, 이 change는 그 호출자를 만들었을 뿐이다. 실제 기동은 인터록 조항 6(ProtectionReady)이 상수 미충족이라 어떤 머신에서도 불가능하며, 그 사실 자체를 `TestTheProtectionClauseIsEnumerated`가 고정한다. 라이브 경로는 fake 인터록(패키지 내부 test-only seam) 하에서만 도달 가능하고, 자동 테스트는 전부 httptest·fake다 |
| 0.2 | OFF = upstream 동작 보존 | 게이트 OFF는 `engine run`이 거부하고 종료한다. 기존 명령은 하나도 바뀌지 않았다(`root.go`에 leaf 추가 1건). upstream 상속 테스트 회귀 0 |
| 0.3 | 손절·비상 청산 즉시성 약화 금지 | exit 관측 루프·flatten 경로 모두 무변경. 감독 2층은 **정지 방향으로만** 작동한다: 비정상 반환은 전체 정지, 지속 열화는 ENTRY_BLOCKED(신규 진입 차단 전용, RISK_REDUCING 청산은 계속 허용). 종료 규율도 "루프 취소 후 완주 대기"이므로 진행 중인 청산을 끊지 않는다 |
| 0.4 | rate limit 예산 계상 | 새 endpoint 0. 루프 3종의 예산은 landed 문서(reconcileloop.go 0.10–0.12 req/s, exitloop.go 0.2 req/s, filldetect 3s 폴링)에 이미 계상되어 있고, 이 change는 그것을 **처음 실제로 쓰게** 만든다. 조립이 추가한 읽기는 스냅샷의 홀딩 1회(raw 경로 선호 — 두 shape를 두 번 읽지 않도록 `AccountSweep` 하나로 통일)와 통화 1개의 잔고 1회뿐이며, 둘 다 Collector의 기존 계상 안에 있다 |
| 0.5 | 운영 설정 변경은 audit 추적 | 이 change는 운영 설정을 만들지도 바꾸지도 않는다. 게이트 flip은 콘솔 밖에 그대로 있다(`TestTheConsoleDecidesNothingAboutTheGate`). 자동 강화 2종은 journal의 operating-mode 이력에 durable하게 남는다 |
| 0.6 | 원장·journal 스키마 변경 | **스키마 변경 0.** 새 테이블·컬럼·마이그레이션 없음 |
| 0.7 | 운영 토글 flip은 사람 승인 | autostart 스크립트는 저장소에 **준비만** 했다. 설치·활성화는 게이트 ON 승인 절차의 항목이며, 스크립트 상단과 `TestTheScriptIsPreparedAndNotInstalled`가 그 경계를 지킨다. 이 머신에는 아무것도 설치하지 않았다 |
| 0.8 | change scope 밖 주문·위험·원장 코드 변경 금지 | 변경한 기존 파일은 (a) journal/operating_mode.go — 트리거 표 additive, (b) obs/event.go — 이벤트 2종 additive, (c) app/engine/reconcileloop.go — 실패 카운터(판정 없음), (d) app/engine/reads.go — 읽기 메서드 1개 추가, (e) filldetect/hints.go — Run의 기존 검사를 `Validate()`로 추출. 전부 scope 안이고 전부 additive |
| 0.9 | 손절·익절·사이징은 보수 방향만 | 해당 로직 무변경 |

**부분 생존 금지의 방향**: 감독 2층은 둘 다 "덜 하는" 쪽으로만 움직인다. ①은
엔진을 통째로 멈추고, ②는 계좌를 ENTRY_BLOCKED로 조인다. 어느 쪽도 주문을 새로
내지 않고, 어느 쪽도 청산을 막지 않는다.

## Pre-Edit 선언 (High-risk 경로 수정분)

```text
Pre-Edit Gate:
- change id / task id: add-engine-runtime / T1.3
- 대상 심볼: engine.ReconcileDriver.RunOnce (named return + 실패 카운터),
             journal.AutomaticTriggers / TargetModeForTrigger (표 2행 추가)
- 기존 동작 파악 근거: reconcileloop_test.go(718줄, RunOnce 직접 구동),
             adoption_manage_forward_test.go, exit_e2e_test.go,
             operating_mode_direction_test.go(트리거 전수 순회),
             operating_mode_test.go(동시 강화)
- upstream 상속 테스트 영향: no (전부 TossOS 고유 경로)
- 실패 테스트 선행 작성: yes — 트리거 개수 단언(4→6)이 먼저 붉게 떴고,
             TestTheDriverCountsConsecutiveCycleFailures로 카운터를 고정
- 안전 불변식 §0 위반 여부 검토: 통과.
  RunOnce는 반환값·부수효과 무변경(defer로 카운터만 갱신),
  트리거 2종은 전부 ENTRY_BLOCKED이며 HALT_ALL 비도달 단언이 그것을 지킨다
```

## 구현 판단 기록 (safe local)

1. **`ReconcileDriver.Health()` 신설** — 지속 열화 임계는 "살아 있는 루프의 연속
   실패 횟수"를 읽어야 하는데, `Run`의 반환값으로는 표현할 수 없다.
   `filldetect.Detector`는 같은 사실을 이미 `Health().Outage.Consecutive`로
   내놓고 있으므로, 대사 쪽에도 대칭으로 카운터를 두고 감독은 한 인터페이스
   (`LoopHealth`)로 두 루프에 같은 질문을 한다. 루프의 재시도 결정은 무변경.

2. **`engine.Runtime`이 루프를 주입받는 이유** — `internal/filldetect`는
   `internal/push`를 거쳐 WTS 클라이언트를 transitive import 한다.
   `internal/app/engine/deps_test.go`가 그 import를 금지하므로 Detector를
   engine 패키지 안에서 조립할 수 없다. 감독 계약은 engine에 두고, 조립은
   cmd/tossctl가 한다. SLO 어댑터(Detector→SLOPressure)도 같은 이유로 cmd 쪽에
   있으며, 이는 exitloop.go가 "the adapter is in the engine's wiring, so this
   package does not depend on the detector"라고 적어 둔 의도 그대로다.

3. **활성 마커를 JSON으로 쓴 이유** — task 2.1이 요구한 "바이너리 stale 경고"는
   실행 중 엔진의 빌드를 알아야 답할 수 있다. `internal/runlock`의 본문은 사람이
   읽는 산문이고 "nothing parses them"이 그 계약이므로, 엔진 마커는
   `internal/enginelock`에 별도 포맷(pid·시작시각·binstamp)으로 두고 **신선도
   판정만 `runlock.Fresh`를 그대로 쓴다**. 수치(1분·5분)는 runlock의 것을
   재사용하며 drift 테스트가 대조한다. 파싱 실패는 "실행 중, 빌드 미상"으로
   degrade 한다 — 대시보드가 포맷 변경으로 안 그려지는 일은 없어야 한다.

4. **verify runlock 검사를 인터록 뒤에 둔 것** — 스펙의 기동 순서(①flock
   ②게이트 OFF ③인터록 ④verify runlock)를 축자 그대로 따랐다. 결과적으로 검증
   중에 기동을 시도하면 계좌 조회 1회를 쓰게 되지만, 안전 사유(인터록)가 예의
   사유(rate 예산)보다 먼저 답해야 한다는 순서는 스펙의 결정이다.

5. **`OfficialReads`에 `HoldingsRaw` 추가** — 대사 스냅샷은 raw 경로를 선호하고
   (`reconcile.RawPositionsReader`), 그 경로가 편입 기록의 `cost_basis`를
   브로커 원문 decimal로 보존한다(position-ledger SHALL). 인터페이스에 읽기
   메서드를 더하는 것은 reads.go가 명시적으로 허용한 확장 방향이며, seal 테스트
   (mutator 미선언)는 그대로 통과한다.

6. **비-Unix 정직한 거부** — soakproc_unix/other 선례를 따르되 결론은 반대다.
   soak의 비-Unix 분기는 *편의*(detached 실행)를 포기하고 계속 동작하지만, flock의
   비-Unix 분기는 *보증*(journal 단일 writer)을 포기하는 것이므로 거부한다.
   `enginelock.ErrLockUnsupported`가 그 이유를 말한다.

7. **`waitForEngineExit`의 타임아웃을 인자로** — 상수를 인라인으로 읽으면 "죽지
   않는 엔진은 죽이지 않고 보고한다"는 성질을 1분짜리 테스트로만 검증할 수 있다.
   프로덕션 호출부는 `engineStopTimeout`을 그대로 넘긴다.

## 발견했으나 이 change에서 고치지 않은 것

- **verify 측의 엔진 마커 검사**: 엔진 실행 중에 `verify run`을 거부하는 반대
  방향은 execution-verification change(2b)의 후속 태스크다(라운드 2 P2 처분).
  이 change는 엔진 측 검사(verify runlock 신선 시 기동 거부)만 소유한다.
- **`obs.Publisher` 미배선**: `engine run`은 Publisher를 넘기지 않으므로 critical
  알림은 outbox에 PENDING으로 남고 entry gate가 latch 된다. 이는
  exitwiring.go의 landed 결정(risk-management: critical 알림 outbox 전달 실패
  지속 → ENTRY_BLOCKED)이 명시한 방향이며, 전송 설정은 audit 대상 운영 설정이라
  별도 change의 소관이다.
- **Guardian 미배선**: 프로덕션 issuer가 아직 없다. 게이트 ON 엔진은 인터록 조항
  1에서 거부되며, 조항 6이 먼저 거부하므로 실질 비용은 0이다.
- **`reconcile.Ingestor.DefaultMarket`**: 엔진 배선에서 여전히 빈 문자열이다.
  드라이버의 `DefaultMarket`("kr")은 편입 판정 쪽만 덮는다. 브로커 payload가
  `marketCountry`를 항상 실어 보내므로 도달 불가에 가깝지만, 두 기본값이 한
  기제의 두 반쪽인 것은 사실이다 — 편입을 실제로 켜는 change가 정리할 자리.

## Manager 판정 (구현 완료, 2026-07-27)

독립 검증: `make gate` 재실행 GATE PASS(병행 세션의 7단계 gate.sh 기준), 전체 스위트 -race **2873**/52pkg 0 FAIL 재실행, flock 시스템콜·ModeTrigger 상수·조립 전 잠금 테스트·콘솔 게이트 개념 금지 테스트 직접 확인. 편차 6건 전건 승인:

- **ReconcileDriver.Health() 추가**: 승인 — 5주기 임계는 Run 반환으로 읽을 수 없고, filldetect의 기존 카운터와 대칭. 루프 행동 무변경.
- **루프 주입 구조**(deps 가드로 엔진 패키지가 filldetect를 직접 import 불가 → cmd 조립): 승인 — 기존 의존성 규율을 지키는 올바른 형태.
- **마커 JSON**(runlock 산문 계약과 분리, Fresh·상수만 재사용+drift): 승인 — stale 바이너리 경고에 binstamp가 필요하다는 근거 타당.
- **Hints.Validate 추출**(조립·런타임 검증 단일 정의), HoldingsRaw 확장, 타임아웃 파라미터화: 전건 승인.
- **비-Unix flock 정직 거부**(soak의 편의 저하와 달리 단일 writer가 걸린 사안): 승인 — 방향 구분이 정확하다.
- found-not-fixed: verifylive/plan.go gofmt 오염은 base 시점 기존 상태(병행 세션 활동 구간 — 소유 불명이므로 무접촉 유지), Ingestor.DefaultMarket 이원화는 adoption 활성 change 소관 기록, obs transport·Guardian 미배선은 landed 방향 그대로 — 승인.
- **병행 세션 무접촉 확인**: Makefile·gate.sh·AGENTS.md·CLAUDE.md·.gitignore·신규 change 디렉터리(harden-net-rr-gate·adopt-stockos-full-sdd)·docs/WORKFLOW.md 전부 스테이징 이력 없음을 커밋 범위로 확인.

**GATE PASS 확정 · archive 진행.** 실전 기동은 여전히 불가(ProtectionReady 상수 미충족 — 2c까지), autostart는 저장소 준비만·활성화는 게이트 ON §0.7 항목.
