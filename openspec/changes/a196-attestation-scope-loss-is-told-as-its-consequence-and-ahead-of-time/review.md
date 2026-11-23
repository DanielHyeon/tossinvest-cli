# a196-attestation-scope-loss-is-told-as-its-consequence-and-ahead-of-time Review

## 계약 작성 (2026-11-23)
- 작성: Manager(Fable). 근거 조사: 읽기 전용 조사원(Opus) — 분기 주장의 AST 는 세션 스크래치 산출물 인용이며 change 안 사본·HEAD 재확인은 tasks 1.3.
- `openspec validate a196-attestation-scope-loss-is-told-as-its-consequence-and-ahead-of-time --strict`: valid (rc 0, 2026-11-23 HEAD 1ea261100).
- proposal-freeze 리뷰: 아래 「1단계」.
- Function Logic Map: 편집 대상 기존 함수의 번들은 tasks 1.5 에서 만든다.

## 1단계 — 계약과 증거 (팀메이트 Opus, 2026-11-23)

worktree `~/tossos-work/a196-impl`, 브랜치 `fix/a196-attestation-deadline-notice`, base `064f59fd99c11ea3a19adcdb804bbcdfed2470cb`. production·시험 Go 편집 0.

### 1.1 base·PM
- `capture_change_base.py` → `base-commit.txt` = `064f59fd9…`. `generate_master_tracker.py --check` rc 0(「hierarchy and generated trackers are current」). Story `STORY-TOS-a196` 은 registry·Feature 에 있음.

### 1.2 CodeGraph(1.6.0, worktree 색인)
- `analysis/code-context/` 세 파일 + `codegraph-callers.txt`(12 심볼). 불일치 1: CodeGraph 가 `attestationWatch.prepare` 의 호출자(`checkOnce`@232)를 못 찾음 → grep 으로 해소, 결론 불변.
- CodeGraphContext·GBrain: not-applicable(사유는 `codegraphcontext-context.md` — 그 채널은 안 봤음).

### 1.3 AST 대조
- `analysis/cited-ast/` 22개(rc 0 전부), 대조표는 `analysis/cited-ast/README.md`. proposal 표 여섯 줄 중 사실 오류 1(줄 번호: `daysLeftCeil` 은 call@112, 종전 :118) → proposal 정정(「AST:」 표시).
- 측정: `analysis/probe/`(overlay 프로브 — 저장소 무편집, `run.sh` 로 재현, `output.txt`).
- 결정을 바꿀 수 있는 발견은 문서에 반영하지 않고 아래 Q 로 올림.

### 1.4 편집 표면
- 고칠 기존 production 함수(전부 `internal/app/engine/attestation_watch.go`): `judgeAttestation`(기한 기준) · `attestationWatch.prepare`(신원·「보냄」 재설정, L286-290) · `stageKey`(누락 집합) · `attestationWatch.eventFor`(문구·outbox Key L370) — **4개**. 조건부 1: `daysLeftCeil`(0 이하 잔여를 이 함수로 처리하면 — Q4). `checkOnce` 는 무편집 가능(「보냄」 기록 L241·L257 의미 불변이면).
- 뒤집히는 a135 시험 함수: **2개** — `TestA135StageBoundaries`(「D-8 은 아직 알리지 않음」·「2일 1초 = D-3」 두 행; 픽스처 증거가 발급 1시간 전이라 기한 = 만료 − 1h, 프로브 실측), `TestA135SameStageOnANewAttestationIsAnnouncedAgain`(전반 — 같은 날 만료 A→B 2건 기대, D5 로 1건). 같은 함수 후반(발급만 +1분)은 D1 식이 유효 기간을 1분 줄여 기한이 바뀌므로 **우연히 통과**(통과 이유가 바뀜). 대안: 헬퍼 `a135Attestation` 의 증거 시각을 발급 시각으로 바꾸면 StageBoundaries 무편집(대신 헬퍼 번들) — Q5.
- D2 전제 시험 패키지: `go list -deps`·`-test -deps` 양방향으로 `internal/app/engine` ↔ `internal/soak` 의존 없음 → **어느 쪽 시험 import 도 순환 없음**. 권고(제안만): `internal/soak` 의 외부 시험(`soak_test`)에 두기 — 이미 `threeCleanDays()`·`criteria()`·`supervisedProofs(at)` 픽스처가 있고(supervised_test.go:25, attest_test.go:20) `TestSupervisedEvidenceOlderThanTheValidityIsNotEvidence` 와 같은 모양이다. 단 그 시험은 engine 의 비공개 기한 함수를 부를 수 없어 기한 식을 시험 안에서 다시 쓴다 → 「같은 값」 단언은 발급 쪽(정각 제외·V−1ns 수용)만 고정하고, 계산 쪽은 engine 시험이 같은 경계값으로 고정하는 둘로 나뉜다. engine 내부 시험에서 `soak.BuildAttestation` 을 직접 부르면 한 시험이 양쪽을 묶지만 soak 기록 픽스처를 engine 시험에 다시 만들어야 한다.
- 구조 고정: a166 census(`internal/accountbudgetruntime/testdata/a166-wiring/shapes.json` 키 90·`sites.json`)에 `attestation_watch.go` 0건, 편집 함수 0건. 그 밖 시험·도구가 이 파일·함수명을 고정하는 곳 grep 0(archive a135 의 변이 스크립트 제외).

### 1.5 FLM 번들(pre-edit)
- `analysis/function-logic/` 6개: production 4(`judgeattestation`·`attestationwatch.prepare`·`stagekey`·`attestationwatch.eventfor`) + 시험 2. ast.json·FLM·BTM·risk(일치 0; 시험 파일의 go-panic 2건은 대상 함수 밖).
- BTM 의 덮는 시험은 실측 커버(`analysis/implementation/base-cover-attestation_watch.out`, `-run 'TestA135'` rc 0). 미커버 2: `judgeAttestation` B8(도달 불가), `prepare` B2(하네스 log nil) — 둘 다 a196 범위 밖.
- `python3 tools/logic-map/check_analysis.py --change a196-…`: rc 0, 「required 0 … evidence complete or diff-proven exempt」(Go 미편집이라 required 0 — 번들 형식 검사만 통과).

### 1.6 validate·정본 대조·proposal-freeze 리뷰
- `openspec validate … --strict`: valid(rc 0, 작업 트리).
- 정본 대조(`diff`, Requirement 블록만): 달라진 곳 = ① 날수 문단(「만료까지」→「재시작이 거절되기 시작할 시각(기한)까지」, 「D-2 와 D-1 에는」→「D-2 이하에는」) + 새 문단 2(기한 정의·날수 알림 내용), ② 거절 문단 끝에 제목·본문 문장 추가, ③ 중복 방지 문단(「같은 attestation(발급 · 만료 시각이 같은 파일)의」→「같은 기한의」, 「파일이 바뀌면 단계를 새로 센다」 삭제 → 「기한이 바뀌면 날수 단계를 새로 센다」 + 재발급 침묵 문장) + 끝에 Scenario 6개. **그 밖 글자 변화 0**(기존 Scenario 11개 그대로). 「D-2 이하」 는 ①에 함께 든 의미 변경(D3)이라 따로 적음.
- 리뷰 1회차: codex-cli 0.165.0 read-only(원문 `analysis/review/proposal-freeze-r1-codex.md`). **P0 0 · P1 2 · P2 3.** 2회차: 미실시 — 1회차 발견이 전부 사실 확인 또는 Manager 결정 대기라, 처분 뒤(freeze 전) 다시 돌리는 것이 의미 있다고 판단.

| # | 등급 | 발견(요약) | 처분 |
|---|---|---|---|
| R1 | P1 | 파일에는 endpoint 당 **첫 유효 증거 하나**만 실리고(acceptSupervised `seen`), CLI 는 KR→US 기록 순으로 넘긴다 → 다른 기록의 더 새 증거가 파일에서 사라져 D1 기한이 실제 POST 상실보다 이르고, 갱신으로 기한이 뒤로 뛴다(D4 「갱신으로 바뀌지 않는다」 거짓). 정각 경계(`age >= validity` 제외) — spec 「그 시각 뒤」는 「그 시각 포함」이어야 하고 D2 의 직전·직후 시험은 `>=`→`>` 변이를 못 잡는다 | 사실 확인(팀메이트 독립 실측과 일치 — 프로브). 결정·spec 문장 변경이라 **Manager 결정 대기 Q1**. 경계 문구도 Q1 에 묶음 |
| R2 | P1 | D3 「그 구간은 최대 6시간」 거짓 — 갱신 실패 시 파일 미기록, Verify 는 증거 나이를 안 봄 → 만료까지 통과. 기존 `daysLeftCeil` 을 0 이하에 쓰면 잔여 +23.5h~−24h 동안 같은 `imminent:1` 키라 **약 48시간 침묵** | 사실 부분은 **반영**(design D3 근거 칸 정정, 「AST:」 표시). 날수·키 처리는 **Manager 결정 대기 Q4** |
| R3 | P2 | 그대로 둔 기존 Scenario 「만료 8일 전」·「만료 7일 전부터 3일 전까지」·「정상 갱신」이 새 본문·새 Scenario 와 동시에 만족될 수 없다(발급자 파일은 증거가 있으면 늘 감독 증거가 기한 — evidence-reconciliation) | **Manager 결정 대기 Q2**(MODIFIED 범위 확장) |
| R4 | P2 | 재시작 뒤 같은 key 는 메모리 `sent` 소실 + outbox `RemindAfter`(1h) 재활성 → 「같은 단계 한 번」 이 재시작마다 깨진다. a135 D2 는 이를 명시 허용했는데 a196 delta 는 그 예외를 말하지 않는다 | **Manager 결정 대기 Q6**(정본 문구에 a135 의 재시작 예외를 적을지) |
| R5 | P2 | §5-11 의 기존 시험 영향 예상이 틀림(StageBoundaries 2행 뒤집힘, SameStage 후반·Renewal 시험은 통과 이유가 바뀜), 변이 6개가 빠뜨리는 것: 수용 경계 `>=`→`>`·복수 기록 대체 증거·음수 잔여 키 재사용·KST→UTC 날짜·단계 전이 key·재시작+outbox | 시험 영향 사실은 1.4 에 기록(팀메이트 실측과 일치). 시험 전략·변이 범위 변경은 **Manager 결정 대기 Q5** |
| — | 정보 | D5 자체가 영구 누락을 만든다는 주장은 확인 안 됨(날짜·단계·누락 집합이 key 에 있으면 별도 key) | 기록만 |

### Manager 결정 대기
- **Q1(R1)**: 두 시장(KR·US) 검증 기록이 함께 있을 때 D1 기한이 이르게(거짓 경보 쪽) 나온다. 받아들이고 spec·D4 문장을 「한 시장 기록일 때」로 좁힐지, 발급자가 endpoint 별 최신 증거를 싣게 할지(범위 확장 — 발급자 편집). 경계 문구(「그 시각 이후(포함)의 첫 갱신」)와 D2 시험에 정각 행 추가 여부도.
- **Q2(R3)**: 기존 Scenario 3개를 「감독 증거가 실리지 않은(또는 만료가 기한인) attestation」으로 한정하거나 감독 증거 판으로 고칠지.
- **Q3**: ntfy 제목은 `Title` 헤더에 한글 UTF-8 원바이트로 나간다(RFC 2047 없음; Go 클라이언트는 허용). ntfy.go:112 주석은 「서버가 ASCII-fold」라고 적었지만 저장소에서 검증 불가. 11-18~23 의 21 건이 기기에서 한글 제목으로 보였는지 Manager 실측이 필요 — 깨져 보였다면 D7 「제목에서 결과를 말한다」를 본문 첫 줄로 옮겨야 한다.
- **Q4(R2)**: 감독 증거의 기한이 지났는데 파일이 통과하는 구간(갱신 실패 시 최대 파일 만료까지)의 날수·키 — 예: 0 이하를 별도 단계·KST 날짜 키로 하루 한 번.
- **Q5(R5)**: a135 시험 수리 경로(StageBoundaries 표 수정 vs 헬퍼 `a135Attestation` 증거 시각 = 발급 시각), SameStage 시험의 재작성·삭제, 변이 목록 확장 여부.
- **Q6(R4)**: 정본에 「재시작 직후 현재 단계를 한 번 더 알릴 수 있다」(a135 D2 의 허용)를 적을지.

## 1단계 — Manager 결정 반영 (2026-11-23)

Manager 결정(Q1~Q6)을 문서에만 반영. production·시험 Go 편집 0.

| Q | 결정 | 반영 자리 |
|---|---|---|
| Q1 | 이른 경보 수용, 발급자 무편집. 기한 = 파일에 **실린** 증거 기준, 정각 = 이미 지남 | spec E2 문장 교체, design D1·D2(V−1ns·V·V+1ns)·D4 근거·§6 잔여, proposal What Changes 1 |
| Q2 | 「만료 N일」 Scenario 4개를 「기한」 기준으로 | spec S1~S4, 중복 초안 「감독 증거의 기한 2일 전」 은 S3 에 합침. **이름은 유지** — `openspec validate --strict` 가 MODIFIED 블록의 정본 Scenario 이름 누락을 거절(「omits scenario(s) the current spec still has」)해서 이름은 그대로 두고 WHEN/THEN 만 고침. 「감독 증거의 기한 7일 전」 은 S2 와 합치지 않음(이른 쪽이 감독 증거일 때의 할 일을 고정 — intent 파일에 사유) |
| Q3 | 결과 문장을 제목과 본문 첫 문장 둘 다 | spec E5·S10, design D7·§5-7·§6(제목 표시 미실측), proposal |
| Q4 | 남은 시간 ≤ 0 → 새 단계 `proof_overdue`, 단계+기한 시각+KST 날짜, 하루 한 번 critical, `daysLeftCeil` 양수에만 | spec E1(「D-2 와 D-1」 로 되돌림)·E4·S8, design D3·D5·D7·§2·§5-3, proposal. AST: `attestationVerdict.critical` 은 default(B3 case@78 → return@79 true)가 새 단계를 흡수 → **무편집·번들 없음** |
| Q5 | 시험 함수 2개만 고침(헬퍼 무편집), 「통과 이유가 바뀐 시험」 목록, 변이 6 → 12, D2 자리 = soak 외부 시험 + engine 계산 시험 | design §5-10·11·12·변이, tasks 3.1·3.3·4.1·머리말, FLM 번들 메모 |
| Q6 | 재시작 뒤 1회 재알림을 정본에 | spec E6 끝 문장, design D5·§6 |

- 정본 대조: `analysis/spec-edit-intent.md` — 의도한 편집 16(본문 6 · Scenario 고침 4 · 추가 6), diff 원문 첨부, 그 밖 글자 변화 0. Scenario 11 → 17.
- 최종 편집 대상: production 4(`judgeAttestation`·`attestationWatch.prepare`·`stageKey`·`attestationWatch.eventFor`) + 시험 2(`TestA135StageBoundaries`·`TestA135SameStageOnANewAttestationIsAnnouncedAgain`) = 번들 6(변동 없음). 새 단계 상수는 const 블록.
- `openspec validate --strict` valid, `check_analysis` rc 0, PM `--check` rc 0.

### 1.6b proposal-freeze r2 (마지막 회차)
- codex-cli 0.165.0 read-only, 원문 `analysis/review/proposal-freeze-r2-codex.md`. **P0 0 · P1 0 · P2 3.** r1 처분(Q1~Q6)에 대한 반박 0 — Q4 는 r1 의 48시간 침묵을 해소, 알림 꺼진 엔진에서 `proof_overdue` 도 일반 등급으로 강등됨(eventFor B1), 전이·재시도에서 새 침묵 반례 없음, `spec-edit-intent.md` 의 diff·16 목록이 실제와 완전 일치 확인.

| # | 등급 | 발견 | 처분 |
|---|---|---|---|
| R2-1 | P2 | Scenario 「감독 증거의 기한을 세는 동안 파일이 다시 발급된다」 의 「그날 알림은 더 나가지 않는다」 와 §5-2 「같은 날」 이, 같은 날 안에서 날수가 하나 주는 정상 전이(D-6→D-5)까지 금지 — 본문의 억제 단위는 「같은 기한의 같은 날수 단계」 | **반영**(결정 무변경 문장 정합): Scenario WHEN 「남은 날수가 여전히 6일인 동안」·THEN 「그 재발급으로 알림이 더 나가지는 않는다」, design §5-2 「같은 날수 단계 안에서」. intent 파일 S6 갱신, validate valid |
| R2-2 | P2 | D7 의 날수 통지 「그 뒤 첫 갱신부터 재시작이 거부된다」·`proof_overdue` 「다음 갱신부터 거부된다」 가 Q1 의 조건(「다른 유효한 감독 증거가 없으면」)을 빠뜨리고 단정 | **Manager 결정 대기** — Q4 본문은 Manager 지정 문구이고 D7 은 「뜻은 고정」이라 바꾸지 않음 |
| R2-3 | P2 | 변이 12 중 시험 계획이 못 죽일 수 있는 것: ④·⑤(outbox Key 만 변형 — 가짜 notifier 는 outbox 중복 억제가 없어 건수로 못 잼, Key 비교 단언 필요), ⑧(binding 을 우회하는 새 기한 계산 — endpoint 별 시각이 다른 입력 미명시), ⑩(KST 날짜만 바뀌는 시각 쌍 없음 — 기존 시험은 25시간 전진), ⑪(scope_lost→expired 전이 시험 없음, 누락 집합이 남으면 단계 제거로도 key 가 갈려 변이 자체가 충돌을 보장 못 함), ⑫(`MissingEndpoints` 가 이미 정렬 — 동등 변이일 수 있음) | **Manager 결정 대기** — 시험 전략·변이 정의 변경 |

### Manager 1단계 검증·freeze 2026-11-23
- Manager 가 커밋 `0d93cfd2b` 의 MODIFIED 본문·Scenario 17개를 직접 읽고 strict validate 재실행 — freeze. 남은 결정: R2-2 조건 삽입(D7), R2-3 시험 보강(i)~(iv)·변이 ⑪ 재정의·⑫ 동등 처리, Q2 이름 유지 승인 — design 반영.

## 2단계 — Pre-Edit 선언 (확정 2026-11-23, tasks 2.1)

```text
Pre-Edit Gate:
- change id / task id: a196-attestation-scope-loss-is-told-as-its-consequence-and-ahead-of-time / 3.1~3.3
- 대상 심볼(패키지.함수): engine.judgeAttestation · engine.(*attestationWatch).prepare · engine.stageKey · engine.(*attestationWatch).eventFor(production 4),
  시험 engine.TestA135StageBoundaries · engine.TestA135SameStageOnANewAttestationIsAnnouncedAgain(2). 같은 파일의 새 함수(restartDeadline 등)·새 단계 상수·
  verdict 필드 추가, 새 시험 파일 internal/app/engine/a196_attestation_deadline_internal_test.go · internal/soak/a196_supervised_deadline_test.go.
  무편집: attestationVerdict.critical · daysLeftCeil · checkOnce · binding · 헬퍼 a135Attestation · internal/obs · internal/attest · internal/soak production ·
  cmd/tossctl. 늘어야 하면 멈추고 보고.
- CodeGraph definition/callers/callees/impact: analysis/code-context/codegraph-baseline.md(1.6.0) — production 호출 자리 각 1(prepare←checkOnce@232,
  judgeAttestation←prepare@283, stageKey←prepare@294, eventFor←prepare@298), a166 census 적중 0
- CodeGraphContext 후보와 evidence reconciliation: not-applicable(사유 기록), evidence-reconciliation.md(불일치 1 — CodeGraph 의 prepare 호출자 누락, grep 으로 해소)
- 기존 동작 파악 근거: analysis/cited-ast(22) · function-logic 6 번들 · 커버(base-cover-attestation_watch.out) · a135 시험 전수(1.4)
- Function Logic Map / Branch Test Map: analysis/function-logic/internal-app-engine--{judgeattestation,attestationwatch.prepare,stagekey,attestationwatch.eventfor,
  testa135stageboundaries,testa135samestageonanewattestationisannouncedagain}
- upstream 상속 테스트 영향: no(엔진 감시는 TossOS 계층, a135)
- 실패 테스트 선행 작성: yes — RED 원문 analysis/tdd/red.txt
- 설정·DB·journal 변경과 rollback: 없음(attestation 파일 형식·원장 스키마·obs 타입 열거 무변경). outbox Key 문자열 모양만 바뀜 — 옛 이미지로 롤백하면
  옛 Key 로 같은 상태를 한 번 더 알릴 수 있음(무해). 롤백 = 이전 이미지.
- 안전 불변식 §0 위반 여부 검토: 통과 — 관측 전용(브로커 호출 0, 진입·청산·손절·운영 모드·게이트 무접촉), 주문 0, 토글 없음, 손절·사이징 무변경.
  critical 알림 수 증가(proof_overdue 하루 1)는 기존 알림기 경로 — 알림 꺼진 엔진은 일반 등급으로 강등(eventFor B1 그대로).
```

## 2단계 — 구현·검증 (팀메이트 Opus, 2026-11-23)

- 3.1 RED: `analysis/tdd/red.txt` — 껍데기 없이(새 시험은 기존 API + 문자열 단계 코드만 씀) 구현 전 컴파일·단언 실패. engine 13 시험 중 11 FAIL,
  2 PASS(KST 날짜 키·단계별 키 — base 가 이미 지키는 성질의 회귀 핀). soak 2 시험은 발급자 전제 고정이라 base 에서 PASS 가 정답, 실패는 변이 M07 로 확인.
- 3.2 GREEN: `attestation_watch.go` — 새 함수 `restartDeadline`, 새 상수 `stageProofOverdue`·`deadlineExpiry`·`deadlineProof`, verdict 필드 2,
  기존 함수 편집 4(`judgeAttestation`·`prepare`·`stageKey`·`eventFor`). `critical`·`daysLeftCeil`·`checkOnce`·`binding` 무편집. internal/obs·attest·soak production 무편집.
- 3.3: GREEN 직후 a135 실패는 예상대로 두 함수뿐(`analysis/tdd/green-before-a135-fix.txt` — StageBoundaries 3행·SameStage). 두 함수만 고침, 헬퍼 무편집.
  번들 6 편집 후 판 + `renumber.txt`(difflib), 편집 전 AST 는 `analysis/pre-edit-ast/`.
- 4.1 변이 원장 `analysis/mutation/`: 13 판(⑧ 을 a·b 둘로) — CAUGHT 12, SURVIVED 1(M12 — 동등 변이, 사유 README). 사본 하네스·대조군 GREEN·닿음 전수 1·sha256 원복.
- 4.2: 선택 시험 `analysis/verify/selected-tests.txt`(engine `-run 'A196|A135|Attestation'` ok, soak `-run 'A196|Supervised|BuildAttestation'` ok, 최상위 PASS 42),
  `make lint` rc 0(`analysis/verify/make-lint.txt` — gofmt + go vet 무태그·`tossos_testseams`), `openspec validate --strict` valid, `check_analysis` rc 0
  (required 6 = 번들 6, 「evidence complete」) — `analysis/verify/validate-and-analysis.txt`. 실제 알림 문장 `analysis/verify/alert-texts.txt`(overlay, 저장소 무편집).
- **미실행 범위**: 전체 스위트(`make test`)·`make test-seams`·`make test-race`·`make gate`·`make sdd-sync`·`codegraph affected` — 지시대로 안 돌림. engine 패키지의
  다른 시험(보조 실행자 배선 등 `-run` 밖), `cmd/tossctl` 시험, internal/obs·journal 시험은 안 돌림(무편집 패키지). 실제 outbox(journal)와의 결합은 Key 문자열 단언으로
  대신했고 journal 위에서 돌려 보지 않음. ntfy 기기의 한글 제목 표시는 미실측(Q3). 4.3(구현 적대 리뷰·gstack review)은 Manager 지시로 보류.

## 구현 적대 리뷰 r1 (2026-11-23, Opus 별도 컨텍스트 + codex 0.165.0)

- 판정(Manager 전달): P0 0 · P1 2 · P2 8. 알림이 영구히 삼켜지는 경로 0, 감시·인터록 불일치 0. 리뷰 원문은 Manager 쪽 컨텍스트에 있고 이 worktree 에는 처분만 받았다.

| # | 발견 | 처분 |
|---|---|---|
| P1-1 | 감독 증거가 기한인 imminent 본문을 고정하는 시험 없음 — `&& v.Stage == stageNotice` 변이 생존(D-2·D-1 이 만료·`soak attest` 를 말함) | **반영** — 시험 `TestA196ImminentProofNoticeTellsTheProofDeadline`, 변이 N1 CAUGHT. production 무편집 |
| P1-2 | 범위 상실 문장이 누락 종류와 무관하게 「주문 · 취소」·원인·`verify run` 단정(GET 만 빠지면 거짓) | **반영** — eventFor 를 누락 집합으로 가름(전부 POST: 「주문 실행 검증이 attestation 에 없다」·원인·`verify run` / GET 섞임: 「엔진이 쓰는 endpoint 가 attestation 에 없다」·원인 없음·재발급, POST 가 섞이면 `verify run` 덧붙임). 판정은 새 함수 `mutationEndpoint`(메서드 접두). 시험 `TestA196ScopeLossWordingFollowsTheMissingKind`(4입력), 변이 N6 CAUGHT. 번들 재생성 |
| P2-1 | prepare 의 「보냄」 재설정 미고정 | **반영** — 시험 `TestA196RejectionAfterAHealthyFileIsAnnouncedAgain`, N2 CAUGHT. 잔여: 같은 날 A→B→A(누락 집합 왕복)의 마지막 A 는 억제됨(하루 한 번 범위 안, 수리 안 함) |
| P2-2 | `restartDeadline` 의 `ExpiresAt ≤ IssuedAt` 조항 미고정 | **반영** — 시험 `TestA196NoProofDeadlineWithoutAValidityWindow`, N3 CAUGHT. 「만료 무시」 R1 = N4 생존 — 동등(사유 mutation/README) |
| P2-3 | overdue key 에서 기한 삭제 변이 생존 | 기록만 — N7 실행해 생존 확인, 동등 사유 기록 |
| P2-4 | soak 전제 시험이 기본 유효 기간만 씀 | **반영** — 10일 유효 기간 행 `TestA196ANonDefaultValidityMovesBothValues`, N5 CAUGHT |
| P2-5 | 날수 통지가 지금 도는 엔진의 손절까지 멈추는 것처럼 읽힘, 「N일 뒤」 와 시각이 최대 하루 어긋남 | **반영** — 「…그 뒤 첫 갱신부터 엔진 재시작이 거부된다. 재시작이 거부되면 그때부터 자동 손절이 돌지 않는다. 지금 실행 중인 엔진은 계속 돈다.」, 날수는 「D-N(<기한> KST)」, proof_overdue 도 같은 기준 |
| P2-6 | 만료가 기한인 통지의 할 일 삭제 변이 | 기록만 — 정본 요구 아님 |
| P2-7 | notice → unreadable → 같은 기한 notice 에서 일반 통지 1회 추가 | 기록만 — base 동작, 회귀 아님 |
| P2-8 | 만료 기준 날수 경계가 더는 고정되지 않음 | 잔여 — 발급자 파일에서는 만료가 기한이 되지 않고 계산 경로(`daysLeftCeil`)가 같음 |

- 상태 전이(리뷰가 실측한 요지, Manager 전달): 날수 ↔ 지남 ↔ 거절, 거절 단계 사이, 기록 실패 재시도, 재시작 — 하루 넘게 삼켜지는 경로 없음.
- 재실행: 선택 시험 engine·soak ok(최상위 PASS 47), `make lint` rc 0, strict validate valid, `check_analysis` rc 0(required 6) — `analysis/verify/`. 변이 run-2: CAUGHT 17, 동등 생존 3(M12·N4·N7). 알림 문장 `analysis/verify/alert-texts.txt` 갱신.

## gstack review r1 (2026-11-23, tasks 4.3)

- 원문 `analysis/review/gstack-review-r1.md`(처분 표), Codex 외부 적대 원문 `analysis/review/gstack-review-r1-codex-adversarial.md`. 1회차로 종료.
- 반영: G1(주석)·G6(주석 + design §6 잔여)·G4(만료 기한 통지 문구 시험, 변이 N8 CAUGHT). production 동작 무변경(주석만).
- **Manager 결정 대기(우선 2)**: G12 — 회복 뒤 같은 날 같은 거절 재발이 ack/1시간 창 안이면 outbox 가 삼키고 그날 침묵(D5 key 의 회귀, journal/outbox.go:382-410 확인). G10 — 할 일 `tossctl verify run` 은 엔진 잠금 때문에 엔진이 도는 동안 실행 불가(verify.go:440·engine.go:209 확인) → 문구가 엔진 정지(손절 없음)를 숨김.
- Manager 결정 대기(그 밖): G5(시장 이름), G7(IssuedAt 0 가드), G11(콘솔 표시). 기록만: G2·G3·G8·G9.
- 재실행: 선택 시험 ok(최상위 PASS 47), `make lint` rc 0, strict validate valid, `check_analysis` rc 0. 변이 run-3: CAUGHT 18, 동등 생존 3(M12·N4·N7).

## G12·G10 수리 (Manager 결정 2026-11-23, 마지막 수리 회차)

- **G12(에피소드)**: 감시가 거절 에피소드의 시작 시각을 메모리에 듦(`attestationWatch.episode`·`episodeAt`, verdict `EpisodeAt`). 거절 신원(단계 + 누락 집합 + 기한)이 직전 판정과 다르면 이번 판정 시각에 새 에피소드.
  거절 단계·`proof_overdue` 의 key = 단계 + (누락 집합 또는 기한) + KST 날짜 + 에피소드 시작. 편집은 `prepare`(분기 4 추가)·`stageKey`(날짜 문자열)·`eventFor` 안. `checkOnce` 무편집.
  예: `attestation-watch:scope_lost:POST /api/v1/orders:2026-10-14:2026-10-14T03:00:00Z`, `attestation-watch:proof_overdue:2026-10-14T02:00:00Z:2026-10-14:2026-10-14T03:00:00Z`.
  시험 `TestA196RecurrenceAfterRecoveryReachesTheRealOutbox`(실제 `obs.Notifier` + 실제 원장 임시 경로, 전송 2 → 같은 에피소드 재발급 0 → 다음 날 1), `TestA196RestartedWatchStartsANewRejectionEpisode`(이름 변경, 재시작 = 새 key). 앞서 잔여로 둔 「A→B→A 의 마지막 A 억제」 해소.
  spec delta 본문 한 문장 + Scenario 「회복 뒤 같은 날 재발」 — 의도 편집 17, Scenario 18(`analysis/spec-edit-intent.md`).
- **G10(할 일)**: 감독 증거가 관련된 세 알림(범위 상실 POST 갈래 · 감독 증거 날수 통지(「기한 전에」) · `proof_overdue`, 그리고 POST 가 섞인 GET 갈래의 덧붙임)의 할 일 =
  「감독하 실주문 검증을 다시 해야 한다. 이 검증은 엔진과 같은 잠금을 써서 엔진을 멈춰야 실행되고, 멈춘 동안에는 자동 손절이 돌지 않는다. 보유 종목의 시장이 닫힌 시간에 엔진을 멈추고,
  터미널에서 `tossctl verify run`(주문을 받는 시장을 `--market` 으로) → `tossctl soak attest` → 엔진 재시작 순서로 하라.」 `--resume`/`--redo` 미사용. 시험 `TestA196ReverifyRemedyTellsTheEngineMustStop`. design D7·§6 갱신.
- 기록만(후속 후보, design §6): G5(시장 이름) · G7(`IssuedAt` 0 가드) · G11(콘솔이 만료만 표시). G2·G3·G8·G9 기록만(앞 절).
- 변이 run-4: E1·E2·E3 CAUGHT. **N2 가 생존으로 바뀜 — 동등(에피소드가 key 를 스스로 가려 「보냄」 재설정은 메모리 크기만 정함)**, 사유 mutation/README.
- 재실행: 선택 시험 engine·soak ok(최상위 PASS 49), `make lint` rc 0, strict validate valid, `check_analysis` rc 0. 번들 4(production) 재생성 + prepare 분기 9.

### 안 한 것
- (1단계 당시) 리뷰 2회차 — 결정 반영 뒤 1.6b 로 실시. `codegraph affected`, CodeGraphContext·GBrain, 전체 시험·게이트, `make sdd-sync`. 운영 파일(`~/.config/tossctl`)은 안 읽음 — 실제 운영에 US 기록이 있는지 모름(Q1 의 현실 영향 미확인).

## Manager 검증 (2026-11-24, Fable — tasks 5.1)
- HEAD 4b333f01c. production diff(attestation_watch.go 한 파일)를 직접 읽음: restartDeadline(만료와 감독 증거 기한 중 이른 쪽 — 파일 값만으로), 새 단계 proof_overdue, 날수 신원 = 기한 시각, 거절 단계 key = 단계 + (누락 집합 또는 기한) + KST 날짜 + 에피소드 시작(발급 신원 없음), 범위 상실 문장의 POST/GET 갈래, 재검증 할 일 문장(엔진 정지 필요·그동안 손절 없음·verify → attest → 재기동).
- 독립 재실행: ./internal/app/engine -run 'A196|A135|Attestation' → PASS 35 · FAIL 0(ok 0.653s); ./internal/soak -run 'A196|Supervised|BuildAttestation' → ok; -race ./internal/app/engine -run A196 → ok 4.829s; go vet 두 패키지 rc 0; openspec validate --strict valid; check_analysis evidence complete.
- 구현 적대 리뷰 r1(P1 2·P2 8)과 gstack review r1(G12 회복 뒤 같은 날 재발의 침묵, G10 할 일이 엔진 정지를 숨김)의 처분 반영을 확인. 잔여: 제목의 기기 표시 미실측, 재검증 플래그 조합 미확인, 두 시장 기록이면 경보가 이른 쪽, 거절 상태에서 재시작마다 알림 1회.
- 판정: 통과. 종결 진행.

## 종결 (2026-11-24, 종결 에이전트 — 사용자 요청 「커밋·push·archive(완료 게이트 포함)」)
- 첫 시도(rebase): `origin/main` `6ec2664c5` 위로 rebase(충돌 0, 새 HEAD `28234eb6a`), rebase 뒤 선택 시험 통과(`analysis/close/rebase-selected-tests.txt`), 착지 `28234eb6a` 기록, 종결 선행 `7376314a7` 에서 게이트 1회차 → **⑤ FAIL**(`analysis/close/gate-1-7376314a7.txt`): base `064f59fd9` → 착지 창이 rebase 로 끼어든 a195 의 Go 커밋까지 담아 `cmd/tossctl/engine_reconcile.go` 의 `newEngineReconcileResolveCmdWithDeps`·`runEngineReconcileResolve`·`buildReconcileResolveRuntime`(a195 의 함수)을 요구 — required 9. ⑤ 에서 멈춰 ⑥~⑪ 은 돌지 않음. Go·시험 내용의 실패가 아니라 절차(창 구성) 문제.
- 처분(절차 수리, Go 편집·base 재고정 없음): 브랜치를 rebase 전 `4b333f01c` 로 되돌리고(rebase 판은 로컬 `backup/a196-rebased-7376314a7` 에 보존), 그 위에서 `--record-landing` → `landed-commit.txt` = `4b333f01c`(도구 계산), 단독 커밋 `cbdd4b38e`. 그 뒤 `origin/main` 을 merge(`30171e63f`, 충돌 0) — 병합 커밋은 조건 8 의 「비병합 커밋」에 들지 않고, 창은 base..`4b333f01c` 로 이 change 의 작업만 담는다. merge 뒤 `check_analysis --change` rc 0(required 6, 증거 완전). `base-commit.txt`(`064f59fd9`) 는 HEAD 의 조상(확인).
- merge 뒤 선택 시험 재실행(`analysis/close/merge-selected-tests.txt`, HEAD 30171e63f): ./internal/app/engine -run 'A196|A135|Attestation' -v → PASS 35 · FAIL 0(ok 0.307s); ./internal/soak -run 'A196|Supervised|BuildAttestation' ok; -race ./internal/app/engine -run A196 ok 4.452s; go vet 두 패키지 rc 0.
- tasks: 5.1 체크(위 Manager 검증), 5.2 **선체크(gate ② 선행 조건, 선례 a193·a195)** — 게이트·archive·Story·push 는 이 체크 뒤 수행. 옛 5.3(운영 반영·엔진 생존 감시 타이머 설치)은 체크박스에서 빼 「6. 종결 뒤(사람 승인)」 문단으로 옮김.
