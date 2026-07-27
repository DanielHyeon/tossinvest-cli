# gstack 리뷰 기록 — harden-execution-base (proposal-freeze)

- 일시: 2026-07-26 · 보이스: codex CLI(ENG/SAFETY/SPEC-QUALITY, 22건) + 독립 Claude Eng 적대(14건) + 독립 Claude CEO/DX(13건) — **총 49건**
- 대상: proposal/design/tasks + specs 5건 · 전부 실코드 검증 기반(orders_reads/orders_write/hybrid/root.go/push 등)

## 합의표 (Eng 축)

crash window 커버 PARTIAL/PARTIAL → 재설계 · shim 회귀 PARTIAL(alias 안전 검증됨, newAppContext 무보호) · 인터록 건전성 **NO/NO**(availableActions 부재, 조용한 WTS 폴백, MCP 우회) · 스펙 테스트 가능성 PARTIAL · task 크기 PARTIAL/NO · live 안전 완결성 NO → **전면 개정 결정**

## 수용·반영 (요약)

1. **상태 모델 재설계**: Intent/MutationAttempt/브로커 주문 노드 분리, mutation별 IN_DOUBT, lineage edge를 journal 트랜잭션으로 (codex-1, Eng-4)
2. **IN_DOUBT 해소 재정의**: 부재≠FAILED — fingerprint + OPEN/CLOSED pagination 완주 + 연속 N회 안정화 + 잔고 delta 교차, 불능 시 UNRESOLVED_IN_DOUBT 운영자 해소, 자동 재제출 무조건 금지, 심볼당 in-flight 1건 (codex-2, Eng-2)
3. **DISPATCH 단계 분리**: RECORDED→DISPATCH_STARTED, RECORDED-only는 NOT_DISPATCHED 안전 종결 (codex-3)
4. **체결 멱등성 재설계**: per-fill ID 부재 → 누적 스냅샷 + 양의 delta, 감소·역순 fail-closed (codex-4)
5. **상태 파생 함수**: OPEN/CLOSED 전제, 우선순위 표, UNKNOWN_BROKER_STATE fail-closed, fixture는 upstream→실측 보강 순서로 역전 해소 (codex-6, Eng-3)
6. **pagination 완주 어댑터** 선행 task화 (codex-5)
7. **cancel/amend 사전 확인 대체**: official에 availableActions 부재 → OrderByID 파생, WTS 만료 상태 테스트 필수 (Eng-1)
8. **엔진 fail-closed 배선**: official 직접 구성·config 무시·자격증명 없으면 기동 거부·정적 import 테스트·전 matrix WTS-spy 0회 (Eng-6, codex-17)
9. **ExecutionGateway 봉인**: GuardianDecision(one-shot nonce) 필수, raw mutator 미노출; MCP 우회는 P4까지 잔존 리스크로 proposal에 문서화 (codex-10, Eng-7)
10. **journal 의무 스코프**: 엔진 프로필 경로 한정 — CLI/MCP upstream 유지로 §0.2 보존 (CEO/DX-1, Eng-5)
11. **flatten-all saga 분해**: cancel-all/reduce-only 2task, --dry-run 리허설, 확인 문자열 강화(계좌·nonce·TTY) (codex-12/13, Eng-11, CEO/DX-4)
12. **알림 등급화**: critical → journal DB outbox + 지속 실패 시 진입 차단, heartbeat 방식, Phase-2 이벤트는 enum 예약 (codex-16, Eng-10)
13. **journal 내구성 계약**: XDG data 경로, FS allowlist, BEGIN IMMEDIATE+synchronous=FULL, 손상 기동 거부, P2 import 계약 (codex-8, Eng-13, CEO/DX-8)
14. **retry matrix 표 산출물화** + 429/staleness 모순 해소 (codex-7)
15. **reconciliation 스냅샷 계약**: 고정 순서·as-of·부분 폐기·epsilon·카운터 리셋 (codex-9, Eng-12)
16. **characterization 테스트 선행** (1.2), DoD 확장 (codex-18, Eng-8)
17. **clock/TZ task 추가**(2.0) + 시간 규율 requirement (CEO/DX-3, Eng-9)
18. **섹션 5 분리** → `verify-execution-capability` change + attestation 기동 인터록으로 기계 연결 (CEO/DX-2, codex-15, CEO/DX-11)
19. **task 분해·앵커·ID 정합·High-risk 태그 정정·Pre-Edit 스코프** (codex-20, CEO/DX-5/6/7/9/10)
20. **어휘 정정**: 레인→심볼 차단, 원장→journal, reason-code enum (CEO/DX-12, codex-22), 메트릭 서버→P4, ntfy 추상화 제거 (CEO/DX-11)
21. **audit task**(4.2), Guardian 인터페이스 "초안" 표기 (Eng-14)

## 기각·보류

- 단일 writer lease를 P1에 구현 (codex-11): **보류 → P4** — P1에는 데몬 프로세스 자체가 없어 경합 표면이 실존하지 않음. proposal에 잔존 리스크 명기로 대체
- soak를 P1 코드 change에 유지 (원안): 분리로 대체 (18)

## 재검증

개정 후 `openspec validate --strict` 통과 (양 change). 반영 커밋: 본 커밋.
