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
