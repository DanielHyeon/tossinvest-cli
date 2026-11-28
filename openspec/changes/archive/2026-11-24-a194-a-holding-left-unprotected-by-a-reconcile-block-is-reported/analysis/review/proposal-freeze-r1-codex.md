# proposal-freeze 리뷰 1회차 — codex (원문)

- 날짜: 2026-11-23. 보이스: codex-cli 0.165.0(OpenAI 계열 — 작성자 Claude 와 다른 계열), `codex exec --sandbox read-only --ephemeral -C <worktree> -o … "<프롬프트>" </dev/null`, rc 0.
- 대상: worktree `fix/a194-unprotected-holding-report`(base `064f59fd9`)의 1단계 문서(미커밋 워킹트리).
- 프롬프트 원문:

```text
TossOS change a194 의 proposal·design·spec delta·tasks 를 적대적 Eng 관점으로 리뷰하라. 대상 디렉터리: openspec/changes/a194-a-holding-left-unprotected-by-a-reconcile-block-is-reported/ (proposal.md, design.md, specs/exit-policy/spec.md, specs/reconciliation/spec.md, tasks.md; 증거는 analysis/ 아래).
이 저장소는 실제 돈을 다루는 Go 자동매매 엔진이고, 이 change 는 대사(reconcile)·편입(adoption) 경로에 알림을 더한다.
계약이 코드와 정본 spec(openspec/specs/exit-policy, openspec/specs/reconciliation, openspec/specs/engine-safety)과 맞는지 현재 소스로 직접 확인하라. 핵심 코드: internal/app/engine/adoption.go(judgeHoldings, blocked, alertUnmanaged, alertAdoptionFailed, unmanagedFact, reconcileUnmanagedKey), internal/app/engine/reconcileloop.go(RunOnce, Context.ReconcileDriver, ReconcileDriverOptions, alert), internal/reconcile/mismatch.go(Tracker.Blocks/EntryAllowed/Permanent/Refresh, blockFromReconcileState), internal/execgw/symbolgate.go(ReconcileReasonFor), internal/obs/event.go·notifier.go·record_only.go, internal/app/engine/alertdelivery.go, 관련 시험(internal/app/engine/reconcileloop_test.go, adoption_include_test.go, a095_*_test.go).
다음을 P0/P1/P2 로 분류해 보고하라: (1) 알림이 빠지는 구멍(침묵 경로), (2) 알림 폭주, (3) critical 미전달이 진입을 잠그거나 운영 모드를 올리는 부작용, (4) 시험 전략이 놓치는 변이, (5) 범위 누락, (6) 문서의 사실 주장 중 코드와 다른 것. 각 발견에 파일:줄 근거와, 그것이 사실 정정인지 결정·범위 변경인지를 적어라.
파일을 수정하지 말 것. 명령은 읽기 전용(cat, grep, sed -n, git show/log 등)만 쓸 것. go test·make 는 돌리지 말 것.
```

## 최종 메시지 원문

**판정: 계약 freeze 보류. P0 0건, P1 4건, P2 4건.** 구현 전 설계 리뷰 결과이며, 아래 위험이 이미 a194 코드로 구현됐다는 뜻은 아닙니다.

기준 HEAD `064f59fd99c11ea3a19adcdb804bbcdfed2470cb`와 현재 워킹트리를 대조했습니다. 아래 change 문서 경로는 `openspec/changes/a194-a-holding-left-unprotected-by-a-reconcile-block-is-reported/` 기준입니다.

1. **P1 — 알림 OFF를 normal로 내리는 실제 경로가 설계에 없다.**  
   **분류: 부작용·범위 누락 / 결정·범위 변경 필요.**

   `design.md:9–11`은 두 critical 타입을 재사용하면서 OFF에서는 normal이라고 정합니다. 하지만 `obs.Event`에는 등급 override가 없고, `Notifier.Notify`는 타입만으로 등급을 정합니다. 두 타입 모두 항상 critical입니다.

   근거: `internal/obs/notifier.go:63–75,149–157`, `internal/obs/event.go:370,386`. `Context.ReconcileDriver`는 설정값을 옵션에 넣지만 알림기는 그대로 연결합니다(`internal/app/engine/reconcileloop.go:373,381–383`).

   따라서 타입을 그대로 `alert()`에 넘기면 **OFF 엔진에도 outbox가 생깁니다.** 전송기 없음도 실패로 계산되고, 계좌 전체 진입 차단·`ENTRY_BLOCKED`로 이어집니다(`internal/app/engine/alertdelivery.go:315–325,446–471`).

   **조치:** 보유 보고와 차단 보고 각각의 OFF용 normal 타입·기록 경로를 명시해야 합니다. OFF·topic 잔존·topic 없음 각각에서 **새 outbox 0, 새 래치 0, 새 모드 전이 0**을 검증해야 합니다.

2. **P1 — D5의 하루 1회는 현재 `Notify` 배선으로 구현되지 않는다.**  
   **분류: 알림 폭주 / 사실 정정 + 배선 결정 필요.**

   `design.md:11`의 날짜 key와 재알림 창 0은 원장 API에서는 가능한 조합입니다. 그러나 드라이버의 `Alerts`는 `Notify` 인터페이스이고(`reconcileloop.go:137–139`), 실제 `Notify`는 `n.remindAfter()`를 사용합니다. 여기에 0을 설정하면 기본 **1시간**입니다(`internal/obs/notifier.go:301,395–399`).

   따라서 날짜 key만 추가하면 전달·승인된 같은 행이 **같은 날 다시 무장되어 시간마다 발송**될 수 있습니다. `RecordOnly.Notify`도 기본 창을 사용합니다(`internal/obs/record_only.go:60`).

   **조치:** 명시적으로 0을 받는 `Notifier.RecordCritical(ctx,e,0)` 등의 연결을 설계에 확정해야 합니다(`record_only.go:79–100`). 공유 Notifier의 창 변경은 다른 critical 보고에도 영향을 주므로 대체안이 아닙니다. 기존 `analysis/code-context/evidence-reconciliation.md:19`에도 남아 있는 미해결 항목입니다.

3. **P1 — D6의 현재형 경고는 해소 후 늦게 배달되면 거짓이다.**  
   **분류: 문서·알림 사실 오류 / 사실 정정.**

   `design.md:12`는 “손절·익절이 걸려 있지 않다”, “운영자 해제 전까지 앱에서 직접 관리”라고 말합니다. 차단 시작 시각만 붙여도 **어느 관측 시점의 무보호인지**는 고정되지 않습니다.

   기존 경로는 이 문제 때문에 실패 관측 시각과 과거형을 사용하고, 이미 편입됐을 수 있으니 현재 상태를 확인하라고 안내합니다(`internal/app/engine/adoption.go:481–505`). 정본도 해소 뒤 배달되어도 참일 것을 요구합니다(`openspec/specs/engine-safety/spec.md:1905–1907,2009–2011`).

   **조치:** 차단 시작 시각과 별도로 **무보호 관측 시각**을 저장하고 “그 시각에는 보호되지 않았다”로 작성해야 합니다. 자동 해제 가능한 `blocked_stale`에 운영자 해제가 필수인 것처럼 안내해서도 안 됩니다. 시험에는 **PENDING → 차단 해제·편입 성공 → 지연 배달**이 필요합니다.

4. **P1 — ADDED delta만으로 기존 무알림 시나리오를 좁혔다고 볼 수 없다.**  
   **분류: 정본 충돌·범위 누락 / 결정·범위 변경.**

   정본 `exit-policy/spec.md:159–161`은 **모든 RECONCILE 무결정 보유**를 조건으로 무알림을 요구합니다. a194 delta는 ADDED만 사용해 같은 입력에 즉시 critical을 요구합니다(`specs/exit-policy/spec.md:1,19–21`). `design.md:49`의 “기존 시나리오는 유예 안에서 여전히 참”은 기존 WHEN에 없는 조건을 추가 해석한 것입니다.

   또한 기존 등급 계약과 코드는 “실제로 편입을 시도했으나 실패”를 critical 사실로 정의합니다(`openspec/specs/engine-safety/spec.md:1885`, `adoption.go:440–452`). 차단 때문에 **시도하지 않은** 두 사실을 추가하는 것은 등급 계약의 확장입니다.

   **조치:** 기존 RECONCILE 시나리오를 MODIFIED로 좁히고, engine-safety에 새 사실·본문 의미·등급의 예외를 함께 정렬해야 합니다. 기존 미전달 처리 알고리즘을 유지하더라도 **그 알고리즘을 발동시키는 상황은 늘어납니다.** 최신 proposal의 모드 승격 설명은 확인했으며, 이를 누락 발견으로 다시 세지는 않았습니다.

5. **P2 — 성공한 대사 주기에만 보고하므로 침묵이 무기한 지속되는 경로가 남는다.**  
   **분류: 침묵·범위 누락 / 범위 결정.**

   `design.md:17–20`대로면 안정화 실패, 로컬 조회 실패, Refresh 실패, Observe 실패 모두 보고를 건너뜁니다(`reconcileloop.go:422–430,451–470`). **이미 복원된 영구 차단도** 다른 종목 때문에 대사가 계속 실패하면 일일 보고가 없습니다. 안정화 미수렴은 오류조차 아니므로 `reconcile.clean`으로 로그될 수 있습니다(`reconcileloop.go:529–533,545–555`).

   추가로 포지션 조회 실패는 차단 분기보다 앞에서 빠집니다(`adoption.go:129–134`). 주석의 “다른 경로에서 이미 보고됨”은 사용자 알림의 근거가 아닙니다. 드라이버는 ingest 알림을 명시적으로 끕니다(`reconcileloop.go:342–347`).

   **조치:** “하루 한 번·즉시”를 **성공한 대사 및 조회 가능한 포지션에 한정**할지, 기존 활성 차단 보고를 성공 조건에서 분리할지 결정해야 합니다. 후자는 범위 변경입니다. 기존 설계가 실패 주기 제외를 공개한 점은 인정하지만, 유예 시계를 보존하는 것만으로 보고의 도달 가능성이 보장되지는 않습니다.

6. **P2 — D1은 `operator_only` 전체를 판별하지 못하며, 본문에 필요한 원인·시각도 반환하지 않는다.**  
   **분류: 침묵·사실 오류 / 사실 정정 + 지원 범위 결정.**

   `blockFromReconcileState`에서 `SNAPSHOT_UNAVAILABLE`, `SNAPSHOT_STALE`은 `ReleaseOperatorOnly`지만 `ReasonReconcileMismatch`입니다(`internal/reconcile/mismatch.go:1398–1412`, `internal/execgw/symbolgate.go:157–164`). D1은 이들을 자동 해제 가능한 쪽으로 분류하여 즉시 보고와 일일 보고를 놓칩니다.

   현재 생산 쓰기 자리가 없다는 조사 결과는 **복원된 행에도 이 술어가 맞는다**는 증명이 아닙니다. `Refresh`는 활성 행을 그대로 복원합니다(`mismatch.go:1057–1065`). `EntryAllowed` 반환에는 `Reason`, `Detail`만 있고 `Cause`, `Since`는 없습니다(`mismatch.go:1326–1329`).

   **조치:** `Blocks()`의 `Release/Cause/Since`를 사용할지, 지원 원인을 제한할지 명시해야 합니다. 두 문제 모두 기존 증거에 이미 기록된 미해결 항목입니다.

7. **P2 — 시험 계획의 관측값으로는 주요 변이가 살아남을 수 있다.**  
   **분류: 시험 전략 / 검증 범위 변경.**

   `design.md:30–43`, `tasks.md:23–29`에 다음 공백이 있습니다.

   - **재알림 창 0 → 기본 1시간:** 원장은 재발송 시 기존 행을 UPDATE하므로 **행 수 1은 그대로**입니다(`internal/journal/outbox.go:334–346`). 전달 성공 후 같은 날 1시간 초과, 승인 후 반복에서도 실제 발송 수를 검사해야 합니다.
   - **critical 앞 메모리 래치 추가:** 변이 목록에는 있지만 RED 목록 1–9에는 기록 실패 주입·저장소 회복 시나리오가 없습니다. 기존 a095 시험은 기존 경로만 검증합니다(`a095_the_unmanaged_report_is_graded_by_its_fact_internal_test.go:856–875`).
   - **UTC 날짜 사용:** 단순히 하루를 진행하면 통과할 수 있습니다. KST 자정 전후와 UTC 자정을 분리해야 합니다.
   - **항상 critical로 발행:** 재사용 하네스 기본값은 `NotificationsEnabled=false`이고 `fakeAlerts`는 이벤트를 배열에 추가할 뿐입니다(`reconcileloop_test.go:154–175`, `exitloop_test.go:122–126`). 실제 Notifier·outbox·배달 실행자를 연결해야 부작용을 검증합니다.

   선택 정규식 `A194|Adoption|ReconcileLoop|Unmanaged`는 **a095 전체를 포함하지 않습니다.** 예를 들어 `TestA095NotificationsOffNeverRecordsACritical`은 빠집니다. 또 `design.md:32`의 “critical 1”은 **보유 보고 타입만 세는지** 명확히 해야 합니다. D5까지 활성인 첫 주기에는 차단 보고도 별도로 생깁니다.

8. **P2 — `ExitEligible`을 “이미 보호되어 청산이 계속 나감”으로 설명한 것은 과장이다.**  
   **분류: 문서 사실 오류 / 사실 정정.**

   `design.md:26`, delta `specs/exit-policy/spec.md:33`의 설명과 달리 `ExitEligible`은 진입 결정·편입 기록에 따른 **자격**입니다(`internal/journal/position_projection.go:515–518`). exit state가 실제 열렸음을 보증하지 않습니다.

   `adoptOne`은 편입 커밋 뒤 exit state 개설에 실패해도 편입 성공으로 처리합니다(`adoption.go:394–401`). 정본 역시 이 보호 미개설 보유를 별도 경계로 남깁니다(`openspec/specs/engine-safety/spec.md:1891–1893`).

   **조치:** 제외 범위를 “이미 exit 자격을 가진 보유”라고 정확히 적고, 보호 미개설 경계는 이번 범위 밖이라고 명시하면 됩니다. 이 리뷰가 그 후속 구현까지 요구하는 것은 아닙니다.

파일 수정·`go test`·`make`·게이트 실행은 **0회**입니다. 지정 문서 5개는 전체를 읽고, 핵심 함수·직접 호출 경로·관련 시험 본문과 정본 관련 조항을 대조했습니다. `analysis/`는 증거 조정·CodeGraph baseline·BTM을 확인했으며 AST 17개 전수 검증이나 기존 커버리지 재실행은 하지 않았습니다. 운영 사고·“133행 전부 해제”는 운영 원본을 조회하지 않아 미검증입니다.

`ai-memory-retrieval`을 읽고 CLI 조회를 시도했으나 로컬 서버 연결이 차단됐습니다. 과거 기억을 판정 근거로 사용하지 않았습니다. `impeccable: N/A — UI 변경 없는 엔진·알림 계약 리뷰`.