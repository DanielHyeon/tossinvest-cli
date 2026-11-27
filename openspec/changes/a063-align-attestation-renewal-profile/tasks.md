## 1. Evidence and proposal freeze

- [x] 1.0 Reserve `STORY-TOS-a063`, create the paired OpenSpec change, and capture
      `da80ce31b6a1ab5d443016768f970a82bab102db` as the pre-implementation base.
- [x] 1.1 Capture the installed attestation unit/timer, both resolved record paths, current timer result
      and active attestation expiry using read-only commands; redact account and credential data.
- [x] 1.2 Record the source-of-truth path flow and select the existing operator surface for the pre-expiry
      warning.
- [x] 1.3 Complete independent security/operations review of this proposal and record the decision in
      `review.md` before implementation.

## 2. RED

- [x] 2.1 Add a contract test that fails when the repository-backed renewal command omits the explicit
      console config profile or overrides only one of record/output.
- [x] 2.2 Add a contract test that fails when the service masks a refused `soak attest` exit status.
- [x] 2.3 Add a test for visible renewal failure/impending expiry that preserves running-engine and
      startup-interlock behavior.

## 3. GREEN

- [x] 3.1 Add the minimal repository-backed service definition using the explicit console profile.
- [x] 3.2 Preserve renewal exit status and expose bounded failure/expiry details on the selected existing
      operator surface.
- [x] 3.3 Keep the survey and attestation path read-only and preserve every operating setting.

## 4. VERIFY and approved operations

- [x] 4.0 *(Retired by a125 on 2026-09-29 — the adoption exception no longer exists; kept as history.)* Complete the explicitly retrospective execution-baseline adoption using the accepted a120
      tools: preserve P and historical maps, bind real independent adversarial and subsequent gstack
      reviews to complete immutable inventories, and verify all E-based maps at clean detached H.
- [x] 4.1 Run focused tests, `make test`, `make vet`, `make validate`, `make sdd-sync`,
      and `make sdd-check`; retain actual results. The final change gate follows operational proof.
- [x] 4.2 With explicit human approval, back up and install the service definition and reload user
      systemd without enabling or changing any trading control. Record approval scope, target profile,
      reviewed template digest and approved survey window before the corresponding operations.
      *(2026-10-22 사용자 "승인". 조작 전 기록 커밋 `ae7000458`, 06:18 KST 설치 — 백업
      `~/.config/systemd/user/a063-backup-20261021T211830Z`, 설치본 digest = 검토 digest, 바이너리는 승인 범위 밖이라
      교체하지 않음. `analysis/adoption/task-4.2-4.4-operation-2026-10-22.md`.)*
- [x] 4.3 Start the console-profile survey in the approved window and retain at least three consecutive
      qualifying days, recording actual dates. The original 2026-08-29 deadline was not verified and
      has passed; current evidence must not be backdated or replaced by synthetic test records.
      *(진행 중: 같은 프로필 survey 가 이미 실행 중이라 중단·중복하지 않음. 셀 일자 2026-10-22(UTC)~, 가장 이른 판정
      2026-10-25 09:00 KST 이후, 상한 2026-10-29(UTC). 기록 파일 §6.)*
      *(완료 2026-11-28 관측: 2026-10-22·10-23·10-24(UTC) 연속 pass — 창 안. 기록 파일 §7.2.)*
- [x] 4.4 Verify a fresh attestation is written to the engine profile, renewal failures are visible, and
      no live engine restart or order mutation is used for verification.
      *(부분: 2026-10-22 06:18 KST 새 attestation(만료 2026-11-20T21:18:31Z)·renewal-status `issued`·console 표시 확인,
      재시작·주문 0. 운영 중 실패 가시성 관측은 아직 없음 — 3일 창 뒤 재판정. 기록 파일 §5~§6.)*
      *(완료 2026-11-28: 새 attestation 2026-11-27T15:00Z→12-27, 실패 가시성 운영 관측 = 2026-11-12~14 refused 9회가 systemd
      `exit-code` 로 남음, 범위 상실의 `issued` 빈틈은 §5 가 고침. 재시작·주문 0. 기록 파일 §7.3.)*
## 5. a196 이관 범위 — 범위를 잃은 갱신은 `issued` 만으로 기록하지 않는다 (사용자 결정 2026-11-28, design D6)

- [x] 5.1 편집 전 AST·FLM 번들(`runSoakAttest`·`RenewalStatus.Validate`·`validReasonCode`·`renewalReasonMessage`·
      `TestSoakAttestRecordsIssuedStatusBesideTheResolvedProfileAttestation`)과 CodeGraph callers 를 남긴다. High-risk 아님 판정 기록.
      *(2026-11-28: `analysis/scope-absorb/pre-edit-ast/` HEAD `baeddfe31`, 번들 `analysis/function-logic/` — 판정 design D6.)*
- [x] 5.2 RED: 감독 증거 없이 발급한 갱신이 `issued` + `live_endpoint_uncovered` 를 남기는 시험, `Validate` 가 `issued` 에 그 코드
      하나만 받는 시험, 콘솔이 그 코드를 고정 문장으로 보이는 시험을 먼저 쓰고 오늘 코드에서 빨강을 확인한다.
      *(2026-11-28: 시험 셋 모두 FAIL — `analysis/scope-absorb/red.txt`, 상수만 먼저 넣어 컴파일은 됨.)*
- [x] 5.3 GREEN: 최소 구현. 종료 코드·attestation 내용·발급 기준·인터록 불변.
      *(2026-11-28: `go test -count=1 ./internal/soak ./internal/console ./cmd/tossctl ./internal/attest` rc 0 — `analysis/scope-absorb/green-packages.txt`.)*
- [x] 5.4 선택 시험 실행, 번들 갱신(같은 파일 편집으로 낡은 `newSoakAttestCmd`·`Console.readAttestation` 포함), 경량 gstack 리뷰 1회
      (최대 2회차)와 처분 기록.
      *(2026-11-28: `analysis/scope-absorb/review-r1.md` — F2 수용(문장에 게이트 조건), 나머지 거절·근거, r2 수렴.)*

## 4 (계속). 재고정과 종결 — §5 뒤에 한다

- [ ] 4.4.1 Immediately before the final gate (after 4.2~4.4), re-pin the comparison base by the `docs/WORKFLOW.md`
      「사람 승인 base 재고정」 procedure: re-measure attribution from P (`da80ce31`) to the then-HEAD (own non-merge Go
      commits touching this directory, plus a human check for Go commits split from their document commit), re-verify every
      attributed function's bundle with the positive-assertion receipt of a125 (`analysis/harness/a063_receipt.py`), and if the
      receipt HEAD differs from the re-pin parent, measure again. Commit `base-commit.txt` alone with the receipt (a125 decision (나)).
- [ ] 4.5 Complete independent diff/test review, PM synchronization and archive only after all evidence is
      recorded. Run `make gate CHANGE=a063-align-attestation-renewal-profile` after all prerequisite
      tasks are complete; archive is contingent on its success and Manager acceptance, never on task
      checkboxes alone. If that final gate fails, reopen this task and retain the failure evidence.
