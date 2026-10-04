모델: GPT-6 기반 Codex · 세부 버전 미공개 · 「~/.codex 미접근」

**REJECT — 구조적 MUST NOT 차단 증명이 부족합니다.**

질문 1~7을 검토했습니다. 아래 재현은 **사본의 시험 스케치**이며, 현재 생산 취약점·실주문 발생을 뜻하지 않습니다. 구현되지 않은 SHADOW 시험 전체가 통과했다고 주장하지도 않습니다.

| 등급 | 지적 | 근거·재현 | 결과 | 제안 |
|---|---|---|---|---|
| **P0** | 봉투 없는 반환형으로 dispatch 접근이 차단되지 않음 | `worker.go:165`, `coordinator.go:203`; `TestReviewShadowStructure` | OFF/OFF·문자열 결과형을 유지하면서 `worker.owns`를 Submit 함수 값으로 재결속. 이름 기반 허용 목록 통과, **실제 조정자 큐 깊이 1** | SHADOW 입력의 dispatch 권한 제거. 호출 대상의 타입 해석과 경계 소스 동결·독립 리뷰 추가 |
| **P1** | 인용 census의 알려진 한계와 보완책 누락 | `strategyhandoff/mint_census_test.go:17`; `TestReviewCensusAliasAndErasure`, `TestReviewShadowSamePackageMint` | 중첩 별칭·`any`·generic instance alias 미검출. 같은 router 패키지의 쌍둥이 구조체 변환으로 **Verified=true, 네 레인 ON**, 인코더 참조 0 | 선례의 **source freeze까지** 적용. 타입 변환·주조 경계도 검사 |
| **P1** | 기존 레인 입력을 사용하면 반사실 대상 유실 | `strategy_lane_runtime.go:266`, `strategy_market_coordinator.go:130`; `TestReviewShadowInputTap` | 미선언 활성화에서는 종목별 승자만 남음. 선언됐지만 닫힌 활성화에서는 **입력 0건** | 실제 관문·중재 **이전**의 전체 레인 제안을 SHADOW로 분기하는 지점 명시 |
| **P1** | `Shadow(shadow,input)`는 활성화 ON 우선순위를 판단할 정보가 없음 | 설계 브리프:51-53; `worker.go:94-113`; 스케치 실행 | 동일 인자에서 활성화 OFF·ON 모두 WOULD_EMIT | 같은 파도의 desired/effective가 **둘 다 OFF**인 자격을 호출 경계에서 강제 |
| **P1** | “Validate가 모든 읽기 경로의 관문”은 부정확 | `strategy_runtime_projection.go:49-60`; `TestReviewReadValidationBoundary` | 동적 레인 투영을 덧씌운 `Context.Read`는 잘못된 상태에도 nil error. 별도 Validate가 거절 | 외부 경계 검증으로 주장을 좁히거나 최종 조립 후 검증 |
| **P1** | 정규 직렬화가 생성기 작성 출처를 증명한다는 논증 오류 | `design.md:240,246`, `production_family_activation.go:509-527` | 정규 바이트 등식은 누가 작성했는지 구분하지 못함 | 신뢰 앵커를 **배포 핀 관리 권한**으로 명시. 생성기 사용은 운영 정책으로 구분 |
| **P2** | SHADOW 실패의 durable lifecycle 분리 미정 | `strategy_lane_runtime.go:208,252,301` | 재사용 후보 evaluate에는 복구·유계 실행·잠금 기록이 존재. 정상 재시작 시험만으로 실패 경로 무기록은 증명되지 않음 | panic·timeout·취소 및 기존 원장 행을 둔 상태에서 전후 변화 0 검증 |
| **P2** | 브리프의 옛 계약 요약·재시작 조건 잔존 | 브리프:3-5,70 ↔ spec:89-93 | `signed` 잔존. “파일 없음”과 개정된 “핀 없음”이 혼용됨 | 파일 유지/핀 제거, 핀 유지/파일 제거를 별도 시험으로 고정 |

**나머지 판정:** 레인 한정 반사실과 digest 핀 전환 자체는 수용 가능합니다. 골든은 기본값만 고정하고 OpenAPI는 runtime enum을 열거한다는 주장은 맞습니다. 공유 검증 함수와 별도 권한 타입의 구분도 타당합니다. **8.6 BLOCKED와 A100 선결 조건은 유지**되어야 합니다.

**공격에도 유지된 방어**

- 영값 shadow → NOT_SHADOWED, 실제 기존 Run의 영값 activation → DORMANT.
- 직접 Envelope 필드는 census가 검출.
- 인코더의 별칭·dot import 우회는 기존 가드가 검출.
- 실제 RPC 읽기는 unknown `shadowOutcome`을 허용하지만 `runtime=SHADOW`는 거절. 네트워크 없는 메모리 transport로 실행했습니다.
- 기존 재시작 시험은 OFF/OFF/UNOBSERVED 및 lease/latch/recovery 0건으로 통과했습니다. 아직 SHADOW 구현 검증은 아닙니다.

[전체 리뷰·질문별 근거·재현 안내](/tmp/claude-1000/-mnt-D-Axipient-workspace-TossOS/4b178b0a-b918-4419-8142-f266a03d8e82/scratchpad/codex-tree-shadow/review-experiments/review.md) · [최종 시험 로그](/tmp/claude-1000/-mnt-D-Axipient-workspace-TossOS/4b178b0a-b918-4419-8142-f266a03d8e82/scratchpad/codex-tree-shadow/review-experiments/final-tests.log)

**실저장소 시작·종료 비교: 동일.** 기존 변경사항도 그대로입니다.

```text
HEAD: c7219640d08408a650941c62cf0df6efff77c036
HEAD 비교: cmp exit 0
status --short 비교: cmp exit 0

 M openspec/changes/a112-run-four-strategy-families-independently/tasks.md
?? .reticle-setup-crash.log
?? openspec/changes/a112-run-four-strategy-families-independently/analysis/review-shadow-freeze-2026-10/
?? openspec/changes/a112-run-four-strategy-families-independently/analysis/shadow-2026-10/
?? w4.log
```

사본의 검토 대상 추적 파일 **399개가 c7219640과 바이트 동일**, go.mod/go.sum도 동일합니다. 추가 시험·보고서·캐시는 사본 내부에만 작성했습니다.