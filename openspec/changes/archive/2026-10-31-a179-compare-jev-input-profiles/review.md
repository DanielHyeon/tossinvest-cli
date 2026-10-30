# Manager proposal-freeze review — 2026-10-31 KST

Status: Manager proposal-freeze APPROVED for the written scope; Story/Feature/registry and generated PM check PASS reported by Terra. The global OpenSpec symlink was broken; an isolated OpenSpec 1.4.1 is available at `/tmp/tossos-openspec-1.4.1/node_modules/@fission-ai/openspec/bin/openspec.js`. Manager ran `node <that path> validate a179-compare-jev-input-profiles --strict --no-interactive`: exit 0, `Change 'a179-compare-jev-input-profiles' is valid`. Base/AST/Pre-Edit evidence accompanies implementation. Manager authors specification; Terra owns production/test edits. User authorized implementation, targeted verification, gstack review, archive/commit/push.

Scope review: 최초 단순 데이터 종류만 복원함; 초기 고수익 구간 설정 복원이라고 부르지 않음. 질문 J1~J4·모델 고정. 실제 운영/측정 호출 제외.

Architecture review: 기존 PublicState, 수집기, 판단 클라이언트, digest 저장, 라벨 공유 구조를 확장하는 계약임. 신규 수집기·클라이언트·의존성 없음. 구현 전 심볼/AST 증거는 Terra가 별도 기록함.

Adversarial review: (1) 확장 결측 문구 유출 → allowlist; (2) 초기 state의 분봉 제거로 평가 손실 → source digest; (3) 한쪽 실패의 생존 편향 → 실패/정상 쌍 분리; (4) legacy 평가 중복 → 비기본 출력 격리; (5) 관측/응답 시점 혼용 → 기준 시각·가격 고정; (6) standalone/다른 모델 응답 오결합 → cohort와 질문·모델 조건 확인.

DX/QA review (initial freeze, superseded below): 처음에는 current 기본값·명시적 별도 실험 출력을 설계했으나 사용자 후속 지시로 자동 compare 기본값·모드별 자동 출력으로 변경함. 실제 실행을 요구하지 않는 fixture 기반 CLI 검증을 유지함.

Risk: Normal shadow tool, no live trading journal. Production high-risk paths not applicable; shadow provenance/restore tests required. UI/impeccable/Reticle browser: not-applicable (no UI). Full repository tests/gate: not run by explicit user scope. No claim of production deployment or improved profitability.

## Independent Terra contract review disposition

Terra reviewer found two ambiguities before broad implementation. Manager ACCEPTED both: session-wide later source bars must feed KR post-hoc prices (including later judgment failures), and gap rows must retain cohort/profile/source identity even when both profiles fail. Design and scenarios now state these explicitly. These clarify the existing failure/source preservation intent without changing trading rules or adding external collection. This read-only contract pass is not claimed as completed gstack code review.

Intermediate core review found repeated same-second runs could collide when ID had profile but no cohort. Manager required cohort in experimental IDs and reopening/restoration coverage; Terra reports that test passing. Digest wording clarified to existing JSON payload bytes excluding the storage newline. Neither correction changes default current IDs or historic hash semantics.

## Requirement change — automatic comparison (user 2026-10-31)

사용자: initial/compare를 명시적으로 선택해야 하는 방식 대신 자동으로 함께 측정해야 한다고 지시함. Manager ACCEPTED: run 기본값 compare, current는 명시적 opt-out, initial 단독 선택 유지. out 생략 시 모드별 자동 경로, report 기본값도 비교 원장으로 전환함. 운영 service/wrapper를 읽기만 했으며 둘 다 profile/out 미지정이므로 스케줄 수정 불필요함. 새 코드가 운영 체크아웃에 반영되어 새로 실행되어야 효과가 생김; 본 구현에서 실제 호출/재시작을 수행하지 않음.

변경분 검토: 기존 원장 혼합 위험은 별도 자동 경로와 명시적 옛 경로 거부로 통제함. 기존 실행 방식이 바뀌는 것은 사용자 요구이며 명시적 current가 기존 동작을 보존함. 추가 모델 요청은 스냅숏당 한 번임. 실제 source 수집·가격 라벨 조회는 반복하지 않음. 새 기본 경로와 report 연결을 대상으로 fixture 검증을 추가함. 최종 gstack/Manager 검토는 이 개정 계약을 기준으로 수행함.

## US evaluation clarification

Independent reviewer correctly identified ambiguous wording: identical stock-day prices versus the explicitly retained US first-flagged-observation rule. Manager retains the existing US rule (no new trading timing rule). Equal prices/returns are required for the same observation pair; profile-selected observation times may differ. Design now says so and requires reporting the number of US stock-days with different selected entry observations. This is selection-policy performance, not a claim of fixed-time pure input effect. J2 differences remain same-snapshot paired. KR keeps its fixed 10:00→14:10 window. Require divergent-observation fixture and transparent output before closure.

## Manager final verification — 2026-10-31 KST

Verdict: APPROVED for scoped implementation closure, archive and feature-branch commit/push. No authorization to claim deployment or live comparison measurements.

Manager reviewed final implementation diffs (collect/main/ledger/selection, projection/guard, comparison reporting/evaluator), documentation, all 20 named Go race test results, three Python test results, empty successful target-vet output, six-function structural FLM PASS, strict OpenSpec PASS and PM receipt. Source hashes are fixed in analysis/verification.md. The functional QA report shows default synthetic comparison reporting, explicit legacy-output rejection before credentials and no ledger creation on rejection, all PASS. The final documentation correctly describes automatic compare and explicit current rollback.

Independent Terra review findings were repaired: cohort collision/restore association; missing/tampered state validation; malformed pair metadata and null label accounting; later source bars; and transparent US selection timing. Manager additionally caught manual-session restart rejection; its RED/GREEN regression passed. Initial-only summary logging was corrected by the implementer. Fresh native in-host Terra adversarial review (user-directed model substitution for the skill's Claude label), capture token `48dee55b-eddb-40c9-a25f-3e3e600641ac`, found no reproducible material FIXABLE or INVESTIGATE defect; tests were read in summary mode and reviewed source hashes match verification.md. This is not a claim of an outside-model review.

Remaining verification limits accepted for this shadow-tool scope: existing socket-backed tests could not run because loopback bind is denied; they remain unchanged. New HTTP boundary tests use injected in-memory transport. Whole-repository tests/gates and actual API calls are deliberately not run. Branch-table completeness is structural, not exhaustive branch coverage. Checkout-wide CodeGraph/CodeGraphContext/GBrain freshness is not established; current source/AST/test evidence is the review basis. No live trading path, broker collection API, operating service or runtime ledger is changed.

Terra shall record final gstack QA/native receipts, finish applicable task checks, archive/sync the new capability and PM path, commit only the reviewed allowlist, push the feature branch without force, and verify the remote SHA. Source operational checkout remains outside this deployment scope.

### Final delta revalidation

Manager reviewed the final invalid-row accounting correction: profile-tagged rows with missing cohort/source now increment invalid counts rather than silently disappear. The changed Go reporting tests pass under race, all three Python evaluator tests and target vet pass, and functional CLI QA was revalidated (analysis/qa-revalidation/report.md, 2026-10-31 KST). Unchanged coverage from the prior 20-test run is reused rather than represented as a second full run.

Fresh scoped native Terra re-review capture `babe75fd-f279-41f4-b1b7-b9d143ec1965` found no reproducible material defect. Final reviewed source SHA-256: compare_report.go `6ffe3b0a679aaec17a83a213cbf444dc612b2eac606ad89c4dfc5ba307931083`; profile_eval.py `0d351dd267ab66a0f35d28aa61028cc08dae88138cd3ceb2c95d9842c8eeef02`. Changed test hashes `03021d80cd75bf050cdb4f6cce61769d59ea22819910b687b2a9068b41636c44` and `e4e2690a6e1983dc11fea5543812e7eb18b5b974718ed4c3c64fd2841616c099` match verification.md. Remaining production hashes are unchanged. The implementation and reviewed delta are APPROVED; proceed with archive/PM sync, commit and non-force feature-branch push. Current operating checkout and active measurement process are not deployed by this approval.
