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

### 안 한 것
- (1단계 당시) 리뷰 2회차 — 결정 반영 뒤 1.6b 로 실시. `codegraph affected`, CodeGraphContext·GBrain, 전체 시험·게이트, `make sdd-sync`. 운영 파일(`~/.config/tossctl`)은 안 읽음 — 실제 운영에 US 기록이 있는지 모름(Q1 의 현실 영향 미확인).
