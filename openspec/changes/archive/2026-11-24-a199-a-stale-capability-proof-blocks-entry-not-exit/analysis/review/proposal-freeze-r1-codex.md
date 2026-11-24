**P0 경로는 확인하지 못했습니다. P1 3건, P2 3건입니다.** 구현 전 설계에 대한 정적 판정이며, 기준 HEAD는 `483d0bbcecdf6ada067e18e9372325d6c0a14ee1`입니다. 파일 변경·네트워크·tossctl·운영 설정 접근·시험 실행은 하지 않았습니다.

아래 `change/`는 `openspec/changes/a199-a-stale-capability-proof-blocks-entry-not-exit/`입니다.

[P1] 빈 endpoint 문자열과 미지원 형식이 ‘낡음만 있는 증명’으로 통과한다

- **근거·입력·순서:** `change/design.md:9,39`는 온전한 증명만 강등하고 빈 endpoint 목록은 거절한다고 정합니다. 하지만 [attest.go:218](/home/daniel/tossos-work/a199-impl/internal/attest/attest.go:218)는 배열 길이만 검사합니다. 계좌 일치·정상 만료 시각·`endpoints: [""]` 또는 `["   "]`이면 불완전 검사를 통과하고, `MissingEndpoints`가 전체 필요 endpoint를 누락으로 판정하여 `ErrEndpointNotAttested`를 반환합니다(:242–255). D1대로면 청산 전용 기동입니다.
- **두 번째 입력:** 같은 파일의 `format_version: -1`. [Load:174](/home/daniel/tossos-work/a199-impl/internal/attest/attest.go:174)는 `> 1`만 거절하므로 통과합니다. 만료나 endpoint 누락을 섞으면 역시 강등됩니다. delta `spec.md:24`의 “알 수 없는 형식은 거부”와 맞지 않습니다.
- **영향:** “이 계좌에서 한때 증명됐고 낡음만 있다”는 전제가 성립하지 않습니다. 진입 개방은 아니지만, 허용하지 않기로 한 불완전·미지원 입력으로 실거래 청산 런타임을 세웁니다.
- **제안:** 지원 형식 범위와 정규화 후 빈 endpoint의 취급을 명시하고 공용 판정에 반영하십시오. `[]`, `[""]`, `["   "]`, 음수 형식을 만료/미만료와 교차해 고정해야 합니다.
- **claim-check:** `:24`의 D1 “참”은 **sentinel 집합으로 구현 가능**하다는 뜻에서만 맞습니다. 온전함·형식 경계까지 검증됐다는 판정으로는 부족합니다.

[P1] ‘실행 중 진입 권한 유지’가 제출 직전 attestation 재검증 요구와 충돌한다

- **근거:** `change/design.md:15,57`은 도는 중 낡아진 엔진이 “진입 권한을 유지”한다고 설명합니다. delta `spec.md:29`도 동작 불변을 포괄적으로 규정합니다. 반면 [정본 engine-safety:522](/home/daniel/tossos-work/a199-impl/openspec/specs/engine-safety/spec.md:522)는 모든 안전 권한의 제출 직전 재검증을 요구하고, `:534–535`는 decision 이후 attestation digest가 바뀌면 신규 attempt를 거절하도록 명시합니다.
- **입력·순서:** 정상 증명 A로 기동 → 진입 decision 생성 → 갱신이 POST endpoint 없는 증명 B를 게시 → dispatch. D6의 권한 유지 설명과 정본의 attestation 변경 거절이 서로 다른 답을 줍니다.
- **영향:** 기동 상태를 유지한다는 약속이 주문별 권한 재검증을 생략해도 된다는 근거로 읽힙니다. **현재 생산 빌드에서 이 입력으로 매수가 제출된다는 증거는 확인하지 못했습니다.** 정본·설계의 권한 범위 충돌입니다.
- **제안:** D6를 “감시가 기동 표지·루프를 직접 전환하지 않는다”로 한정하고, activation manifest·digest·expiry의 주문별 검사는 계속 적용된다고 명시하십시오.
- **claim-check:** `:38`의 “감시 쓰기 0”은 맞지만, 그것만으로 “실행 중 진입 권한 유지”까지 정당화하지는 못합니다.

[P1] 보호가 프로세스 수명에 묶인다는 필수 경고가 조건부 작업으로 남아 있다

- **근거·입력·순서:** 낡은 증명 + 보호 `UNWIRED` + 다른 조항 정상으로 청산 전용 기동 → `Verified=false` → [logGateDecision:479](/home/daniel/tossos-work/a199-impl/internal/app/engine/interlock.go:479)의 경고 분기를 건너뜁니다. `change/analysis/edit-surface.md:20`은 이 수정을 여전히 **조건부**로 둡니다.
- **영향:** delta `spec.md:48–49`가 요구하는 “보호는 프로세스 수명에 묶여 있다”는 기동 전달이 빠집니다. “청산 전용으로 떴다”는 출력만으로 브로커에 손절 주문이 남지 않는다는 사실은 전달되지 않습니다.
- **제안:** P11을 필수 편집으로 확정하고 `Verified || ExitOnly`인 보호 미배선 기동에서 경고를 검증하십시오.
- **기존 발견 여부:** claim-check `:32`가 이미 정확히 발견했습니다. 현재 설계에서 처분이 확정되지 않은 문제입니다.

[P2] 감시 문구 수정 범위가 실제 상태 전이를 덮지 못한다

- **입력·순서 ①:** 유효한 증명이 만료 D-2 또는 감독 증거 기한 초과 상태가 됨 → `notice/imminent/proof_overdue` → [eventFor:408](/home/daniel/tossos-work/a199-impl/internal/app/engine/attestation_watch.go:408)의 “그 뒤 재시작 거부·자동 손절 없음” 문장이 남습니다(:408–421). D7은 `expired/scope_lost`만 수정 대상으로 명시합니다.
- **입력·순서 ②:** 청산 전용 기동 → 파일만 만료 30일 뒤의 정상 증명으로 교체 → 감시는 `healthy` → [prepare:345](/home/daniel/tossos-work/a199-impl/internal/app/engine/attestation_watch.go:345)에서 알림 없이 반환합니다. 프로세스의 진입 래치는 그대로인데 “현재 청산 전용”이라는 안내도 사라집니다.
- **영향:** 첫 경우는 잘못된 재시작 결과를 안내하고, 둘째는 파일 회복과 런타임 회복을 구분하지 못합니다.
- **제안:** 미래 귀결 문구까지 수정하고, `기동 상태 × 현재 파일 판정`을 별도로 정의하십시오. 파일 회복 시 재시작 필요를 한 번 알릴지 등급·키 규칙까지 결정해야 합니다.
- **기존 발견 여부:** claim-check `:40–41`의 판정에 동의합니다. 다만 edit-surface `:24`의 `prepare` 무편집 방침과 아직 조정되지 않았습니다.

[P2] ‘유효한 증명으로 재기동하면 래치 없음’은 알림·운영 모드와 함께 성립하지 않는다

- **근거·입력·순서:** 알림 ON으로 청산 전용 기동 → 만료 critical 알림 기록 → 전달 실패 지속 → [alertdelivery.go:446](/home/daniel/tossos-work/a199-impl/internal/app/engine/alertdelivery.go:446)가 알림 래치와 `ENTRY_BLOCKED`를 설정합니다(:470–471). 이후 증명을 고쳐 재기동해도 [buildGateway:287](/home/daniel/tossos-work/a199-impl/internal/app/engine/gateway.go:287)의 알림 복원과 `:291`의 운영 모드 복원이 차단을 유지합니다.
- **영향:** `change/design.md:50`의 “래치 없음(원장에 남는 것 없음)”과 proposal `:40`의 정상 복귀 절차는 이 입력에서 사실이 아닙니다. 새 attestation 사유가 사라지는 것과 실제 진입 가능은 다릅니다.
- **제안:** “**attestation 사유 래치만** 사라진다”로 한정하십시오. 다른 래치와 함께 재기동하는 시험, `alert-ack`·`mode-release`가 stale 사유를 지우지 않는 시험을 추가해야 합니다. 기존 차단을 자동으로 풀어 해결해서는 안 됩니다.
- **claim-check:** D4-b의 메모리 래치 설명 자체는 맞습니다. 이를 전체 진입 차단의 복구 보장으로 확대하면 틀립니다.

[P2] 청산 생성자 성공을 실제 청산 루프 실행 증거로 대신하고 있다

- **근거:** `change/design.md:44–45,49`의 시험은 생성자 성공·게이트웨이 매도/취소·런타임 조립 진입으로 나뉩니다. 하지만 [engineRuntime:722](/home/daniel/tossos-work/a199-impl/cmd/tossctl/engine.go:722)는 recovery를 연결하며, [Runtime.Run:294](/home/daniel/tossos-work/a199-impl/internal/app/engine/runtime.go:294)는 recovery 성공 전에는 exit와 감시를 포함한 어떤 루프도 시작하지 않습니다.
- **입력·순서:** 낡은 증명 + 정상 인터록 + 복구할 미결 상태 → Context·ExitObserver·ReconcileDriver 생성 성공 → recovery 대기 또는 오류 → exit 루프 미실행. 앞선 생성자 시험은 모두 통과할 수 있습니다.
- **영향:** delta의 “exit 관측·대사·체결 감지는 돈다”를 현재 시험 뼈대로 입증하지 못합니다. **a199가 recovery 결함을 새로 만든다는 주장은 아닙니다.**
- **제안:** 격리된 정상 recovery를 거쳐 실제 exit 한 사이클이 Guardian 발급→Gateway 매도 제출에 도달하는 연결 시험을 추가하십시오. recovery 실패·대기 시 청산 미가동이라는 기존 제한도 명시해야 합니다.
- **claim-check:** `:17` P8은 **루프 등록 집합 불변**으로 한정하면 참입니다. “실제로 돈다”까지는 미검증입니다.

**확인했지만 문제없음**

- 설계대로 Context 반환 전에 동일 `wiring.entry`에 래치를 세우면, 루프보다 늦게 차단되는 기동 창은 보이지 않습니다.
- 매수와 수량·가격 증가 정정은 `checkEntry`를 거칩니다. 매도 Place와 Cancel은 `raisesExposure=false`여서 새 래치 자체가 막지 않습니다.
- `Verified=false`를 유지하면 전략 worker는 휴면입니다. `ReplayInDoubt`는 생산 `Attested:nil`로 차단됩니다.
- 만료가 계좌 불일치를 가리는 문제는 D1a가 다룹니다. `sameAccount` 재사용 방향도 맞습니다.
- Guardian 위험 감소 발급은 매수 한도·예약 경로를 사용하지 않습니다. 공급된 보유 수량보다 많은 매도는 거절합니다. 브로커 실제 보유 한도까지 검증한 것은 아닙니다.
- 새 상태를 원장 운영 모드로 추가하지 않으므로 `EXIT_ONLY` 삭제 규칙과 직접 충돌하지 않습니다.
- 새 사유의 열거·golden·래치 순서·a098 핀은 편집 표면에 포함돼 있습니다. 콘솔 문구는 누락을 숨긴 것이 아니라 명시적으로 유예한 범위입니다.