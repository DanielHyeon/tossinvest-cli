# proposal-freeze 리뷰 1회차 — codex (원문)

- 보이스: codex-cli 0.165.0, `codex exec -s read-only -C <worktree> --ephemeral --color never -o <파일> "<프롬프트>" </dev/null`, rc 0, 2026-11-23.
- 기준: worktree `~/tossos-work/a196-impl`, HEAD `064f59fd9`(1단계 산출물 미커밋 상태 — proposal :112 정정·analysis/ 는 작업 트리에 있었음).
- 원문 = codex 의 마지막 메시지(`-o`) 그대로. 도구 호출 전체 로그(4,605줄)는 커밋하지 않음.
- 리뷰 뒤 저장소 상태: `git status` 에 change 디렉터리 밖 변경 0(읽기 전용 확인).

## 프롬프트

```text
TossOS change a196 의 proposal·design·spec delta·tasks 를 적대적 관점으로 리뷰하라. 저장소 루트는 현재 디렉터리(~/tossos-work/a196-impl)다. 파일을 수정하지 말 것(읽기 전용).

대상: openspec/changes/a196-attestation-scope-loss-is-told-as-its-consequence-and-ahead-of-time/proposal.md, design.md, specs/engine-safety/spec.md(정본 openspec/specs/engine-safety/spec.md 의 Requirement 「capability attestation 의 만료와 범위 상실은 재시작 전에 알려진다」 를 MODIFIED), tasks.md.
배경: attestation 감시(관측 전용 보조 실행자)의 알림 변경이다. 관련 코드: internal/app/engine/attestation_watch.go(judgeAttestation·binding·daysLeftCeil·prepare·stageKey·eventFor·checkOnce), internal/app/engine/a135_attestation_watch_internal_test.go, internal/attest/attest.go, internal/soak/attest.go(BuildAttestation·acceptSupervised), cmd/tossctl/soak.go(supervisedProofs·runSoakAttest·loadSoakSummary), internal/verifylive/endpoints.go(SucceededEndpoints), internal/obs/notifier.go(Notify·eventKey·RemindAfter), internal/journal/outbox.go(recordAlertTx·claimOwed), internal/obs/ntfy.go, 원 change openspec/changes/archive/2026-10-14-a135-attestation-expiry-is-announced/design.md(D2·D6).

다음을 현재 소스로 직접 확인해서 P0/P1/P2 로 보고하라:
1. 감독 증거 기한의 계산식(ProofAt + (ExpiresAt − IssuedAt))이 발급자(internal/soak 와 cmd/tossctl/soak.go)의 실제 동작과 어긋나는 경우.
2. 중복 방지 신원 변경(D4·D5)이 알림을 영구히, 또는 하루 넘게 삼키는 경우 — 기록 실패 뒤 재시도, KST 날짜 경계, 상태 전이(날수 단계 ↔ 거절 단계, 범위 상실 ↔ 만료), 엔진 재시작과 outbox 의 같은 key.
3. 정본 MODIFIED 가 같은 Requirement 의 기존 Scenario 와 모순되는 곳.
4. design §5 시험 전략이 놓치는 변이, 그리고 기존 a135 시험 중 이 change 로 뒤집히거나 통과 이유가 바뀌는 단언.
각 발견에 파일:줄 근거를 달고, 추측과 확인한 사실을 구분하라. 새 기능 제안은 하지 말고 결함만.
```

## 응답

**P0 0건, P1 2건, P2 3건. 현 계약은 수정 필요합니다.**

기준: HEAD `064f59fd99c11ea3a19adcdb804bbcdfed2470cb`의 현재 작업 트리. 파일 수정·시험 실행 없음. 아래 결과는 **소스로 확인한 사실과 그에 따른 정적 반례**이며, a196 구현 실행 결과가 아닙니다.

1. **[P1] D1의 계산값을 실제 POST 범위 상실 시각으로 단정할 수 없다.**

   [design:7–10](~/tossos-work/a196-impl/openspec/changes/a196-attestation-scope-loss-is-told-as-its-consequence-and-ahead-of-time/design.md:7)은 파일에 실린 증거의 기한이 지나면 다음 갱신에서 POST가 빠진다고 전제합니다.

   **확인 사실:** CLI는 여러 검증 기록을 읽고, 기록마다 endpoint의 최신 성공 시각을 가져옵니다. 그러나 발급자는 endpoint마다 **첫 번째 유효 증거 하나만** 파일에 싣습니다. 다른 기록의 더 최신 증거는 파일에 남지 않습니다.

   근거: [supervisedProofs:418–472](~/tossos-work/a196-impl/cmd/tossctl/soak.go:418), [SucceededEndpoints:91–103](~/tossos-work/a196-impl/internal/verifylive/endpoints.go:91), [acceptSupervised:331–340](~/tossos-work/a196-impl/internal/soak/attest.go:331).

   **반례:** 동일 endpoint에 29일 전 증거 A와 1일 전 증거 B가 있고 A가 먼저 선택되면, 파일로 계산한 기한은 하루 뒤입니다. 하루 뒤 발급에서는 A가 탈락하고 B가 선택되므로 **POST는 유지됩니다**. 새 검증을 실행하지 않아도 기한이 뒤로 이동합니다. D4의 “감독 증거의 기한은 갱신으로 바뀌지 않는다”도 성립하지 않습니다.

   기간 변수 자체는 일치합니다. CLI의 `criteria.Validity`가 증거 수집과 발급에 전달되고, 발급자는 같은 `c.Validity`로 수용 기간과 만료를 정합니다([CLI:510–525](~/tossos-work/a196-impl/cmd/tossctl/soak.go:510), [발급:242·264–265](~/tossos-work/a196-impl/internal/soak/attest.go:242)). **결함은 기간 산술보다 파일에서 사라진 대체 증거를 고려하지 못하는 데 있습니다.**

   추가로 정확한 경계는 `age >= validity`에서 제외됩니다. 따라서 [delta:7](~/tossos-work/a196-impl/openspec/changes/a196-attestation-scope-loss-is-told-as-its-consequence-and-ahead-of-time/specs/engine-safety/spec.md:7)의 “그 시각 **뒤**”는 정확히 그 시각에 실행된 갱신도 포함해야 합니다. D2의 직전·직후 시험만으로는 `>=`→`>` 변이를 잡지 못합니다.

2. **[P1] D3의 ‘최대 6시간’ 전제가 틀리고, 기존 올림 함수 재사용은 알림을 약 48시간 침묵시킨다.**

   [design:9](~/tossos-work/a196-impl/openspec/changes/a196-attestation-scope-loss-is-told-as-its-consequence-and-ahead-of-time/design.md:9)은 증거 기한 경과 후 파일이 통과하는 구간을 최대 6시간으로 한정합니다.

   **확인 사실:** 갱신이 실패하면 새 파일을 저장하지 않습니다([runSoakAttest:520–542](~/tossos-work/a196-impl/cmd/tossctl/soak.go:520)). `Verify`는 감독 증거 나이를 검사하지 않으므로 기존 파일은 자신의 만료까지 통과할 수 있습니다([Verify:211–237](~/tossos-work/a196-impl/internal/attest/attest.go:211)).

   기존 [daysLeftCeil:125–131](~/tossos-work/a196-impl/internal/app/engine/attestation_watch.go:125)은 양수 전용입니다. 음수에서도 나머지가 있으면 1을 더하므로 다음 값이 나옵니다.

   | 감독 증거 기한까지 잔여 시간 | 반환값 |
   |---|---:|
   | +23시간 30분 | 1 |
   | −30분 | 1 |
   | −23시간 30분 | 1 |
   | −24시간 30분 | 0 |

   **설계대로 기존 함수를 재사용할 때의 반례:** 매시 판정이 기한 정각을 비껴가면, 기한 23시간 30분 전에 보낸 `imminent:1`이 기한 23시간 30분 후까지 계속 재사용됩니다. D4의 기한 신원도 그대로라 [prepare의 sent 검사](~/tossos-work/a196-impl/internal/app/engine/attestation_watch.go:285)가 억제합니다. 다음 새 키는 기한 24시간 30분 후이며, **알림 사이가 48시간**입니다.

   이는 제안된 재사용 방식의 정적 반례입니다. 아직 작성되지 않은 구현이 영구 침묵한다고 단정하지는 않습니다. §5는 기한 직후만 검사하고, 갱신 실패·음수 잔여 시간·장시간 같은 키 유지 조합을 빠뜨립니다.

3. **[P2] MODIFIED 본문과 그대로 남긴 기존 Scenario가 동시에 만족될 수 없다.**

   정본의 기존 시나리오는 [engine-safety:2218–2240](~/tossos-work/a196-impl/openspec/specs/engine-safety/spec.md:2218)에 있고, delta에도 조건 보강 없이 유지됩니다.

   - **“만료 8일 전”**: [delta:16–18](~/tossos-work/a196-impl/openspec/changes/a196-attestation-scope-loss-is-told-as-its-consequence-and-ahead-of-time/specs/engine-safety/spec.md:16)은 만료가 7일 1초 이상 남으면 무조건 무알림입니다. 만료 30일·감독 증거 기한 7일인 [새 Scenario:60–62](~/tossos-work/a196-impl/openspec/changes/a196-attestation-scope-loss-is-told-as-its-consequence-and-ahead-of-time/specs/engine-safety/spec.md:60)는 같은 조건에서 알림을 요구합니다.
   - **“만료 7일 전부터 3일 전까지”**: [delta:20–22](~/tossos-work/a196-impl/openspec/changes/a196-attestation-scope-loss-is-told-as-its-consequence-and-ahead-of-time/specs/engine-safety/spec.md:20)은 일반 등급을 요구하지만, 감독 증거 기한이 2일 남았다면 본문은 critical을 요구합니다.
   - **“정상 갱신”**: [delta:36–38](~/tossos-work/a196-impl/openspec/changes/a196-attestation-scope-loss-is-told-as-its-consequence-and-ahead-of-time/specs/engine-safety/spec.md:36)은 만료만 30일 뒤로 바뀌면 무알림·단계 초기화를 요구합니다. 같은 감독 증거를 유지한 갱신에서는 D4상 신원이 유지되며, 그 사이 D-2→D-1이 되었다면 새 critical도 필요합니다.

   기존 Scenario에 감독 증거 조건이 없으므로 구현 선택으로 해소할 수 없는 계약 모순입니다.

4. **[P2] 같은 outbox key는 ‘같은 단계 한 번’의 재시작 보장을 제공하지 않는다.**

   [delta:10](~/tossos-work/a196-impl/openspec/changes/a196-attestation-scope-loss-is-told-as-its-consequence-and-ahead-of-time/specs/engine-safety/spec.md:10)은 같은 기한·같은 날수 단계의 알림을 한 번으로 제한합니다. 그러나 D4·D5는 기존 재시작 예외와 이를 조정하지 않습니다.

   **확인 사실:** 메모리 `sent`는 재시작하면 사라집니다. critical은 같은 key라도 `RemindAfter`가 지나면 outbox가 다시 활성화됩니다. 기본값은 1시간입니다([notifier:394–399](~/tossos-work/a196-impl/internal/obs/notifier.go:394), [claimOwed:382–410](~/tossos-work/a196-impl/internal/journal/outbox.go:382)). 일반 등급은 outbox 중복 검사 자체를 거치지 않습니다([Notify:149–157](~/tossos-work/a196-impl/internal/obs/notifier.go:149)).

   **반례:** 아직 유효한 D-2 파일로 알림을 보낸 뒤 2시간 후 엔진을 재시작하면, 같은 기한·같은 D-2 key가 다시 전송됩니다. D-7 일반 알림은 1시간을 기다릴 필요도 없습니다.

   원 a135는 이를 명시적으로 허용했습니다([a135 design:55–61](~/tossos-work/a196-impl/openspec/changes/archive/2026-10-14-a135-attestation-expiry-is-announced/design.md:55)). a196은 그 예외를 유지하는지 계약에 반영하지 않고 한 번을 요구합니다. §5의 가짜 notifier는 outbox의 재활성화를 구현하지 않아 이 모순을 검출하지 못합니다.

5. **[P2] §5-11의 기존 시험 영향 예상이 틀리고, 변이 범위가 핵심 반례를 제외한다.**

   [design:47](~/tossos-work/a196-impl/openspec/changes/a196-attestation-scope-loss-is-told-as-its-consequence-and-ahead-of-time/design.md:47)은 깨질 단언을 “거절 단계에서 파일 변경 시 다시 알림”으로 한정합니다. 실제 a135 픽스처는 **증거 시각 = 발급 − 1시간**이므로 새 기한은 항상 **만료 − 1시간**입니다([픽스처:36–47](~/tossos-work/a196-impl/internal/app/engine/a135_attestation_watch_internal_test.go:36)).

   | 기존 단언 | a196 적용 시 정적 결과 |
   |---|---|
   | `TestA135StageBoundaries`: 만료 `7일+1초` → healthy/D-8 | 감독 기한은 `6일23시간+1초` → notice/D-7. **뒤집힘** |
   | 같은 시험: 만료 `2일+1초` → notice/D-3 | 감독 기한은 `1일23시간+1초` → imminent/D-2. **뒤집힘** |
   | 같은 시험: 만료 `1초` → imminent/D-1 | 감독 기한은 이미 지남. 기존 음수 계산 때문에 **같은 값으로 통과하지만 이유가 틀림** |
   | `TestA135SameStageOnANewAttestationIsAnnouncedAgain`의 발급 시각만 +1분 | `ExpiresAt−IssuedAt`이 1분 줄어 감독 기한도 바뀜. **통과해도 발급 신원 초기화를 증명하지 않음** |
   | `TestA135RenewalResetsAndScopeLossIsImmediate`의 정상 갱신 | 새 픽스처가 감독 증거도 교체함. **같은 증거를 유지한 갱신의 침묵·초기화는 검증하지 않음** |

   단언 근거: [경계:73–94](~/tossos-work/a196-impl/internal/app/engine/a135_attestation_watch_internal_test.go:73), [신원:474–494](~/tossos-work/a196-impl/internal/app/engine/a135_attestation_watch_internal_test.go:474), [갱신:262–268](~/tossos-work/a196-impl/internal/app/engine/a135_attestation_watch_internal_test.go:262).

   [tasks:3·30](~/tossos-work/a196-impl/openspec/changes/a196-attestation-scope-loss-is-told-as-its-consequence-and-ahead-of-time/tasks.md:3)이 한정한 6개 변이에는 다음이 없습니다.

   - 수용 경계 `>=`→`>` 및 복수 기록의 대체 증거.
   - 음수 잔여 시간에서 기존 D-1 키 재사용.
   - KST→UTC 날짜 변경: 기존 일일 시험은 25시간을 전진하므로 둘 다 통과합니다([시험:499–513](~/tossos-work/a196-impl/internal/app/engine/a135_attestation_watch_internal_test.go:499)).
   - 날수↔거절, 범위 상실↔만료의 실제 outbox key 충돌·복귀.
   - 같은 key를 가진 새 watch와 기존 outbox를 결합한 재시작.

중복 방지 검토에서 **확인하지 못한 결함도 구분합니다.** 현재 `checkOnce`는 기록 실패 시 `sent`를 설정하지 않고 재시도합니다([248–257](~/tossos-work/a196-impl/internal/app/engine/attestation_watch.go:248)). D5대로 KST 날짜·단계·누락 집합을 모두 outbox key에 보존한다면 날짜 변경과 서로 다른 거절 단계는 별도 key입니다. 따라서 **D5 자체가 영구 누락을 만든다는 주장은 확인되지 않았습니다.** 이 부분의 문제는 시험 누락이며, 확인된 하루 초과 침묵 반례는 2번입니다.

검증 한계: 기존 인용 AST 22개의 소스 SHA 일치를 확인하고 관련 원문을 대조했습니다. AST 재생성·CodeGraph 실행·시험·운영 데이터 확인은 하지 않았습니다. ai-memory는 MCP 미제공, CLI 연결은 샌드박스 `Operation not permitted`로 조회하지 못했습니다. `impeccable: N/A — 이번 작업은 화면 변경 없는 읽기 전용 계약·소스 리뷰`.