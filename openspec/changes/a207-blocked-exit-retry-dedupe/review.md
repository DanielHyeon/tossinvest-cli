# a207 review

## 1단계 경량 리뷰 r1 (2026-11-27 — 상한 2회 중 1회)

- 리뷰어: Codex CLI 0.165.0(`codex exec --sandbox read-only`, 작성자 Claude 와 다른 계열). 지시: 읽기 전용, 저널 · `~/.config/tossctl` 접근 금지,
  시험 · 네트워크 금지, 새 가드 · 기능 · 범위 제안 금지. 범위: proposal · design · spec 델타 · claims · evidence-reconciliation · 측정 하네스 둘.
- 질문 5: (1) B2 · N1 이 어떤 주기의 제출 여부 · 시점을 바꿀 수 있는가, (2) C1~C12 의 정확성, (3) 리더 · 불변식 · a206 · 롤백, (4) 측정의
  「키 구간」이 D2 와 같은가 · 하네스 안전, (5) 수치 불일치.

| # | 등급 | 지적 | 처분 |
|---|---|---|---|
| 1 | P1 | 즉시성 동등성이 성공 경로에만 성립함. (가) 주기 끝 원장 집계(옛 D5 · D6)가 다음 주기 시작을 늦춤 — `Run` 은 `ObserveOnce`(지연 실행 `endExitOrderCycle` 포함) 뒤에 잠듦(exitloop.go:432-440). (나) 판정 트랜잭션에 더한 질의의 오류가 제출을 막는 새 실패 경로가 됨. (다) 지연 측정이 단독 사본이라 경합을 재지 않음 | 수리. (가) D5 를 원장 읽기에서 `release` 의 메모리 셈으로 바꿈 — 관측 루프의 원장 읽기 0, 주기 끝 몫은 막히지 않는 `Offer` 하나(design D5 · D6 · §5.4). (나) D1b 추가: 질의 오류면 전체 증거 행으로 쓰고 판정을 계속함. 시험 T8 · 변이 M11 추가, spec 에 시나리오로 못 박음. (다) F9 · §5.4 · R7 에 「단독 사본 값 · 경합 미측정」을 적고, T1 이 2단계에서 시험 안 계측으로 잼 |
| 2 | P2 | 25,000 은 D2 술어의 수가 아님 — 스크립트가 사유 일치와 현재 무장의 해제를 더 요구했고, D1 의 증거 · source 조건은 보지 않았음. 「resolved = 수용」도 입증 안 됨 | 수리. 측정에 D1+D2 정확 술어를 더해 재실행: **25,946 행 · JSON 168,121,152 B**, 그중 그 무장 자신이 비해제인 것 8(수용 6 · 결과 미상 2). 에피소드 뒤 첫 비해제 무장의 사유도 셈(수용 6 · 결과 미상 2). proposal · design F3 · F4 · §3.1 · §3.2 수치 교체 |
| 3 | P2 | F4 간격 범위가 132.8 s 를 빠뜨림(에피소드 평균임), F10 의 ≈21,800 근거 없음, F9 의 출력 · 27,072 행이 보관되지 않음 | 수리. F4 「에피소드 평균 간격 5.2~132.8 s」, F10 「상한 ≈25,100(2 s 마다 한 번)」, `a207_latency.out.txt` 보관 · 행 수 출력(id 없음) |
| 4 | P2 | C9: `Normal, false` 는 Normal 재생이 아니라 「출처 미입증 → 재생 거절」(replay.go:309-318) | 수리. C9 · design D-0 · proposal 비목표 문구 정정 |
| 5 | P2 | C2(attempt 없는 갈래 열거) · C7(JSON 은 둘~셋, legacy 는 0) · C8(같은 견적 ≠ 같은 decision — 입력 지문에 baseline · high water · rung · pending 포함) 문구가 근거보다 넓음 | 수리. 세 주장 정정(C2 에 해제 훅이 사유를 싣지 않는다는 apply_hook.go:832-834 를 더함) |
| 6 | P2 | 리더 · 롤백 문구가 orders 의 연결 행 수 판정(hydrate 전 ambiguous, account_views.go:545-578)을 무시함 | 수리. D4 · §6 · spec 의 리더 Requirement 를 「계좌 이력 · 판정 이력, orders 는 단일 연결일 때만」으로 좁히고 다중 연결 시나리오 추가 |
| 7 | P2 | a206 정비가 구간 첫 행을 비우면 「증거는 구간의 첫 행」이 깨짐 | 수리. §3.2 · R6 에 적고 사람 결정 6 으로 올림(a206 대상 술어 변경은 a206 의 몫 — 새 가드를 여기서 만들지 않음) |
| 8 | P2 | N1: 일반 등급에는 key 중복 제거가 없고 `NormalRelay.Offer` 는 가득 차거나 멈추면 버림. 해제 훅은 사유를 싣지 않아 attempt 없는 거절의 사유는 원장에서 못 셈 | 수리. D6 에 「한 번」의 근거(셈을 떼는 자리 하나)와 버림 경로를, R5 에 사라지는 경우 넷을 적음. 해소 알림 내용에서 「사유별 건수」를 뺌(메모리 셈은 해제 종류만 앎). §3.2 에 사유 기록이 압축과 무관함을 적음 |

- 재측정: `analysis/a207_measure.out.txt`(06:08Z) · `analysis/a207_latency.out.txt`(06:08Z), 같은 HEAD `e9266bfb6`.

## 1단계 경량 리뷰 r2 (2026-11-27 — 상한 2회 중 2회, 마지막)

- 같은 리뷰어 · 같은 규칙. 범위: r1 처분이 현재 소스에 맞게 고쳐졌는지와 고치며 새 오류가 생겼는지만 봄.

| # | 등급 | 지적 | 처분 |
|---|---|---|---|
| 1 | P1 | D1b 가 과장됨: SQLite 는 일부 오류(SQLITE_IOERR 등)에서 읽기 문장만으로도 트랜잭션을 되돌리고(modernc.org/sqlite v1.54.0 드라이버 소스), ctx 취소도 `sql.Tx` 를 되돌림. 그러면 오류를 무시해도 뒤의 INSERT · commit 이 실패하고 그 주기는 제출하지 않음 | 수리(문구 — 새 가드 없음). D1b · §5.2 · spec 을 「조회 오류를 판정 오류로 돌려주지 않음. 트랜잭션이 되돌려진 경우는 같은 트랜잭션의 기존 읽기(:439 · :476 · :494 · :794, 하한 판정이면 :451 · :456)와 같은 부류의 실패 — 새 부류 없음, 기회가 문장 둘만큼 늘어남」으로 좁힘. spec 시나리오에 「트랜잭션이 살아 있는 한」을 넣음 |
| 2 | P2 | D6 의 비용이 「Offer 하나」가 아님: `RecordOnly.Notify` 가 로그 한 줄을 동기로 쓰고(record_only.go:51) `Offer` 는 뮤텍스 아래 송신 · 버림 로그(normal_relay.go:42-57). D5 · D6 의 메모리 작업은 순차 루프에서 뒤 포지션의 Place 보다 앞섬 | 수리(문구). §5.3 · §5.4 · D6 를 「그 포지션의 Place 뒤, 뒤 포지션의 Place 는 map 연산만큼 늦어짐」, 「로그 한 줄 + 뮤텍스 아래 송신, 해소 하나에 한 번」으로 고침 |
| 3 | P3 | C2 의 「CANCELLED 무장은 전부 attempt 없음」이 측정과 다름(1 건은 마지막 PLACE 가 `in_doubt_unresolved`) | 수리. C2 를 「12,996 중 12,995 는 attempt 없음, 1 은 in_doubt_unresolved」로 정정 |

- 3회차는 열지 않음(상한). r2 의 셋은 모두 문구 · 수치 정정이며 설계 선택(B2 · N1)을 바꾸지 않음. 2단계 다중 적대 리뷰가 D1b 의 실패 부류 주장과
  §5 의 비용을 다시 봐야 함(design §7.3 ①).

## Function Logic Map

Function Logic Map: not-applicable — 1단계는 기존 Go 함수 본문을 편집하지 않음(생산 · 시험 diff 0). 분기 주장은 FLM 번들이 아니라
`analysis/ast-evidence/`(logic-map AST 21 개, claims.md)를 근거로 씀. 2단계에서 고치는 기존 함수 여섯은 FLM 번들이 필요함(tasks 2.2).

## 변이 원장 · 다중 적대 리뷰

not-applicable(1단계) — 문서 · 측정 하네스뿐임(비례 원칙). 2단계는 청산 · 원장 경로라 High-risk 이며, 변이 원장 M1~M12 와
다른 계열 보이스 둘의 적대 리뷰가 의무임(design §7.2 · §7.3).

## 2단계 — 사람 승인과 Pre-Edit 선언 (2026-11-27)

승인: 사용자 「a207 추천대로 승인」(Manager 전달, 2026-11-27). 결정 1 B2 · 2 콘솔 문구 「반복 거절(요약)」 계열 + 롤백 중 계좌 이력의
`invalid_event_evidence` 수용 · 3 1h 재알림 유지 · 4 해소 알림 일반 등급 · 수용 확정만 · 5 토글 없음 · 6 a206 대상 술어에서 압축 구간의
앵커 행 제외(design §4 D8, a206 브랜치에 같은 술어 반영).

```text
Pre-Edit Gate:
- change id / task id: a207-blocked-exit-retry-dedupe / 2.1~2.4
- 대상 심볼(패키지.함수) — 편집하는 기존 함수 6:
  journal.appendExitEventTx(exit_state.go:792-851) · journal.hydrateExitEventEvidence(exit_snapshot.go:392-417) ·
  engine.ExitObserver.release(exitloop.go:1645-1650) · engine.ExitObserver.submit(exitloop.go:1513-1638, B9 에 호출 하나) ·
  engine.ExitObserver.endExitOrderCycle(exit_order_log.go:235-257, 호출 하나) · operatorview.reasonText(exit_line.go:198-245, case 하나).
  새 함수는 새 파일(journal/a207_compact_evidence.go · engine/a207_unblock_notice.go)에만 둠.
  설계 대비 자리 이동(safe local): D1 의 압축 판정을 recordExitJudgementTx 가 아니라 그 트랜잭션이 부르는 appendExitEventTx 안에 둠 —
  같은 트랜잭션 · 같은 행 값을 쓰고, 59 분기 함수를 편집하지 않아도 됨. 판정 술어 · 결과는 design D1~D3 그대로.
- CodeGraph definition/callers/callees/impact: evidence-reconciliation.md(1.6.0). appendExitEventTx 호출자 5(판정 · 해제 훅 · OPENED ·
  재편입 · 조정 닫힘) — 압축 갈래는 Evaluation 이 있고 action · intent 가 찬 무장 행에서만 열리므로 나머지 넷은 바이트 불변.
  hydrate 호출자 3(AccountExitEvents · BrokerOrderExitLinks · scanExitEvent). release 호출자 submit 하나(5 자리). reasonText 호출자 BuildExitLine 2 자리.
- CodeGraphContext 후보와 evidence reconciliation: not-applicable(1단계 기록과 같음 — 닫힌 열거).
- 기존 동작 파악 근거: 현재 HEAD + analysis/ast-evidence(편집 전 AST 23 개) + 기존 시험 앵커(design §7.1).
- Function Logic Map / Branch Test Map: analysis/function-logic/ 의 6 번들(편집 전 AST 로 열고 편집 뒤 갱신).
- upstream 상속 테스트 영향: no — exit 관측 · 판정 증거는 TossOS 고유 경로.
- 실패 테스트 선행 작성: yes — T1~T8 을 먼저 쓰고 실패를 본 뒤 구현(상수 · 사건 종류만 먼저 선언해 컴파일 실패가 아니라 단언 실패로 RED 를 냄).
- 설정·DB·journal 변경과 rollback: 스키마 무변경(v36). 새 effective_source 값 하나 · 새 unknown 사유 하나 · 새 일반 사건 종류 하나.
  롤백 = 이전 이미지(압축 행은 계좌 이력에서 invalid_event_evidence, 엔진 무영향 — design §6).
- 안전 불변식 §0 위반 여부 검토: 통과 — 해제 5곳 · 재무장 경로에 분기 추가 0, 대기 · 백오프 · 상한 0, Place 앞 작업 0(design §5).
- a166 census: internal/accountbudgetruntime/testdata/a166-wiring/shapes.json 90 키 중 편집 파일 0(rg · json 키 대조).
```

## 2단계 다중 적대 리뷰 r1 (2026-11-27 — 상한 2회 중 1회)

- 보이스 둘, 둘 다 작성자(Claude)와 다른 계열: Codex CLI 0.165.0 `codex exec --sandbox read-only`. ① 적대 Eng(즉시성 · 중복 가드 · 재생 출처),
  ② 원장 · 리더 계약 · 롤백 · 알림. 대상 `git diff 9d694ffb3 -- internal/`(= 6f58010bf). 새 가드 · 기능 제안 금지.
- 두 보이스 모두 **P1 0**.

| # | 보이스 | 등급 | 지적 | 처분 |
|---|---|---|---|---|
| 1 | ① | P2 | T1 은 주기 단위 횟수만 재고 지연 · 경합은 안 잼. T8 은 트랜잭션을 망가뜨리지 않는 오류만 주입함 | 받아들임(문구): design T1 · R7 에 「시간 · 경합 미측정」, D1b 에 이미 적힌 한계를 유지. 새 계측은 만들지 않음 |
| 2 | ① | P2 | T7 이 `ExitProposalForIntent` 에서 멈추고 게이트웨이의 `Critical, true` 분류는 안 봄 | 받아들임(범위): 분류 코드(a166_replay_origin.go:51-55)는 이 change 가 손대지 않았고 기존 a166 시험이 고정함. T7 의 몫은 「압축 행도 action 을 남겨 출처를 입증한다」까지 |
| 3 | ① | P2 | T5 가 action 조건 제거 변이에 대해 공허함(판정 전용 행이 intent 도 없고 arm suppression 도 있어 다른 조건이 대신 막음) — 변이 M6 · M6b 생존과 같은 원인 | 수리: 압축 갈래 조건을 `row.ProposedIntentID != ""` 하나로 줄이고(나머지 셋은 서로를 가리는 군더더기 — saved_monotone 은 발의를 지우고 무장은 늘 recomputed), 술어 호출 계수 시험 `TestA207OnlyArmRowsAskTheRepeatPredicate` 추가(판정 전용 · legacy 행에서 0 회, 무장에서 1 회) |
| 4 | ① | P2 | T4 는 유일 인덱스만 지워도 통과(명시 가드가 남음) | 받아들임: 두 겹 방어이고 둘 다 무변경 — 이 change 는 두 겹 모두에 decision_id 를 남기는 것만 고정함(M5) |
| 5 | ① | P3 | design §5.3 「둘 다 Place 뒤」는 하한 0 주 · Guardian 거절 · intent 무효(Place 없이 해제)에서 거짓 | 수리: 「해제 커밋 뒤」로 정정 |
| 6 | ② | P2 | 연속 해제 셈이 재편입 경계를 넘음 — 재편입은 포지션 id 를 두고 세대만 올리므로 세대 1 의 거절이 세대 2 첫 수용의 「해소」로 붙음 | 수리: 셈에 수명주기 세대를 담고 다른 세대면 새로 셈 · 알리지 않음. 내부 시험 `TestA207AStreakDoesNotCrossALifecycleGeneration`, 변이 M15 · M15b |
| 7 | ② | P3 | 「반복 거절(요약)」이 수용된 재시도 · 취소 뒤 재시도도 거절로 읽힘 | 수리(문구, 사람 결정 2 의 계열 유지): 「반복 거절(요약) — 같은 조건이 앞서 거절·취소된 뒤의 재시도라 전체 근거는 그 조건의 첫 판정 행에 있다」 |

- ②의 확인: 리더 셋 · OpenReadOnly · legacy 판독은 압축 행 외 모양에서 무변경, 롤백(schema 36) 설명(§6)이 e9266bfb6 과 맞음 — 옛 `ExitEvents` 는
  요청 전체가 오류, 계좌 이력은 그 행만 `invalid_event_evidence`, orders 는 모호함이 먼저. 알림은 일반 등급, 거절 critical 의 key · 재알림 무변경.
- ①의 확인: decision_id 는 압축 행에도 쓰이고 명시 가드 · 부분 유일 인덱스가 그대로 섬. 술어는 훅 7 종 + intent 없는
  `ADJUSTMENT_CLOSE_FILL_RECOGNISED` 를 건너뛰고, 현재 쓰기 경로에서 앵커 없는 압축 행은 없음.
- 수리 커밋: `092b15d48`.

## 2단계 다중 적대 리뷰 r2 (2026-11-27 — 상한 2회 중 2회, 마지막)

- Codex CLI 0.165.0 read-only, r1 의 두 보이스 지적을 한 프롬프트로 묶어 수리 커밋 `092b15d48` 만 검증(새 가드 · 기능 제안 금지).
- 판정: **B-P2 해소, 새 P1/P2 0**. 세대별 셈이 맞고(다른 세대의 거절은 셈을 갈아치우고, 수용은 셈을 떼되 같은 세대일 때만 알림), 단순화한
  압축 갈래 조건은 현재 호출자에서 안전함(해제 훅은 intent 를 갖지만 평가가 없어 B2 에서 빠지고, OPENED · 재편입 · 조정 닫힘도 평가가 없음),
  계수 시험은 의미 있음.

| # | 등급 | 지적 | 처분 |
|---|---|---|---|
| 1 | P3 | 새 주석 「훅 행은 intent 가 없음」이 틀림 — `clearExitProposalTx` 가 intent 를 줌. 안전은 평가가 없다는 데서 옴. logic map 도 같은 오류 | 수리(주석 · FLM 문구만, 동작 무변경): 「훅 행은 intent 를 갖지만 평가가 없어 바깥 갈래(B2)에서 이미 빠짐」 |

- 3 회차는 열지 않음(상한).

## 변이 원장 · Function Logic Map (2단계)

- 변이 원장: `analysis/mutation-ledger.md` — 최종 16/16 CAUGHT(1 회차 생존 2 는 조건 단순화로 해소), M10 은 구조상 NA.
- FLM 번들 6: `analysis/function-logic/`. `python3 tools/logic-map/check_analysis.py --change a207-blocked-exit-retry-dedupe` → `evidence complete or diff-proven exempt`.

## 2단계 시험 기록 (최종 코드, 2026-11-27 — 주석 수리 커밋 기준)

| 명령 | rc | 결과 |
|---|---|---|
| `go test -timeout 30m ./internal/journal/ -count=1` | 0 | ok 692.1s |
| `go test -timeout 30m ./internal/app/engine/ -count=1 -run 'Exit\|Refus\|A197\|A094\|A129\|A207\|A091\|A087\|A198\|A193\|A163\|A090\|A074\|A084\|A112'` | 0 | ok 189.6s |
| `go test -timeout 30m ./internal/console/ ./internal/operatorview/ ./internal/obs/ -count=1` | 0 | ok 63.1s · 0.004s · 25.0s |
| `go test ./internal/execgw/ -count=1 -run 'A166\|Replay'` | 0 | ok 24.9s |
| `go test -timeout 45m -tags tossos_testseams ./internal/journal/ ./internal/app/engine/ ./internal/obs/ ./internal/operatorview/ -count=1 -run 'A207\|Exit\|Refus\|A197\|A094\|A129\|A091'` | 0 | ok 170.0s · 224.9s · 0.6s · 0.003s |
| `make lint GOFMT=/usr/local/go/bin/gofmt`(gofmt -l + `go vet ./...` + 태그 vet) | 0 | 위반 0 |
| `python3 tools/logic-map/check_analysis.py --change a207-blocked-exit-retry-dedupe` | 0 | evidence complete |
| `openspec validate a207-blocked-exit-retry-dedupe --strict` | 0 | valid |

- 미실행: 전체 시험(`make test` · `make test-seams` 전 범위 · `make test-race`) · `make gate` · `make sdd-sync`/`sdd-check` — 요청 없음.
  engine 패키지는 위 -run 밖 시험을 돌리지 않았음. a206 쪽 시험은 a206 에 생산 코드가 없어 해당 없음.
- RED/GREEN 증거: `analysis/tdd/red-journal.txt` · `red-engine.txt` · `red-operatorview.txt` · `green-journal.txt` · `green-engine.txt`.
