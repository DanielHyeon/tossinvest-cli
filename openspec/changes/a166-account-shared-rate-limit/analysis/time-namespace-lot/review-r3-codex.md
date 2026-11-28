지정 diff에서 **P3 문서 집계 오류 1건**을 발견했습니다. 승인된 계약을 벗어나는 동작 변경이나 머지 차단 결함은 찾지 못했습니다.

**[P3] 변이 검출 건수가 1건 과다 집계됨**

- 근거: [mutation-ledger.md:8](/home/daniel/tossos-work/a166-time-namespace/openspec/changes/a166-account-shared-rate-limit/analysis/time-namespace-lot/mutation-ledger.md:8)은 `CAUGHT 13`으로 기재합니다.
- 반증: [mutation-run-2.log:2](/home/daniel/tossos-work/a166-time-namespace/openspec/changes/a166-account-shared-rate-limit/analysis/time-namespace-lot/mutation-run-2.log:2)의 전체 항목은 **CAUGHT 12 · SURVIVED 1 · CONTROL 1**입니다. 원장 표도 같은 구성입니다.
- 수정안: 제목의 `CAUGHT 13`을 `CAUGHT 12`로 정정. 재실행이나 정책 변경은 필요 없습니다.

**A. 코드·시험 검토**

- [bootstamp.go:64](/home/daniel/tossos-work/a166-time-namespace/internal/clock/bootstamp.go:64): B2·B3·B4·B8의 namespace 불일치 항만 제거했습니다. 표본 형식, uptime 역행, 다른 boot 보류, anchor 선택, 만료 산술은 유지됩니다. 제시된 표본 계약 아래 의도 밖 동작 반례는 발견하지 못했습니다.
- [bootstamp_test.go:43](/home/daniel/tossos-work/a166-time-namespace/internal/clock/bootstamp_test.go:43): 기존 세 시험에서 여전히 필요한 단언을 삭제하지 않았습니다. 첫 시험은 역행 거절을 유지하면서 다른 boot anchor의 전체 기간 보류를 고정합니다. 두 번째는 namespace 차이 허용과 duration 0에서도 역행 거절을 고정합니다. 세 번째는 기존 early-anchor 거절을 유지하고 다른 namespace의 정상·역행 anchor를 추가합니다.
- [bootstamp_crossns_test.go:15](/home/daniel/tossos-work/a166-time-namespace/internal/clock/bootstamp_crossns_test.go:15): 실제 잔여 시간과 `NeedsAnchor`를 함께 검사합니다. 단순 오류 없음만 확인하는 시험이 아닙니다. duration 0 + anchor 역행 사례는 B4 삭제를 직접 검출합니다.
- **M6 논증은 타당합니다.** B7을 통과하면 now와 base는 같은 boot입니다. base=saved이면 B3, base=anchor이면 B4가 역행을 먼저 거절합니다. 정확히는 B8의 **오류 반환 경로**가 도달 불가이며, 해당 조건문 자체는 실행될 수 있습니다.
- [observation_crossns_linux_test.go:39](/home/daniel/tossos-work/a166-time-namespace/internal/accountbudget/observation_crossns_linux_test.go:39)의 fake는 **순차적인 다른 참여자의 표본 입력**을 대표합니다. `ObservationStore`는 hint를 객체에 캐시하지 않고 잠금 아래 파일을 다시 읽으므로 복원·Finish·prune 경로 검증은 유효합니다. 다만 독립 프로세스의 재개방·경합·실제 Snapshot 검증 자체를 대표하지는 않습니다. 별도 `run-after-fix.log`는 실제 provider와 독립 실행의 순차 재현 기록이며, 이번 리뷰에서 재실행한 증거는 아닙니다.

**B. 계약 문안 대조**

기준: `analysis/time-namespace-lot/contract-delta.md`. 줄바꿈·괄호 등 편집상 차이는 제외했습니다.

| 대상 문장 | 판정 |
|---|---|
| design.md:77 — 같은 boot에서 잔여 시간 계산 | 같음 |
| design.md:77 — 현재=자식 namespace, 두 offset 0 | 같음 |
| design.md:77 — namespace 신원이 달라도 같은 시계로 비교 | 같음 |
| design.md:77 — 신원은 표본 내 비교·저장 형식 검사에 유지 | 같음 |
| design.md:77 — 신뢰된 procfs·실행 중 namespace 조작 없음 | 같음 |
| design.md:77 — 무결성 실패 시 poll fail-closed | 같음 |
| design.md:101 — offset 0·현재=자식 검증, 참여자 신원 일치 불요 | 같음 |
| proposal.md:24 — 서로 다른 offset 0 namespace 지원 | 같음 |
| poll-clock-contract.md:12 — 첫 지원 범위 보강 | 같음 |
| poll-clock-contract.md:20 — 동일 boot 차감·역행 거절·다른 boot 비차감 | 같음 |
| poll-clock-contract.md:28 — 양성 사례·커널 재현 증거 추가 | 같음 |
| spec.md:86 — namespace가 바뀐 재시작 포함 | 같음 |
| spec.md:90 — 다른 namespace 참여자의 유효 hint 복원 WHEN | 같음 |
| spec.md:91 — 신원 차이 허용·만료 삭제·무효 표본 거절 THEN | 같음 |

같은 문서의 다른 문장과 새로 생긴 모순은 발견하지 못했습니다. 실행 중 namespace 교체 금지와 서로 다른 고정 namespace 참여자 지원은 양립합니다.

검토 기준 HEAD는 `33e4074cf`, 작업 트리는 깨끗했습니다. 소스·시험·AST·원문 로그를 정적으로 대조했으며 시험 실행·파일 변경·네트워크·docker·브로커 접근은 하지 않았습니다. ai-memory MCP는 제공되지 않았고 CLI 조회는 네트워크 금지 경계를 지키기 위해 수행하지 않았습니다. 과거 메모는 현재 판정 근거로 사용하지 않았습니다. `impeccable: N/A — backend/documentation-only`.

머지 가능