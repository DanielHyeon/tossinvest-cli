# a166 runtime lot3a 독립 검토

판정: **조건부 PASS — 고정 후보에서 재현된 P0/P1 없음.** 대상 `d2ce75b6229107a8bc5f8d3c9ef1c9ef93a806f1`; 사용되지 않는 내부 factory 검토다. 전체 a166 완료·운영 배선·3 req/s 실계좌 보장이 아니다.

## P0–P3와 잔여 판단

- P0/P1: 지정 범위에서 재현된 신규 결함 없음. 시험을 통과하지 않은 모든 환경의 부재 증명은 아니다.
- P2 배선 조건 — `internal/accountbudgetruntime/runtime_linux.go:50`: `FromEnv`는 변수가 없으면 `ModeOff`를 돌려준다. required 참여자가 이 helper의 부재 결과만으로 실행되면 무예산 OFF가 될 수 있다. 현재 후보에는 production 호출자가 없고, 작성자도 launch 책임으로 분명히 기록했다. 후속 lot에서 독립적인 required launch 선언/참여자 목록으로 누락을 wire 0으로 거부하는 검사와 fixture가 필요하다. 이 부분승인은 운영 활성화 권한이 아니다.
- P2 증거 범위 — `runtime_test.go:148`: 두 자식은 순차적으로 정상 root와 복제 root를 열어 거절을 검사한다. 이는 동시에 실행되는 모든 참여자의 namespace 일치나 TLS 첫 plaintext Write rolling rate를 입증하지 않는다. `runtime_test.go:218`의 두 actual official client는 같은 Budget의 bootstrap 차단·요청 완료를 검증하지만 첫 Write 시각을 직접 계측하지 않는다. 후속 배선에서 직접 계수·운영 전 namespace 검증이 필요하다.
- P3/지원 경계 — `runtime_linux.go:141`: manifest는 외부의 신뢰된 launch 입력이다. 동일 UID/관리자에 의한 의도적인 변경, bind-mount alias 및 전체 파일시스템 변경을 방어했다고 주장하지 않는다. `Budget.VerifyIdentity`의 검증과 후속 `Acquire` 재검증도 영구적인 원자적 namespace 봉인이 아니다. 이는 문서화된 지원 조건이다.

## 근거

- AST 도구로 `Open`, `readManifest`, `Factory.NewClient`, `Factory.Close`, `Factory.Budget`, `Budget.VerifyIdentity`를 대상 SHA에서 각각 추출했다(`/tmp/a166-runtime-independent-*.json`). 구조 열거는 AST; 파일 소유·비교·수명 순서의 해석은 현재 원문을 손으로 읽었다.
- `runtime_linux.go:64-101`: explicit mode와 required 입력을 판별하고 NOFOLLOW/openat 기반 manifest를 읽은 뒤 `clock.System`으로 기존 group을 열고 외부 기대 root/queue/wire device+inode를 비교, 실패 시 budget Close를 오류에 결합한다. `readManifest`는 파일 mode 0600·UID·regular·Nlink·최대 4096 bytes와 strict 중복/미지정 JSON 검사를 적용한다. 파일 자체에 저장된 기대 ID가 신뢰 가능한지는 launch trust boundary다.
- `identity_linux.go:18-38`, `budget_linux.go:171-244`: 초기 비교는 pinned root/lock 통계와 기대값을 대조하며 각 lock 획득에서도 root/lock identity를 재검증한다. 두 검증 사이는 별도 시점이며 전체 환경의 원자적 봉인으로 해석하지 않는다.
- `runtime_linux.go:103-138, 276-295`: NewClient/Budget/Close는 factory mutex로 직렬화, required에서 budget/cleanup 누락은 오류, 활성 owner가 있으면 Close 오류를 유지한다. OFF는 기존 `official.New`를 사용한다. concrete required client는 `official.NewWithAccountBudget`를 통해 token hc와 guard가 결속된다(`internal/official/account_budget.go:11-37`). 반환 Budget은 신뢰된 내부 caller에게 빌려준 포인터이며 본 API 자체가 trade authority·readiness를 보증하지 않는다.
- 동일 최종 commit의 작성자 `a166-runtime-final4-{normal,race}.txt` 원문 및 15개 테스트 이름을 확인한 뒤 별도 detached ext4 `/home/daniel/tossos-work/a166-runtime-review`에서 같은 이름 선택 normal/race를 재실행했다. Go 1.26.5: `/tmp/a166-runtime-independent-normal.log` 2.499초 PASS, `/tmp/a166-runtime-independent-race.log` 5.584초 PASS. normal/race 모두 local TLS·tempdir fixture만 사용. 전체 패키지 목록 대신 명시된 15개 테스트만 선택했다.
- 수동 gstack checklist 범위: mode completeness, state/identity trust boundary, lock lifecycle, error propagation, raw fallback, client construction, unsupported 플랫폼 및 source/시험 범위 검토. 설치·업그레이드·telemetry·specialist 실행 없음. Go1.25 별도 재실행·다른 OS 실행·전체 테스트/gate·실계좌/운영·실제 power loss는 미실행.
- 코드/시험은 수정하지 않았다. `git diff --check HEAD^ HEAD` PASS, review worktree의 tracked status clean, CodeGraph는 미초기화다. runtime source SHA-256 `524a354b569c9b8783a97fca0e12c00e7a94dbbb746b3d2b5dd8e04d657e0f4c`; identity source SHA-256 `913196b6f71847798e076a49b73bbe9ef18ce04b5db0c30a5775b6f7fd605c0b`.

참고만: production exported observer를 추가하지 않고 `accountbudget/export_test.go` test-only helper와 `package accountbudget_test`에서 runtime factory를 import하여 내부 firstWrite seam을 사용할 수 있는지 별도 연구할 수 있다. 이 배치의 컴파일·실행은 이번 read-only 범위에서 수행하지 않았다.
