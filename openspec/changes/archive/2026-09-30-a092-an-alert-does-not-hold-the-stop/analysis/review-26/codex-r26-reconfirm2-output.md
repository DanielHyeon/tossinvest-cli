**R1·R2·R3 닫힘. 이번 범위에서 새 P0/P1은 찾지 못했습니다. 복합 실패의 CLI 표시에 P2 한 건이 남습니다.**

현재 코드와 `R26B-FIX.diff`의 정적 판정입니다. 시험·변이 통과 주장은 근거로 사용하지 않았습니다. 승인된 `Flush`·`ClaimDisposition` 이월은 제외했습니다.

| 기존 항목 | 등급 | 판정 | 파일:줄 | 근거(코드 인용) |
|---|---|---|---|---|
| **R1 — 반납 선점 무기록** | P1 | **닫힘** | [alertdelivery.go:365](/tmp/claude-1000/a092-r26-tree/internal/app/engine/alertdelivery.go:365) | `releaseOK && (AlreadySettled || LeaseLost)`에 `EventAlertClaimLost`, `alert_id`, `outcome`을 기록한다. 시도 기록이 `Applied`여도 반납에서 발견한 선점이 남는다. 뒤의 오류·시도 한도 판정은 계속 실행된다. |
| **R2 — 재조회 실패를 전달 완료로 표시** | P1 | **닫힘** | [modeops.go:132](/tmp/claude-1000/a092-r26-tree/internal/app/engine/modeops.go:132), [engine_mode_release.go:107](/tmp/claude-1000/a092-r26-tree/cmd/tossctl/engine_mode_release.go:107) | 현재 모드 실패는 `ReReadError`, 통지 목록만 실패하면 `NoticeReadError`로 분리한다. 후자는 이미 읽은 `Mode·Seq·EntryBlocks`를 보존한다. CLI는 모드 미확인 표시를 먼저 정하고, 통지 조회 실패도 `NoticePending`보다 먼저 분류한다. “이미 전달 처리됨”은 제거됐다. |
| **R3 — 미지 결과를 선점 로그로 분류** | P2 | **닫힘** | [notifier.go:681](/tmp/claude-1000/a092-r26-tree/internal/obs/notifier.go:681) | `NotFound`·`AlreadySettled`를 먼저 처리한 뒤 `Outcome != SettleLeaseLost`는 `EventAlertUndelivered` 오류로 기록한다. `ClaimedBy` 분기는 이제 `LeaseLost`에만 적용된다. 이 함수에 `Applied`를 넘기는 호출 경로는 확인하지 못했다. |

| # | 등급(P0/P1/P2) | 파일:줄 | 무엇이 틀렸나 | 근거(코드 인용 또는 정적 반례) | 제안 |
|---|---|---|---|---|---|
| 1 | **P2** | [engine_mode_release.go:136](/tmp/claude-1000/a092-r26-tree/cmd/tossctl/engine_mode_release.go:136) | **통지 기록 실패와 통지 목록 조회 실패가 함께 있으면, 텍스트 CLI에서 조회 실패가 사라진다.** | `Release`는 통지 기록 실패 뒤에도 재조회를 계속하므로 `NotifyError != "" && NoticeReadError != ""`가 가능하다. `modeReleaseNoticeState`는 첫 분기에서 `NotifyError`만 반환하고, 출력 함수에는 `NoticeReadError`를 별도로 출력하는 자리가 없다. JSON은 두 필드를 보존한다. | 통지 기록 결과와 통지 목록 조회 오류를 독립적으로 표시한다. 복합 실패 조합을 확인한다. |

요청한 경계별 판정은 다음과 같습니다.

| 확인 항목 | 판정 | 코드 근거·한계 |
|---|---|---|
| **(a) `detachedAnnouncer`가 커밋 전 작업을 분리하는가** | **아니오** | [modeops.go:106](/tmp/claude-1000/a092-r26-tree/internal/app/engine/modeops.go:106)은 원래 요청 `ctx`로 `TransitionOperatingMode`를 호출한다. 원장은 같은 문맥으로 `BeginTx`, 조회, INSERT를 수행하고, audit → `Commit` → 투영 → announcer 순서다([operating_mode.go:475](/tmp/claude-1000/a092-r26-tree/internal/journal/operating_mode.go:475)). `WithoutCancel`은 실제 announcer 호출 안에서 적용된다. |
| **요청 취소 뒤 audit·commit이 절대로 없다는 주장** | **그 절대 표현은 근거보다 강함** | audit의 `RecordAction`에는 원래부터 context 인자가 없다. 취소가 audit 실행을 즉시 멈춘다는 보장은 없고, 취소와 commit의 경합도 별개다. 이번 wrapper가 **커밋 전 취소를 제거하는 우회**를 만들지는 않았다. |
| **(b) 고정 문구가 400/500 분류를 지우는가** | **아니오** | [mode_control_transport_unix.go:160](/tmp/claude-1000/a092-r26-tree/internal/app/engine/mode_control_transport_unix.go:160)의 `errors.Is` 분류를 먼저 유지한다. 입력·승인 오류는 400, 표면 부재는 503, 나머지만 고정 문구의 500이다. |
| **통지 실패와 재조회 실패 구분** | **API는 유지, CLI 복합 실패만 위 P2** | `NotifyError`, `ReReadError`, `NoticeReadError`가 각각 존재한다. 커밋된 전이의 `Changed·TransitionID`도 보존된다. |
| **(c) `logFailure` 자체의 계좌 가림** | **현재 계좌 원문의 동일 문자열은 가린다** | [modeops.go:173](/tmp/claude-1000/a092-r26-tree/internal/app/engine/modeops.go:173)은 `err.Error()` 전체에서 `strings.TrimSpace(o.accountRef)`를 `ReplaceAll`한다. 현재 원장이 `%s`로 넣는 해당 계좌와 계좌 포함 키는 이 방식으로 가려진다. 다른 표기·다른 계좌까지 가리는 일반 마스커는 아니며, 이번 신규 호출 경로에서 그 누락의 구체적 생산 반례는 찾지 못했다. |
| **(d) 차단·승격·거절 판정 변경** | **분류 수리가 판정을 완화하지 않음** | 반납 선점 추가는 로그뿐이다. `Runtime.escalate`는 원장 호출 뒤 오류 설명만 분리한다. `RiskGuardian.escalateFor`도 `%w`를 보존하며, 호출자는 이미 결정된 `chainRefusal(verdict)`에 오류를 합친다([riskguardian.go:430](/tmp/claude-1000/a092-r26-tree/internal/execgw/riskguardian.go:430)). |
| **취소 문맥 변경의 실제 효과** | **의도된 동작 변경 있음** | 커밋 후 요청 취소만으로 통지 기록·재조회가 실패하지 않게 된다. 따라서 그 취소 때문에 생기던 기록 실패 차단·재승격도 피한다. 실제 기록 오류에 대한 `Block`·`escalate` 경로는 유지된다. |
| **필드 제거가 원장 사건을 바꾸는가** | **아니오** | `withoutFields`는 값 사본의 `Fields`만 비운다. 로그 뒤 `recordCritical`에는 원래 `e`가 전달되어 키·payload·재무장 입력은 유지된다([record_only.go:50](/tmp/claude-1000/a092-r26-tree/internal/obs/record_only.go:50)). |

**“원문 오류는 모두 `logFailure`에서 가려진다”는 주장은 성립하지 않습니다.** 통지 기록 실패는 [record_only.go:140](/tmp/claude-1000/a092-r26-tree/internal/obs/record_only.go:140)에서 원문을 먼저 로그에 남기고, `ModeOperations.Release`의 `NotifyError` 분기는 `logFailure`를 호출하지 않습니다. 원장 오류에는 계좌 포함 사건 키가 들어갈 수 있습니다([record_alert.go:26](/tmp/claude-1000/a092-r26-tree/internal/journal/record_alert.go:26)).

다만 이 경로와 기존 `FieldAccount` 로그는 [Manager 기록:4982](/tmp/claude-1000/a092-r26-tree/openspec/changes/a092-an-alert-does-not-hold-the-stop/review.md:4982)에 **명시적으로 이월**돼 있습니다. 이번 신규 결함이나 재차 BLOCK 사유로 계산하지 않았습니다. 새 가림의 보장 범위는 `logFailure`를 통과하는 오류입니다.

파일 생성·수정, 시험·게이트 실행은 하지 않았습니다. 판정은 정적 수리 재확인이며 배포 완료 판정은 아닙니다.

Recommendation: R1·R2·R3 수리는 수용 가능합니다. CLI의 통지 기록·목록 조회 동시 실패 표시를 보완하고, 계좌 가림의 완료 주장은 `logFailure` 적용 범위로 한정하십시오.