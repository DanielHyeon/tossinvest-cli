**P0 경로는 확인하지 못했습니다. P1 2건, P2 2건입니다.** 기준 HEAD `1d472656260a34acf7f93ef3002e7883a67936f9`. 구현 전 설계와 현재 코드의 정적 대조이며, 파일 변경·네트워크·`tossctl`·운영 설정 접근·시험 실행은 하지 않았습니다.

아래 `change/`는 `openspec/changes/a199-a-stale-capability-proof-blocks-entry-not-exit/`입니다.

**[P1] 회복 알림의 일반 등급 조건이 imminent·proof_overdue의 critical 의무와 겹친다**

- **근거:** `change/design.md:16,54`, delta `spec.md:138,221–223`은 청산 전용 기동 후 파일이 판정을 통과하면 일반 등급 회복 알림을 요구합니다. 하지만 delta `:133,136,153–155,201–203`은 같은 파일의 `imminent`·`proof_overdue`에서 critical을 요구합니다.
- **입력·순서:** 청산 전용 기동 → endpoint를 복구하되 기한이 D-1인 파일 게시. 또는 아직 만료되지 않았지만 감독 증거 기한은 지난 파일 게시. 두 파일 모두 `Verify`를 통과합니다. [judgeAttestation:124](/home/daniel/tossos-work/a199-impl/internal/app/engine/attestation_watch.go:124)는 그 **뒤에** `imminent`·`proof_overdue`로 분류합니다.
- **영향:** 회복 분기가 기존 이벤트를 일반으로 덮으면 critical 의무가 사라집니다. 기존 분기를 유지하면 결정 7의 일반 회복 알림이 빠집니다. 두 알림을 내는 해석도 가능하지만, 현재 `prepare`는 이벤트 하나를 반환하고 파일 단계 키 하나로 보냄 여부를 관리합니다([:348](/home/daniel/tossos-work/a199-impl/internal/app/engine/attestation_watch.go:348)). 어느 의미를 보존할지 설계가 정하지 않았습니다. 이는 전달 실패 시 outbox·진입 래치·운영 모드 승격 여부까지 바꿉니다.
- **제안:** 파일 위험 경보와 현재 기동 상태 안내의 병존·우선순위를 명시하십시오. 기존 `imminent`·`proof_overdue` critical은 보존하고, 회복 안내가 이를 일반으로 낮추지 않도록 해야 합니다. 알림 OFF의 최종 일반 등급 변환도 양쪽에 적용해야 합니다.
- **설계 결정을 바꿔야 하나: 예.** 결정 7의 “판정 통과 → 일반”과 기존 경보 규칙의 결합을 확정해야 합니다.

**[P1] 결정 3의 가드를 ‘낡은 파일에만’ 적용한다는 B2가 새 거절 계약을 우회한다**

- **근거:** `change/analysis/claim-check.md:52`는 가드를 낡은 파일에만 적용하고 **`Verify` 통과 파일은 변경하지 않는다**고 합니다. 반면 `change/design.md:40`은 빈 항목을 만료·미만료 모두 거절하고, delta `spec.md:25`는 빈 endpoint·무효 형식을 기동 거절로 규정합니다.
- **입력·순서:** 계좌 일치·미래 만료·필요 endpoint 전부 포함인 파일에 `""` 항목 하나 추가. 별도 사례로 같은 정상 내용의 `format_version: 0` 또는 `-1`. [Load:174](/home/daniel/tossos-work/a199-impl/internal/attest/attest.go:174)는 형식 상한만 검사하고, [Verify:211](/home/daniel/tossos-work/a199-impl/internal/attest/attest.go:211)는 형식 번호와 개별 빈 항목을 검사하지 않습니다. 따라서 모두 `Verify == nil`입니다.
- **영향:** B2 순서대로 구현하면 결정 3이 거절하기로 한 입력이 **정상 기동**으로 빠집니다. 청산 전용 집합만 좁혀서는 닫힌 판정 계약이 성립하지 않습니다. 실제 진입 개방까지 입증한 것은 아니므로 P0로 세지 않았습니다.
- **제안:** 공용 판정의 형식 하한·빈 항목 검사를 정상 성공 반환보다 앞에 두십시오. B2의 “정상 입력”을 “발급자가 생성한 정상 파일”로 한정하고, `필요 endpoint 전부 + 빈 항목`, `형식 0/-1 + 미만료`도 수용 기준에 넣어야 합니다.
- **설계 결정을 바꿔야 하나: 아니오.** 결정 3과 delta는 이미 거절을 요구합니다. B2의 적용 범위 설명을 바로잡으면 됩니다.

**[P2] 기존 중복 제거 상태를 재사용하면 healthy 회복 알림이 같은 날 반복된다**

- **근거:** `change/analysis/edit-surface.md:21–22`는 기존 키 규칙 재사용을 계획합니다. 그러나 [prepare:327](/home/daniel/tossos-work/a199-impl/internal/app/engine/attestation_watch.go:327)는 `Deadline`이 바뀔 때 `sent` 전체를 비웁니다. `healthy`의 기본 키는 날짜·에피소드로 만들어집니다([stageKey:358](/home/daniel/tossos-work/a199-impl/internal/app/engine/attestation_watch.go:358)).
- **입력·순서:** 청산 전용 기동 → D-8 이상 파일로 회복 → 일반 회복 알림 전송 → 같은 날 새 증거 또는 갱신으로 기한만 이동, 여전히 `healthy` → `sent` 초기화 → 동일 날짜의 회복 알림 재전송.
- **영향:** 결정 7의 “하루 한 번”이 깨집니다. 일반 이벤트는 [Notifier.Notify:153](/home/daniel/tossos-work/a199-impl/internal/obs/notifier.go:153)에서 바로 best-effort 전송하므로, 같은 키를 써도 critical outbox가 중복을 막아주지 않습니다. **현재 구조를 유지하는 설계에 대한 정적 예측이며, 아직 구현된 재현 결과는 아닙니다.**
- **제안:** 기동 상태 안내의 일별 전송 기록을 파일 기한 변경으로 초기화하지 않도록 정의하십시오. 기존 a196 경보의 기한·재발 에피소드 규칙은 유지해야 합니다. 같은 날 여러 갱신, 날짜 변경, 회복 후 재발, 재시작을 구분한 시험이 필요합니다.
- **설계 결정을 바꿔야 하나: 아니오.** 하루 한 번이라는 결정은 유지하고, 중복 제거 설계를 보완하면 됩니다.

**[P2] 새 회복 알림이 결정 8에서 제거한 ‘재시작하면 정상’ 보장을 다시 만든다**

- **근거:** `change/design.md:16,54`와 delta `spec.md:223`은 “재시작하면 정상으로 뜬다”고 알립니다. 그러나 같은 design `:62`와 delta `:120`은 다른 차단이 재시작 뒤에도 남는다고 명시합니다.
- **입력·순서:** 청산 전용 기동 → stale critical 전달 실패 → [alertdelivery.go:448](/home/daniel/tossos-work/a199-impl/internal/app/engine/alertdelivery.go:448)의 알림 래치 및 [:470](/home/daniel/tossos-work/a199-impl/internal/app/engine/alertdelivery.go:470)의 `ENTRY_BLOCKED` 승격 → 파일 회복 → 새 일반 알림이 정상 복귀 안내 → 재시작. [buildGateway:287](/home/daniel/tossos-work/a199-impl/internal/app/engine/gateway.go:287)와 `:291`은 기존 차단을 복원합니다.
- **영향:** 결정 8은 재기동 수용 기준을 attestation 사유로 한정했지만, 새 운영자 알림은 다시 전체 정상화를 약속합니다. “정상”을 attestation 기동 분류만 뜻한다고 해석할 수는 있으나, 사용자에게 전달되는 문장에는 그 한정이 없습니다.
- **제안:** “증명은 회복됐다. 재시작하면 **attestation 사유의 진입 차단**이 해소된다. 다른 차단은 별도 확인이 필요하다”로 한정하십시오. 기존 래치를 자동 해제해서 해결하면 안 됩니다.
- **설계 결정을 바꿔야 하나: 아니오.** 결정 8의 한정을 결정 7의 알림에도 적용하면 됩니다.

단계 × 기동 표지 대조 결과는 다음과 같습니다. 등급은 알림 ON 기준입니다.

| 파일 단계 | 정상 기동 표지 | 청산 전용 기동 표지 |
|---|---|---|
| healthy | 기존 무알림 유지 | 일반 회복 안내. 일별 중복 제거 보완 필요 |
| notice | 기존 일반 기한 경보 | 회복 안내와 기한 경보의 문장·키 결합 미정 |
| imminent | 기존 critical 기한 경보 | 일반 회복 안내와 critical 요구 충돌 |
| expired | 재시작 시 청산 전용 안내 | 현재 청산 전용 사실 추가 가능 |
| scope_lost | 재시작 시 청산 전용·누락 endpoint 안내 | 현재 청산 전용 사실 추가 가능 |
| account_mismatch | 재시작 거절 안내 | 재시작 거절과 현재 청산 전용을 함께 말해야 함 |
| unreadable | 재시작 거절 안내 | 위와 같음 |
| incomplete | 재시작 거절 안내 | 위와 같음 |
| proof_overdue | 기존 critical·다음 갱신 이후 귀결 안내 | 파일은 통과하므로 일반 회복 규칙과 충돌 |

마지막 세 거절 단계에서 design D7(4)의 “문장 무변”은 **재시작 결과 문장에만** 적용해야 합니다. 현재 상태 안내까지 생략하면 delta `:138`의 기동 표지 조건을 충족하지 못합니다.

**확인했지만 문제없음**

- B2의 **발급자 인용 자체는 맞습니다.** `BuildAttestation:234–239`는 빈 GET 항목을 거절하고, 감독 endpoint는 닫힌 집합에서 받으며, 형식 번호는 `1`로 씁니다. 문제는 B2 다음 문장의 가드 적용 범위입니다.
- 만료 정각은 `!now.Before(ExpiresAt)`로 만료입니다. 같은 숫자열의 하이픈 표기는 허용되고, 마스킹된 계좌를 임의로 같은 계좌로 인정하지 않는 D1a도 일치합니다.
- 알림 OFF에서 critical을 일반으로 바꾸는 기존 `eventFor:385` 규칙과 일반 전송 실패가 승격되지 않는 경로는 확인했습니다. 결정 7이 이 경로를 보존하면 새 충돌은 없습니다.
- 새 기동 표지를 읽는 행위 자체는 래치·운영 모드를 변경하지 않습니다. 기존 critical 전달 실패의 간접 차단은 별개이며, 새 attestation 래치도 `raisesExposure=false`인 청산을 직접 막지 않습니다.

메모리 조회·그래프 재생성·새 AST 산출·실행 검증은 수행하지 않았습니다. 위 판정은 지정 문서와 관련 원본 코드에 한정합니다.