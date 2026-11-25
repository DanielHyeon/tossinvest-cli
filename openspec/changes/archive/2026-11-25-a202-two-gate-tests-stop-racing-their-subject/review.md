# a202 review

## 리뷰 등급과 생략 사유 (2026-11-25)

- **gstack 리뷰: 생략.** 사용자 지시 「간단한건 리뷰 할 필요 없음」(Manager 전달). 근거: 시험 전용 change — 바뀌는 것은 기존 시험 함수 두 개의 대기
  문장뿐이고 생산 diff 0(`git diff <base>..HEAD --stat -- '*.go' ':!*_test.go' 'schemas/'` 빈 출력, `analysis/verify/production-diff.txt`).
- **변이 원장: not-applicable** — `docs/WORKFLOW.md` 「비례 원칙」: High-risk 경로 · 게이트 판정을 바꾸지 않는 시험 전용 변경. 대신 사본 대조 실험이
  그 자리를 맡는다: 편집 전 RED(생산 unlock↔close 사이 sleep 에서 옛 시험 FAIL · 부모 지연에서 옛 시험 FAIL), 편집 뒤 같은 지연에서 GREEN,
  양성 대조(사본에서 생산 닫기 호출 삭제 → 새 시험 FAIL), 무변이 대조군 GREEN(design §3, `analysis/red/` · `analysis/green/`).
- **다중 적대 리뷰: not-applicable** — 같은 비례 원칙. 주문 · 위험관리 · 원장 · 대사를 건드리지 않는다.
- **proposal-freeze 리뷰:** 경량 — `openspec validate --strict` + 작성자 셀프 검토 + Manager 검증(WORKFLOW 「리뷰 게이트」의 UI · 문서 · 도구 등급).

## Function Logic Map

- 적용(면제 아님). 편집한 기존 Go 함수 2개(시험 함수)마다 번들 — `analysis/function-logic/internal-shadowsession--testchildlockandcrash` ·
  `cmd-tossctl--testthereplacedreaderisclosed`. 편집 **전에** 만들어 커밋했고 편집 뒤 AST 를 재추출해 갱신했다.
- 도우미(`a109WaitFor` · `fixture`)는 고치지 않았다.

## Pre-Edit 선언 (2.x 착수 전)

- 대상 심볼: `TestChildLockAndCrash`(`internal/shadowsession/session_linux_test.go:210`), `TestTheReplacedReaderIsClosed`(`cmd/tossctl/a109_the_reattachment_does_not_strand_itself_test.go:191`).
- 호출부: 없음(시험). 생산 호출 대상 `Begin` · `strategyRuntimeAttachment.attempt` 는 바꾸지 않는다.
- a166 census: 대상 파일 · 함수명이 `internal/accountbudgetruntime` 에 없음(`analysis/code-context/a166-census-check.txt`).
- 불변식: 생산 diff 0, 단언 · 기한 · 헬퍼 불변, 새 가드 · 헬퍼 · 의존성 0.
- 실패 시험(RED): `analysis/red/` 네 파일.
- rollback: 커밋 revert. 설정 영향 없음.

## 구현 · 검증 기록 (2026-11-25, 구현 팀메이트)

| 실험 | 원문 | 결과 |
|---|---|---|
| RED 자식 직접 실행(편집 전) | `analysis/red/shadowsession-child-before.txt` | `READY` 뒤 0.10 s exit 2, `fatal error: all goroutines are asleep - deadlock!` |
| RED 사본 부모 300 ms 지연 + 옛 시험 | `analysis/red/shadowsession-old-test-parent-delay.txt` | 3/3 FAIL `lock: shadowsession: invalid or unavailable state`(게이트 1회차와 같은 문구) |
| 대조 무변이 사본 + 옛 tossctl 시험 | `analysis/red/tossctl-old-test-control.txt` | 5/5 PASS |
| RED 사본 생산 unlock↔close 50 ms + 옛 시험 | `analysis/red/tossctl-old-test-prod-sleep.txt` | 5/5 FAIL 「0번 닫았다, want 1」(게이트 2회차와 같은 문구) |
| GREEN 자식 직접 실행(편집 뒤) | `analysis/green/shadowsession-child-after.txt` | 5 s 뒤에도 살아 있음(probe 가 kill, rc 137), stderr 빈 줄 |
| GREEN 사본 부모 300 ms 지연 + 새 시험 | `analysis/green/shadowsession-new-test-parent-delay.txt` | 3/3 PASS |
| 대조 무변이 사본 + 새 tossctl 시험 | `analysis/green/tossctl-new-test-none.txt` | 5/5 PASS |
| GREEN 사본 50 ms + 새 시험 | `analysis/green/tossctl-new-test-prod-sleep.txt` | 5/5 PASS |
| 양성 대조 사본 생산 닫기 호출 삭제 + 새 시험 | `analysis/green/tossctl-new-test-prod-noclose.txt` | 5/5 FAIL 「밀려난 reader 닫기: 2s 안에 일어나지 않았다」 |
| `-count=200` | `analysis/verify/count200-*.txt` | 각 200/200 PASS |
| 두 패키지 전체 무태그 · `tossos_testseams` | `analysis/verify/packages-*.txt` | 둘 다 ok, rc 0 |
| `make lint` · gofmt | `analysis/verify/make-lint.txt` · `gofmt.txt` | rc 0 · 위반 0 |
| 부하 속(엔진 시험이 도는 동안, loadavg 12/20코어) | `analysis/verify/under-load.txt` | 두 시험 `-count=50` 100/100 PASS — **판별력 없음**: 옛 시험은 같은 부하에서 안 돌렸다 |
| gate ⑤ 단독 | `analysis/verify/check-analysis.txt` | required 2, rc 0 |
| gate ⑪ 단독 | `analysis/verify/gate11-validate-step.txt` | (가) strict valid · (나) 오류 0, rc 0 |
| 생산 diff | `analysis/verify/production-diff.txt` | 빈 출력 |

- 하네스: `analysis/harness/copy-harness.sh`(워크트리를 `openspec/` 빼고 사본으로 복사 → `git init`·커밋 → `rev-parse --show-toplevel` 이 사본인지 단언 →
  변이 적용 → `go test`) · `analysis/harness/child-probe.sh`(자식 모드 직접 실행). 사본 경로 · 시작 sha 는 각 원문 머리에 있다. 실제 worktree 의 생산 코드는
  건드리지 않았다(`production-diff.txt`).

## Manager 검증 (2026-11-25) — **구현 수용** (Manager 전달 원문을 옮겨 적음)

- 브랜치 HEAD `05ffeefd3`, 작업 트리 깨끗. `git diff 7bc35e5d1..HEAD -- '*.go' ':!*_test.go' 'schemas/'` 빈 출력(생산 diff 0). Go diff 는 시험 2파일 +6/−1 — 읽고 동의.
- 독립 재실행: `go test ./internal/shadowsession -run '^TestChildLockAndCrash$' -count=100` ok 1.735s, `go test ./cmd/tossctl -run '^TestTheReplacedReaderIsClosed$' -count=100` ok 0.249s.
- RED 원문(자식 0.10 s 뒤 exit 2) · 양성 대조 원문(닫기 삭제 사본에서 5/5 FAIL 「밀려난 reader 닫기: 2s 안에 일어나지 않았다」) 확인.

### Manager 결정 (2026-11-25)

1. `skip_specs: true` 수용 — 생산 동작 불변 · 요구사항을 지어내지 않는다. 종결 때 `openspec archive` 의 표지 처리를 실측해 적고, archive 전후
   `openspec/specs/` 무변을 `git diff --stat -- openspec/specs` 로 보인다.
2. a166 prototype `TestKilledLockOwnerReleases` 의 같은 결함은 여기서 고치지 않는다 — a166 `issues.md` 에 한 항목만 추가(`issues.md` I2).
3. `!inFlight()` 대기 자리는 a202 `issues.md` I1 에 후속 후보로(분석 안 함 명시). 1분 고아 상한 수용(I3).

### 종결 선행 (2026-11-25, 종결 에이전트 = Opus 5.5 팀메이트)

- `git fetch origin` → origin/main `65a274aec`(a166 time-namespace 로트 7커밋; Go 변경은 `internal/clock/bootstamp.go` 와 a166 시험들 — a202 대상 파일과 겹침 0)를
  브랜치에 merge(rebase 아님) → `da3455960`.
- merge 뒤 두 패키지 전체(`-count=1`, `-trimpath` 없음): `analysis/close/selected-tests-merge-notag.txt` — shadowsession ok 1.561s · cmd/tossctl ok 68.139s (rc 0);
  `analysis/close/selected-tests-merge-tossos_testseams.txt` — ok 2.235s · ok 69.459s (rc 0).
- tasks 4.1 체크 · 4.2 선체크(gate ② 선행 조건) · tracker 재생성 → 커밋 `edee1b435`.
- 착지 기록 `check_analysis.py --change a202-two-gate-tests-stop-racing-their-subject --record-landing` rc 0 → `landed-commit.txt` =
  `05ffeefd3ff560f0d19fbfdb3cd36de669b9c865`(도구 계산, base `7bc35e5d1`, 원문 `analysis/close/record-landing.txt`).

### 종결 (2026-11-25)

- 격리 worktree `~/tossos-work/a202-close`(ext4 홈, detached `7a9de809c`): `make sdd-infra` rc 0 → `make sdd-sync`(시한으로 감싸지 않음, 14:03:51~14:39:29,
  rc 2 — advisory `codegraphcontext index` 300 s · gbrain project sync 시한 초과, CodeGraph(hard) 성공, HEAD 무이동 `7a9de809c`) → `codegraph status .`
  워밍업 3회 `Index is up to date`(0.64 · 0.56 · 0.55 s). 원문 `analysis/close/sdd-sync.txt` · `codegraph-status.txt`.
- **gate 1회차 rc 0 — 11/11 GATE PASS**(14:40:06~15:21:48, 시작 load average 5.43 · 5.48 · 6.08): ① tasks OK ② 미완료 0 ③ 짝 선언 없음 ④ review.md OK
  ⑤ `base 7bc35e5d1bcb → landed-commit 05ffeefd3ff5… required 2 function(s) — judged at HEAD 7a9de809c985` evidence complete
  ⑥ sdd-check OK(advisory codegraphcontext · gbrain WARN, `CodeGraph hard-evidence index matches the worktree`) ⑦ make test OK(shadowsession ok 4.463s ·
  cmd/tossctl ok 96.464s) ⑧ test-seams OK(ok 4.505s · ok 101.117s) ⑨ test-race OK ⑩ vet OK ⑪ (가) 닫는 change strict valid (나) 저장소 전체 오류 0
  (advisory strict 6 passed / 51 failed). `--- FAIL` 0 줄. 원문 `analysis/close/gate-1.txt` · `gate-1-rc.txt` · `gate-1-meta.txt`.
- archive: `openspec archive -y`(1.16.0, `--skip-specs` 플래그 **없이** 표지만) → rc 0, 출력에 「Specs to update」 없음 · proposal Why 1000자 초과 경고 1(non-blocking).
  `skip_specs` 실측: 이 change 는 `specs/` 가 없어 갱신할 delta 가 0 이고 archive 는 spec 을 건드리지 않았다 — `openspec/specs/` 전체 sha256 합 archive 전후
  `7546634446eb…` 같음, `git diff --stat -- openspec/specs` 전후 빈 출력(`analysis/close/openspec-archive.txt` · `specs-{before,after}-archive.txt`).
- Story 경로 갱신, tracker 재생성 · `--check` current, `openspec validate --all` Totals 56 passed · 0 failed rc 0(`analysis/close/validate-all.txt`).
- 운영 반영 없음(시험 전용, 이미지 · 컨테이너 · 토글 · 브로커 호출 0).
