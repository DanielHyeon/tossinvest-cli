# 남은 작업 정합성 — 2026-10-20

이 문서는 현재 소스와 보존된 read-only 증거를 대조한 상태 요약이다. 최초 상태 조사에서는 실행·측정·주문·서비스 조작·설정 변경·테스트·게이트를 수행하지 않았다. 이후의 선별 오프라인 시험은 아래 후속 검증에 별도로 기록한다. 완료 표시는 기존 OpenSpec task와 그 task가 요구한 증거에만 따른다.

## 후속 검증

- [a156 기존 순수 경로 검증](../openspec/changes/a156-gate-live-strategy-policy-rollout/analysis/offline-follow-on-verification-2026-10-20.md): 기존 평가기를 재사용하며, 미구현 범위는 실제 권위 입력 복원부터 qFinal·원자적 승인·주문 연결까지의 통합이다. 시험 이름보다 실제 assertion과 기록된 실행 범위로 판단한다.
- [a100 T2-A 준비 교차표](../openspec/changes/a100-wire-fill-to-broker-protection/analysis/t2a-readiness-crosswalk-2026-10-20.md): 기존 task 2.7이 로트표에서 빠진 부분을 정정했다. M-A·0.11 실측 선행조건은 유지한다.

## 검토 기준과 재현 범위

- source snapshot: 2026-10-20, HEAD 49a8954bf6e57171c875ecb686e7d1acc1b3e22a. 이후 계속 기록되는 ledger·state·로그·profile 파일의 상태는 달라질 수 있다.
- OpenSpec 정본: [a087 tasks](../openspec/changes/a087-a-protective-exit-is-a-market-order/tasks.md), [a100 tasks](../openspec/changes/a100-wire-fill-to-broker-protection/tasks.md), [a105 proposal](../openspec/changes/a105-protection-supervision-goes-wired/proposal.md), [a107 proposal](../openspec/changes/a107-retire-the-second-protection-core/proposal.md), [a156 design](../openspec/changes/a156-gate-live-strategy-policy-rollout/design.md).
- inspected conditional-trigger records: ~/.local/share/tossos/capability-verify.jsonl (63 rows, SHA 5e216f801abd44b87aa11c8886f47459e7acc5bd9a69fa65a0db6f8d7987fb6c), ~/.local/share/tossos/capability-verify-us.jsonl (34 rows, SHA 212b08b41c39409c6211f814b1321ed57f6d3dd108f25b44e271c371108f30fc), ~/.config/tossctl/capability-verify.jsonl (27 rows, SHA 7ab9420f2c52578b66e7f1f8ff62241fa9c3a74c937fa83a113ba668a860b218). 개인 계좌·종목·주문 식별자는 기록하지 않는다.
- a063/a074 retained-store snapshot도 2026-10-20 기준이다. 이후 서비스·파일·저장소 상태를 이 문서가 대변하지 않는다.

## 완료된 좁은 오프라인 근거

| 항목 | 근거 | 완료 범위 |
|---|---|---|
| a156 보호 회귀 | [a159 test](../internal/app/engine/a159_protection_restart_test.go), commit 5f350659 | KR synthetic partial-fill, entry block, reopen, duplicate replay, blocked new leg. broker, qFinal, deadline, US 또는 a156 완료를 뜻하지 않는다. |
| a156 stored snapshot | [reader](../internal/journal/campaign_continuation_snapshot.go#L32), commit 67810038 | campaign/position/leg/watermark를 read-only transaction으로 읽는다. FullPlanUnavailable이며 sellable, risk, qFinal, next-leg authority는 제공하지 않는다. |
| a156 detached inspection | a161 artifact, commit cd5f272 | capture-time plan/risk inspection only. current live ledger, weekly reservation, authority restoration, atomic capture 증명이 아니다. |
| a100 T1 | commit 30694fe, retained review receipt | 승인된 T1 test scope만 완료. D3 REPLACE pending partial-fill defect는 T2-A 전에 남아 있다. |
| a087 Phase 1 | [sell intent](../internal/app/engine/exitloop.go#L1788)와 retained pin | 실제 하한가 제출 모집단 0을 N/A로 처분하고 관측가 guard를 pin한 완료분이다. 0.5 session census는 별도 준비 증거이며 D7 처분은 미결이다. MARKET Phase 2 구현이 아니다. |

## 아직 구현되지 않은 오프라인 계약

| change | 미구현 범위 | 현재 소스 근거 |
|---|---|---|
| a156 | actual fill·held reservation·sellable·fresh risk·nonretreating stop으로 next-leg qFinal을 계산하고 atomic admission/lease/dispatch로 결속하는 경로 | [PlanCampaignLeg](../internal/journal/position_campaign.go#L522)는 planned leg만 저장한다. 현재 production caller가 발견되지 않았다. first-leg만 [RecordQFinalCampaignFirstLeg](../internal/journal/strategy_first_leg_atomic.go#L103)와 [gateway](../internal/execgw/riskguardian_first_leg.go#L240)로 연결된다. |
| a156 | intraday campaign-only deadline/auction closeout, authoritative residual/fill flat proof, weekly/adopted isolation | generic [flatten saga](../internal/journal/flatten.go)는 있으나 a156 deadline contract에 결속한 경로·receipt는 확인되지 않았다. |
| a100 | M-A raw-status contract, T2-A D3/child owner repair, 이후 worker assembly | [gateway adapter](../internal/protectionofficial/gateway.go#L259)는 status/child ID를 축약한다. [engine wiring](../internal/app/engine/protection_wiring.go#L39)은 KR/US 모두 Wired:false다. |
| a087 | Phase 2 MARKET support, D7 session policy disposition, 5.2 session-boundary proof, 5.3 canonical measurement update | [shape guard](../internal/execgw/failclosed.go#L53)와 [trading support](../internal/trading/service.go#L269)는 현재 비분수 MARKET을 허용하지 않는다. |
| a105 | a100 착지 뒤 flat-window, protection supervisor, market-specific Wired production, manifest/signature/latch consumption | a100가 선행이다. Wired:false를 변경하는 것으로 대체할 수 없다. |
| a107 | a100 이후 최신 consumer census를 바탕으로 dormant/security core를 범위 한정 제거 | a105와 기술적 직접 SHALL 의존성은 확인되지 않았다. 다만 a105 → a107은 최신 사용자 요청의 실행 순서이며, 둘 다 a100 후 shared seal surface ownership 조율이 필요하다. |

a156은 추가 fixture만으로 tasks 2.2/2.3을 완료로 바꾸면 안 된다. Manager가 lane별 immutable plan/version/digest, actual-state authority, qFinal admission binding, stop monotonicity, intraday deadline/flat proof를 먼저 확정해야 한다.

## 실측 또는 사람 승인이 필요한 항목

| change | 필요한 승인·증거 | 현재 한계 |
|---|---|---|
| a156 | target market/symbol/lane, policy/version/digest, loss/exposure/window, rollback과 사람 승인; live later-leg, closeout, recovery, authoritative residual reconciliation | task 4.1·4.2 open. offline result는 live-ready 또는 archive 근거가 아니다. |
| verify-execution-capability | conditional actual trigger, child identity/latency, expiry/partial/OCO/concurrent-sell 등 선택한 속성의 evidence | 검사한 세 JSONL의 conditional-trigger 항목은 deferred/observed=false, child exposure=unverified다. 구현된 tool은 실행 증거가 아니다. |
| verify-observes-the-trigger | explicit symbol exclusion/config decision, polling budget coordination, plan inspection, supervised single run, residue reconciliation | trigger tool과 regression은 있어도 live section 3은 open이다. 자동 start/stop/cancel 권한은 없다. |
| a063 | approved install/reload, exact profile/template digest/window, 3 consecutive qualifying survey days, failed-renewal observability | installed unit은 expected renewal-status flag와 다르고 shell failure masking이 남는다. current attestation·historical 21-day receipt는 approved window를 대체하지 않는다. |
| a074 | deployed-cycle quarantine-creation event | inspected config/legacy store와 selected logs에서 exact event 0. 이 결과는 조회 범위 내 0이며 rotated/system journal·외부 archive 전체 부재 주장이 아니다. |

## a087 exact-ID 귀속 한계

기존 a087 5.1 KR MARKET 관측은 completed screen에 order ID가 없어 시각·방향·수량으로 간접 연결됐다고 스스로 기록한다. 수용 계약은 order-ID 상태 확인을 요구한다. 기존 checked evidence를 되돌리지 않지만, 이 차이는 exact-ID proof로 승격하지 않는다. D7 사람 처분, 5.2, US condition은 계속 open이다.

## 순서와 다음 결정

1. a100 M-A의 raw-status/measurement precondition과 T2-A D3 contract를 사람 승인으로 동결한다.
2. a105 → a107 순서를 최신 사용자 요청의 실행 순서로 유지한다. 이는 기술적 직접 SHALL 의존성 주장이 아니다. 둘 다 a100 착지 뒤에만 시작하며, shared dormant/security seal surface의 단일 writer를 정한다.
3. a156은 별도 offline change로 분할할 수 있으나, plan/risk/qFinal/deadline 계약을 먼저 확정한다.
4. verify changes, a063, a074는 표의 운영 증거가 승인되기 전에는 task completion·archive·policy promotion을 주장하지 않는다.

승인 필요 목록은 위 표의 live trigger, service install/reload, attestation survey window, deployed quarantine observation, a156 live rollout이다. 이 문서는 승인 요청이나 실행 명령을 포함하지 않는다.

적용성: research-first는 기존 report·source·receipt를 재사용했다. grill-me는 기존 승인 계약을 바꾸지 않아 N/A다. impeccable은 UI 변경이 없어 N/A다.

## 운영 승인 전 준비 부록 — 실행하지 않은 정적 대조

기존 [a063 migration plan](../openspec/changes/a063-align-attestation-renewal-profile/design.md), [발동 관측 설계](../openspec/changes/verify-observes-the-trigger/design.md), [a074 설계](../openspec/changes/a074-critical-events-reach-the-operator/design.md)를 재사용한다. 아래는 미정 승인 필드와 현재 파일 대조이며, 새 운영 절차의 승인·실행 또는 구현 완료를 뜻하지 않는다.

### a063 unit과 바이너리

| 대상 | 2026-10-20 SHA-256 | 정적 확인 |
|---|---|---|
| [repository template](../deploy/systemd/tossos-attest.service) | `0bf50972660fa14ece6953b1d43aa51fe2df407a51783c7ea0c6decb3100de96` | direct ExecStart, `--config-dir %h/.config/tossctl`, `--record-renewal-status`, oneshot |
| `~/.config/systemd/user/tossos-attest.service` | `1524b07ae9765b883c232fe855aead90ef6320286dfce22a8b67a66a88fc055b` | 같은 명시적 config root이나 shell wrapper의 `|| printf`가 실패 종료값을 가린다. renewal-status flag 없음 |
| `~/.local/bin/tossctl` | `5287fe3cb0bda6ed91ef2aa5c55df7e8deab84e07e7f644cb280acbd60873807` | Go build metadata: revision `488b370fddb4903568aef70615fa0fb953a71ba1`, modified=true, go1.26.5 |

설치 바이너리를 실행하지 않고 build metadata와 flag 문자열만 읽었다. 해당 revision의 소스와 [현재 flag 등록](../cmd/tossctl/soak.go#L235)에는 renewal-status가 있으나, dirty 바이너리를 현재 소스에 정확히 결속하는 build receipt는 확인하지 못했다. **문자열 존재는 flag 지원·정상 동작 증거가 아니다.** 설치 전 운영자가 확인할 조건은 선택 바이너리의 지원 근거, template digest, 동일 profile의 record/output/interlock 경로, backup·rollback 대상, install/reload 및 필요한 바이너리 교체의 명시 승인, 실제 연속 3 qualifying survey dates다. 현재 [status 기록 분기](../cmd/tossctl/soak.go#L478)는 이 flag와 `--record`/`--out` override를 함께 쓰면 거부한다. 임의 경로 override로 profile 계약을 대체하지 않는다. unit 설치·reload·attest·survey 실행은 하지 않았다.

### verify 실행계획에서 아직 확정할 필드

과거 종목·보유량·노출 승인과 당시 실행 바이너리를 현재 실행계획으로 재사용하지 않는다. 확정할 항목: 현재 시장/대상/보유·수량, 손실·노출 상한, 관측 시간창, 선택한 측정 속성, 기존 기록과 resume 대상, 운영자 입회, polling 예산 조율 및 정지 권한, 중단·잔여물 처분, 영수증 보관 위치와 바이너리 provenance다.

현재 [CLI 계약](../cmd/tossctl/verify.go#L199)의 `--include-trigger`는 M0의 `--confirm-each`, `--resume`, `--redo conditional-trigger` 한정과 기존 owner-only 0700 `--trigger-receipt-dir`를 요구한다. 일반 verify의 배치 승인을 이 발동 경로에 대신 적용하지 않는다. TTL edge는 별도 두 번째 라이브 주문을 만들므로 별도 opt-in이며 발동 검증과 묶어 승인된 것으로 보지 않는다. 현재 소스와 운영 바이너리의 M0 일치 여부도 아직 확정하지 않았다. 이 부록에는 라이브 실행 명령을 만들지 않는다.

### a074 보존 이벤트 추가 검색과 재사용 절차

최근 7일 창 `2026-10-13T09:48:27.274441Z`–`2026-10-20T09:48:27.274441Z`에서 로컬 user/system journal의 exact event 문자열을 읽기 전용 검색했다. 각각 20초 제한시간을 넘겨 **불완전**이며, 0건 또는 권한 부족으로 판정하지 않는다. 권한 승격·새 이벤트 발행·서비스 조작 없이 종료했다. 조사한 config-profile `engine.log*`와 legacy `audit.log*` glob에는 각각 현재 파일만 있었고 별도 rotated 파일은 발견되지 않았다. 다른 경로·보관소 전체 부재는 뜻하지 않는다.

후속 오프라인 대조는 기존 방식으로 충분하다: profile별 journal.db를 read-only/query_only transaction으로 열어 `alert_outbox.event_type=exit.snapshot_quarantined`의 관측창별 count·created_at·state·attempts·delivered_at만 집계하고, 접근 가능한 동일 창 로그와 대조한다. 계좌·종목·event payload는 출력하지 않는다. outbox 존재는 이벤트 생성 근거, delivered 상태는 별도 전달 근거이며 서로 대체하지 않는다. [생성 이벤트 함수](../internal/app/engine/exit_quarantine_announce.go#L50)와 [기존 격리 재판정 로그](../internal/app/engine/exitloop.go#L665)도 구분한다. 앞 절의 검사 범위 내 0건은 유지하되 이번 journal timeout으로 검색 범위를 넓혀 완료했다고 주장하지 않는다. a074 11.1은 열린 상태다.
