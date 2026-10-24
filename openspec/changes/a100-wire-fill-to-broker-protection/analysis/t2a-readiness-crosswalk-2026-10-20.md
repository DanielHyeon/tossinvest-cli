# T2-A 준비 교차표 — 2026-10-20

상태: **준비 문서만**. M-A PASS와 tasks 0.11 raw-status 판정표 동결 전에는 T2-A 구현을 시작하지 않는다. 새로운 구현 승인·브로커 capability 추정·실측 결과·시험 PASS를 만들지 않는다. PM/task 체크는 변경하지 않았다. 현재 다른 작업자가 편집하는 a087 analysis도 건드리지 않았다.

## 정본과 순서

- [tasks](../tasks.md) 0.2·0.11, 2.1–2.7, 3.10이 계약이다. [review](../review.md)의 2026-10-05 T1 acceptance와 [T1 결함 기록](t1-unreachable-branches.md)은 과거 실행 근거다.
- 로트표의 T2-A 범위가 `2.1~2.6`으로 남아 있던 추적 불일치는 Manager가 이번 작업 중 `2.1~2.7`로 정정했다. **새 계약이 아니다.** 기존 task 2.7은 D3 수리를 다른 T2-A 상태기계 배선보다 먼저 닫도록 이미 요구한다.
- 구현 순서: M-A/0.11 → T2-A(2.7 선행) → T2-B API seal → T3. 이후 T4-A child/flat/취소 경계를 닫기 전 T4-B runtime construction/start 금지. T1의 선행 예외는 시험뿐이며 core 수정까지 포함하지 않는다.
- [measurement-prereq](../measurement-prereq.md)는 M-A의 원본 절차다. 오래된 preflight HOLD와 계좌 상태를 현재 상태로 재사용하지 않는다. 신규 LIVE 실행·조회·설정·재시작 승인은 이 문서 범위 밖이다.

## 계약 → 구현 지점 → 후속 검증

| 기존 계약 | 현재 파일/심볼(탐색 앵커) | 재사용 근거와 M-A 입력 | 구현자가 수행할 RED/수용 조건 — 지금 실행하지 않음 |
|---|---|---|---|
| 2.7 D3 | [lifecycle.go](../../../../internal/protectionlifecycle/lifecycle.go) `applyFill`; [state.go](../../../../internal/protectionlifecycle/state.go) `validPending` | [D3 기록](t1-unreachable-branches.md), [실제 재현 probe](harness/t1_probe_fill_during_replace_pending_test.go.txt), 아래 hash 일치 AST. M-A 없이 이미 확인된 결함이나 수정 착수는 T2-A gate를 따른다. | 기존 REPLACE_PENDING/REPLACE_UNKNOWN 부분 fill 반례를 RED로 승격하고 CANCEL_PENDING 대조 유지. 수리 뒤 유효 봉인·누적 수량·재시작 및 다른 시장 전이를 검증. 이미 전송된 pending 명령의 수량을 임의로 재작성하거나 상태를 새로 생성해 결함을 숨기지 않는다. exact repair 설계는 fresh FLM/BTM과 Manager 검토 대상. |
| 0.11 + 2.1/2.2 raw 상태·mapping | [gateway.go](../../../../internal/protectionofficial/gateway.go) `Gateway.adapt`, `lifecycle`; [domain.go](../../../../internal/protection/domain.go) `BrokerProtection`; [protection_reads.go](../../../../internal/official/protection_reads.go) `adaptProtectionConditional` | M-A에서 실제 armed/pre-trigger/triggering/terminal raw 문자열, parent→child ID 노출/조회 순서 확보. 기존 adapter 손실 지점은 [historical current-main-evidence](current-main-evidence.md)의 child-order 절. Gateway 현 분기의 fresh AST는 이 문서에서 만들지 않았다. | lossless mapping 왕복(property)과 실제 M-A fixture. PAUSED/unknown은 미수렴·mutation 금지·alert, ownership 미완료 triggering은 ACTIVE/terminal로 조기 분류하거나 재등록하지 않음. 실측이 없는 상태를 WATCHING과 동등하다고 새로 정의하지 않음. |
| 2.3–2.6 저장·crash | [schema.go](../../../../internal/journal/schema.go); [exit_snapshot.go](../../../../internal/journal/exit_snapshot.go) `scanExitStateResult`; lifecycle `recoverSubmit` | additive-nullable, parent/client/raw/child/generation/desired-at/watermark 및 pending canonical owner가 기존 요구다. 8월 schema·scanner 근거는 현 branch 권위가 아니다. | old rows의 NULL은 보호 미설치이고 기존 exit/fill 유지. newer schema를 old binary가 열 때 ErrSchemaTooNew. plan→pending commit→transport→response commit crash 지점과 재개, 응답 opaque ID로 scope 변경 금지. 스키마 번호는 구현 시 main과 다시 대조. |
| 3.10 child owner 등록 | [fills.go](../../../../internal/journal/fills.go) `TrackedFillOrders`, `confirmedFillOwners`, `RecordFill`; [position_projection.go](../../../../internal/journal/position_projection.go) `resolveFillOrigin` | M-A parent/client/scope/generation/child exact 결속과 owner-before-fill 인과가 전제. 위 심볼은 현재 위치만 확인했으며 최신 분기/transaction 증명은 재생성 대상. [filldetect ledger](../../../../internal/filldetect/ledger.go)는 변경 금지 경계. | 정상 owner 선등록→기존 RecordFill 경로. fill snapshot/event가 먼저 있으면 registrar 거절 + durable ATTRIBUTION_FAILED reconcile/alert; 소급 owner·delta 재생·hook 재실행 금지. duplicate/conflicting ordinary+child owner, 다른 account/day/market/generation, restart와 동시 registrar/RecordFill을 검증. 공식 대사 복구 전 child/position mutation·재등록 금지. |
| 3.10 watermark | 같은 journal owner/fill 경계 + lifecycle 적용 | durable cumulative fill snapshot을 읽어 lifecycle watermark를 멱등 전진한 뒤 다음 mutation 계획. [fills tests](../../../../internal/journal/fills_test.go)는 기존 scope/replay 회귀 후보이며 신규 child 계약 증명은 아님. | 중복 registrar·중복/부분 fill·crash/reopen·순서 역전에서 canonical owner와 watermark 한 번 전진, 타 포지션/시장 보존. 기존 일반 주문의 external/late-owner 시험 기대를 protection child의 소급귀속 허용으로 복사하지 않음. |

이 표는 기존 계약을 실행 가능한 검토 항목으로 연결한다. 신규 schema/API/worker framework를 설계하거나 authority를 발급하지 않는다. source 파일이 존재한다는 사실은 runtime 배선/승인과 다르다.

## AST 재사용 경계

이번에는 기존 AST의 `source_sha256`과 실제 파일 SHA256만 비교했다. AST 추출·인덱스 갱신·시험 실행은 하지 않았다.

| bundle | 결과 | 허용되는 사용 |
|---|---|---|
| [applyFill AST](function-logic/internal-protectionlifecycle--applyfill/ast.json) | MATCH | 기존 D3 분석의 현재 소스 대응 근거 |
| [prepareRegister AST](function-logic/internal-protectionlifecycle--prepareregister/ast.json) | MATCH | 기존 T1 준비등록 분석 재사용 |
| [validPending AST](t1-ast/validPending.ast.json) | MATCH | 기존 pending 등식 분석 재사용 |
| [official adapter AST](function-logic/internal-official--adaptprotectionconditional/ast.json) | MATCH | 해당 official adapter에 한정; protectionofficial Gateway와 혼동 금지 |
| [scanExitStateResult AST](function-logic/internal-journal--scanexitstateresult/ast.json) | STALE | 현 분기 근거로 사용 금지. fresh AST/FLM/BTM 전 의미 추론 없음 |

[current-main-evidence.md](current-main-evidence.md)의 제목은 2026-08-15 base를 뜻한다. 전체 문서를 오늘 current로 올리지 않는다. 특히 현재 journal owner/scanner/RecordFill의 호출 순서·분기 수·원자성은 그 과거 prose로 인증하지 않는다. 현재 위치 확인은 분석 범위를 정하는 탐색일 뿐이다.

## 고정된 읽기 스냅샷

HEAD: `3415cf89f713ccdf61604963f29da9e0d5116294`

- `internal/protectionlifecycle/lifecycle.go`: `de50441bc89c79ec5cfeb8308a837db1cffede7e3ab52c661eccb9515d1688e5`
- `internal/protectionlifecycle/state.go`: `df5c5459c6d2add80bcfdcadb03f4af1dbcb19cce1424687c2d855a40fd66cbe`
- `internal/official/protection_reads.go`: `3983478598ac261eae4b6101be509d22b94b8645598c0e893fe28ab5ccb68487`
- `internal/protectionofficial/gateway.go`: `0fa5df0f4af3190bad2126e78125f0f09193f705945f5607eda52f8be5a79ebf`
- `internal/protection/domain.go`: `c446e3bba27b66af3e6d46f0a8ee43eeb1fd12eba1b6503a4baba052141af0c5`
- `internal/journal/schema.go`: `602b8ec0587ab6f01f4edea222657107b6685a63c9b56db94d67162a059b6bde`
- `internal/journal/fills.go`: `92bc44f947eead2a045815b29358122d2fd0dd2ffd3e5bada4f2dc44eaa5c4c5`
- `internal/journal/position_projection.go`: `5b62d380b860e67884ceb5f5e217a12fa0d8a190c54190f370ae711133002748`
- `internal/journal/exit_snapshot.go`: `19ebe0eeb9879bcbe16664a39d2ed1785c8cadb55dc97bc3b3eb1455980ff5ce`

읽기 결과 schema 상수는 36이다. 이것은 새 migration에 37을 예약하는 승인이 아니다. 코드 착수 때 다른 작업자의 최신 schema와 다시 맞춘다.

검증 범위: 로컬 Markdown 링크 존재 및 문서 whitespace 확인만. 제품/시험/PM 수정, tests/gate, 운영·계좌·브로커 호출 0. 이 문서의 최종 수용은 gstack 교차리뷰와 Manager 검토 후다.
