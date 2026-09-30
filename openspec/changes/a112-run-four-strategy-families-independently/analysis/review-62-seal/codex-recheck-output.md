**APPROVE — 수리 범위 승인. 접근 위반 기록: 앞선 턴에서 `~/.codex/memories/MEMORY.md`를 검색했으며, 이번 재확인에서는 추가 접근하지 않았습니다.**

읽기 전용 정적 검토입니다. 테스트·변이는 실행하지 못했습니다. 커밋된 실행 기록은 확인했지만 독립 재실행 결과로 주장하지 않습니다.

| 항목 | 판정 | 근거 |
|---|---|---|
| **#1 공유 배열** | **CLOSED** | 생성자가 분리 함수를 호출하고 KR·US 각각 새 배열을 만듭니다. loader 생성 **후** 외부 원소를 교체하는 양 시장 시험도 추가됐습니다. [생성자:210](/tmp/claude-1000/a112-review-e5a335bb/internal/app/engine/strategy_account_first_leg_authority.go:210), [배열 분리:62](/tmp/claude-1000/a112-review-e5a335bb/internal/app/engine/strategy_first_leg_owner_scope.go:62), [시험:247](/tmp/claude-1000/a112-review-e5a335bb/internal/app/engine/a112_first_leg_owner_scope_seal_test.go:247) |
| **#2 census 간접 경로·빌드 제약** | **CLOSED — 지적 경로 기준** | 함수 값 별칭도 식별자 언급으로 집계하고, 봉인 필드 주소·슬라이싱을 금지합니다. 호스트 빌드 여부로 제외하지 않으며 `!cgo`와 GOOS/GOARCH 접미사를 거절합니다. [언급 검사:277](/tmp/claude-1000/a112-review-e5a335bb/internal/strategyflow/seal_census_test.go:277), [빌드 모델:67](/tmp/claude-1000/a112-review-e5a335bb/internal/strategyflow/seal_census_test.go:67) |
| **#3 직접 선택 시험·진술 정정** | **시험 CLOSED / 주석 OPEN(T)** | 두 범위 각각의 반환 identity와 없는 범위 거절을 직접 확인합니다. 설명은 정정됐지만 함수 내부에 옛 거짓 주석 한 줄이 남았습니다. [직접 시험:225](/tmp/claude-1000/a112-review-e5a335bb/internal/app/engine/a112_first_leg_owner_scope_seal_test.go:225), [잔존 주석:168](/tmp/claude-1000/a112-review-e5a335bb/internal/app/engine/a112_first_leg_owner_scope_seal_test.go:168) |
| **#4 하위 시험·대조군 사건** | **CLOSED — 명시한 보장 기준** | 하위 이름의 PASS를 요구할 수 있고, 요구 부모 아래 skip을 실패 처리합니다. 변이 대조군도 명령마다 시험 PASS 사건을 요구합니다. [하위 시험 처리:34](/tmp/claude-1000/a112-review-e5a335bb/openspec/changes/a112-run-four-strategy-families-independently/analysis/harness/verify_named_tests.py:34), [대조군 확인:378](/tmp/claude-1000/a112-review-e5a335bb/openspec/changes/a112-run-four-strategy-families-independently/analysis/harness/a112_lot_mutate.py:378) |
| **새 생산 결함·토글 OFF** | **CLOSED — 새 위반 미발견** | 생산 변경은 생성자 저장식·분리 함수·주석으로 한정됩니다. 발급 조건과 비활성 시장 분기는 바뀌지 않았습니다. 정상 입력의 값을 보존하고 외부 교체의 전파만 차단합니다. [분리 구현:62](/tmp/claude-1000/a112-review-e5a335bb/internal/app/engine/strategy_first_leg_owner_scope.go:62), [비활성 분기:40](/tmp/claude-1000/a112-review-e5a335bb/internal/app/engine/strategy_dispatch_handoff.go:40) |

**분리는 완전한 깊은 복사가 아닙니다.** 원소 안의 route에는 slice가 남고, `ProductionAuthority.weekly`도 포인터를 공유합니다. 다만 현재 봉인이 비교하는 `Result`·계보·실행 조건은 값 필드이며, `WeeklyBinding()`은 스칼라만 든 구조체를 복사해 반환합니다. 따라서 확인한 잔여 공유가 기존 dispatch 교체 경로로 봉인 원본을 바꾸지는 못합니다. [route:167](/tmp/claude-1000/a112-review-e5a335bb/internal/strategyrouter/production.go:167), [weekly 접근자:124](/tmp/claude-1000/a112-review-e5a335bb/internal/strategyproposal/production.go:124), [binding 구조:42](/tmp/claude-1000/a112-review-e5a335bb/internal/journal/strategy_first_leg_atomic.go:42).

실행 기록도 구분해야 합니다. `red-6.2-seal-fix.log`는 초기 단일 시험 실패 기록이고, **KR·US 각각 발급되는 RED는 재기록한 `red-6.2-seal.log`에 있습니다.** [양 시장 RED:31](/tmp/claude-1000/a112-review-e5a335bb/openspec/changes/a112-run-four-strategy-families-independently/analysis/measurements/lot-6.2-seal/red-6.2-seal.log:31).

새 발견·잔여 사항은 요청대로 **5.2.2.2 이월**합니다.

- **P2:** 대조군의 추가 JSON 실행은 `completed.returncode`를 검사하지 않습니다. 일부 PASS 뒤 실패해도 앞선 실행이 GREEN이면 대조군 GREEN으로 기록할 수 있습니다. 사건 수와 종료 코드를 함께 확인해야 합니다. [하네스:383](/tmp/claude-1000/a112-review-e5a335bb/openspec/changes/a112-run-four-strategy-families-independently/analysis/harness/a112_lot_mutate.py:383).
- **(T):** 남은 “`entries[0]`이면 identity 거절” 주석 삭제. 현재 직접 시험의 유효성에는 영향 없습니다.

Recommendation: APPROVE — 기존 P0·P1 지적 경로의 수리를 인정합니다. 잔여 P2·주석은 5.2.2.2로 이월하며, 개수 관문 제거의 안전성까지 승인한 것은 아닙니다.
