## Context

기준 HEAD `5f3506599fd9cbd40b888aa954ae880244bedf9b`. a159는 보호 복합 회귀 완료이며 후속 admission 구현은 아니다. 정본 strategy-engine의 continuation 8:4:2, reversal 2:4:8, weekly 최대 7 leg와 immutable budget 계약을 그대로 유지한다. Manager 직접 재조사에서 세 전략의 계획/위험 모델과 첫-leg 주문 경계, plan 원문 복원의 한계를 확인했다. 관련 사전 AST는 `/tmp/tossos-a160-prep/` 및 `/tmp/tossos-nextleg-plan-evidence/`에서 보존 후 analysis에 복사한다.

## Goals / Non-Goals

**Goals:** 현재 원장에 있는 상태만 동일 snapshot으로 읽고, 없는 full immutable plan/budget preimage를 unavailable로 명시한다. 후속 권한 설계의 작은 선행단계다.

**Non-Goals:** actual-risk 복원, monetary bucket 합산, sellable 조회, qFinal 계산, decision/예약/lease 발급, 후속 신호/engine/CLI/API/UI 연결, schema·기존 함수 변경, 서비스 실행/재시작/배포/실주문. a156 완료나 live-ready를 주장하지 않는다. a157·기존 증거·원본 checkout은 보존한다.

## Decisions

1. 새 journal 파일에서 기존 **ReadOnly 핸들만** 조회 메서드를 제공한다. 별도 DB 연결/프레임워크를 만들지 않고 기존 `positionCampaignQuery`와 position/leg/order scanners를 재사용한다. 기존 공개 CAS 메서드를 바꾸거나 writer API를 추가하지 않는다.
2. 요청은 account, market, symbol, campaign ID, expected actual position generation (>0)을 필수로 한다. 기존 시장/종목 정규화를 따르되 빈 값과 KR/US 밖 시장은 거절한다. account/campaign는 exact identity다. nil context/handle·닫힌 handle·취소 context·schema 불충족은 정상 오류이며 부분 snapshot을 반환하지 않는다.
3. 한 read transaction에서 campaign 및 동일 account/market/symbol의 **최신 position**과 projection version을 읽는다. requested generation, campaign actual generation, latest position generation이 같아야 한다. position은 OPEN, campaign은 현재 actual position에 결속된 비-CLOSED 상태여야 한다. 보호 때문에 RECONCILE/entry-blocked인 상태는 관측 가능해야 한다. PROSPECTIVE, missing position/version, 예전 generation과 새 position 혼합, scope 불일치는 typed unavailable/error로 처리한다. 이를 0·FLAT·빈 목록의 성공으로 치환하지 않는다.
4. position의 비어 있지 않은 EntryDecisionID와 campaign.DecisionID도 정확히 같아야 한다. 같은 scope/generation이라도 다른 decision 결속은 unavailable/error다. 결과는 campaign의 기존 state/version/stop/latch, position의 기존 quantity/state/generation 및 authoritative projection version, sequence순 legs와 sequence/order-ID순 watermarks다. campaign ID 및 watermark-leg 관계가 정확해야 한다. leg가 없는 actual campaign, orphan watermark, 깨진 scan은 부분 성공이 아니다. watermark가 아직 없는 유효 planned leg는 빈 watermark 목록으로 표현 가능하며 체결 0의 근거로 추론하지 않는다. 읽은 값은 저장된 사실이지 새 보호 적합성 판정이나 replay audit 결과가 아니다.
5. `FullPlanUnavailable`에 해당하는 안정된 typed reason을 모든 성공 snapshot에 명시한다. 현 schema로 전체 계획/고정 예산 원문을 확정할 수 없기 때문이다. budget=0, known-risk=0, Ready=true, qFinal, decision, lease 등의 필드를 추가하지 않는다. 5개 monetary dimension을 더해 campaign stop-risk라고 부르지 않는다. 이 표시는 실제 주문 경로를 차단하는 신규 gate가 아니라 비권한 조회의 한계다.
6. SQL snapshot은 읽기 transaction의 최초 조회 시점 기준이다. 반환 후 최신성·broker 권한을 보장하지 않는다. external 시세/계좌/운영 DB 조회는 없다. 테스트는 임시 DB와 별도 연결의 동시 writer만 사용한다.

## Verification

- 공개 ReadOnly 경로에서 실제 기존 synthetic admission/fill helper를 가능한 한 재사용해 current campaign position/leg/watermark를 확인한다. KR/US 및 peer scope는 별도 대조한다.
- 잘못된 account/symbol/campaign/generation, 최신 generation 교체, projection version 결손, prospective/closed, blocked/reconcile 관측, orphan/missing/오류를 검증한다. 오류는 usable partial snapshot을 남기지 않는다.
- 결정적인 barrier로 첫 read 뒤 다른 연결이 fill/version/watermark를 commit하게 하고 같은 read가 전후 상태를 섞지 않음을 증명한다. sleep 기반 확률시험은 피한다. production test hook/공개 seam은 추가하지 않는다. 기존 driver fixture 또는 내부 tx reader의 실제 공개 wrapper 결속을 검증하는 최소 fixture를 사용한다.
- close/reopen 같은 저장 상태의 결과 동일성과 조회 전후 logical table rows 불변을 확인한다. 물리 SQLite 파일 hash는 WAL/checkpoint 때문에 row 무변경의 대체 근거가 아니다. concurrent test는 의도한 writer의 변화와 reader mutation을 분리한다.
- 테스트 경계가 transaction/scope 누락을 실제로 검출하도록 최소 한 개의 disposable behavioral mutation 또는 결정적인 fault injection을 남긴다. 컴파일 실패는 검출 증거가 아니다. focused/race, 독립·gstack 리뷰, Manager source/receipt 검증, 전체 gate를 수행한다.

## Research / Skill Decisions

ai-memory-retrieval과 search-first 적용. 연결 memory 도구 부재 및 explicit upstream 프로젝트 CLI not-found/전역 빈 조회는 성공한 기억 회수로 표시하지 않는다. 내부 code/scanner/read-only DB를 재사용할 수 있으므로 외부 패키지/새 의존성·라이선스 변경 없음. CodeGraph/AST는 현재 소스 SHA와 조정하고 CGC 혼합 worktree 결과는 advisory다.

grill-me/grilling: 사용자 승인된 순서·오프라인·기존 정책 보존을 적용했다. 추가 매매 비중/예산을 다시 결정하지 않는다. 기술적 reader 범위는 Manager가 위와 같이 확정했다. impeccable: N/A — backend read-only library, 사용자 화면 없음.

## Risks / Trade-offs

읽기 결과를 주문 승인으로 오인 → authority fields/Ready 제거 및 unavailable reason 고정. 다른 시점 혼합 → 단일 tx와 동시 writer 검증. 레거시 원장 누락 → 오류/불가이지 zero 복원 아님. 읽기 범위 확대 → risk bucket/full event replay/새 schema를 이번에 추가하지 않는다.

## Migration Plan / Open Questions

전체 gate에서 `TestExitEligibilityHasOneSpelling`이 새 reader의 `EntryDecisionID` 사용 미등록을 탐지했다. Manager가 해당 검사와 reader를 직접 대조했다: exact campaign/position decision 연결 확인이며 일반 청산 자격 판정이 아니다. `internal/position/eligibility_drift_test.go`의 기존 `eligibilitySpellers` map에 새 reader 파일과 이 사유만 등록한다. 검사 함수·토큰·청산 술어는 변경하지 않는다. 실패 원본을 보존하고 이 최소 delta를 독립·gstack 재검토한 뒤 최종 sync/check/전체 gate를 재실행한다. 새 정책 결정이나 운영 변경은 없다.

migration 및 운영 반영 없음. 새 사용처 없는 읽기 코드·시험을 추가한다. 기존 `readonly_test.go:TestTheReadOnlyHandleHasNoWriteMethods`의 정확한 허용목록에는 새 SELECT-only 메서드 이름 하나와 이유만 등록한다. 보호 assertion을 완화하지 않으며 이 test-only 편집은 사전/사후 AST·도구가 요구하는 FLM/BTM 증거를 갖춘다. 다른 기존 함수 편집이 필요해지면 작업 전에 Manager에 보고한다. 실제 후속 admission과 durable plan 보존은 별도 단계이며 이 change 완료로 해소됐다고 쓰지 않는다.
