# verify-execution-capability 종료 전 감사 — 2026-10-21

## 범위와 한계

이 문서는 로컬 소스·기존 기록·설치 unit을 읽기 전용으로 대조한 감사다. 계좌 조회, 주문, 서비스 기동·재시작, 설정 변경, 원격 engine/docker 저장소 조사는 수행하지 않았다. 원격 engine/docker 저장소와 활성 engine에 대해서는 이 감사 범위가 결론을 낼 수 없으며, 이는 부재의 증거가 아니다.

현재 설치 바이너리와 현재 소스의 정확한 동일성 build receipt도 확인하지 않았다. 설치 바이너리의 문자열·build metadata 또는 로컬 기록은 현재 소스 동작과 배포된 engine 동작의 동일성 증거가 아니다.

## attestation 및 soak profile

설치된 `~/.config/systemd/user/tossos-attest.service`의 `ExecStart`는 `--config-dir %h/.config/tossctl`을 사용한다. 이 profile의 `engine.automation_gate.attestation_file` override는 비어 있어, 소스의 경로 해석(`cmd/tossctl/soak.go:639`, `internal/config/engine.go:185-192`)에 따라 실제 대상은 `~/.config/tossctl/capability-attestation.json`이다. 예전 기본 경로 또는 `attestation.json`의 존재·부재는 이 설치 profile 판정에 사용하지 않는다.

| 로컬 읽기 결과 | 값 | 해석 한계 |
| --- | --- | --- |
| attestation artifact | SHA-256 `4bcd23619659a90eabf5311e7bbcfa4ed8a722f9c0a8ea356e9239da44501ad5`, 1,392 bytes | `issued_at=2026-10-20T17:25:44.663832726Z`, `expires_at=2026-11-19T17:25:44.663832726Z`의 순서·미래 여부, endpoint 11개, `supervised_by` 2개, rate-limit 필드만 구조적으로 확인했다. 현재 account 결속·engine interlock 통과는 확인하지 않았다. |
| 설치 profile soak record | `~/.config/tossctl/capability-soak.jsonl`, SHA-256 `1edbe5878199b1e0ff7bf2a6129d0c688e2167f5a8cfb4af814df47315a4e634` | 5,858개 local records. 10/18 90개·refresh 2회, 10/19 92개·refresh 4회, 10/20 68개·refresh 1회와 endpoint/completeness OK가 기록되어 있다. 이는 사람의 현재 criteria review 입력일 뿐 자동 승인·현재 engine 증명은 아니다. |
| 설치 unit | `~/.config/systemd/user/tossos-attest.service` | shell wrapper가 오류 종료를 가리고 `--record-renewal-status`를 전달하지 않는다. unit 설치·reload·attest 실행은 하지 않았다. |

`internal/soak/attest.go:63-192`의 기본 criteria는 연속 3일, endpoint 성공, 48시간 record freshness, token observation/refresh 및 completeness를 fail-closed로 평가한다. 설치 binary의 local-only `soak status`는 이 record를 위 criteria로 다시 평가해 `READY`를 반환했다. 다만 해당 binary의 현재 source binding, attestation 발급 시 resolved account 검증, 실제 engine interlock은 별도 증거가 필요하다.

## 미완료 task 대조

| Task | 기존 근거 | 현재 판정 |
| --- | --- | --- |
| 1.2 | 설치 profile record의 최근 3일 aggregate | `--config-dir ~/.config/tossctl soak status`는 local-only 읽기에서 exit 0 및 `READY`를 반환했다(credential streak 5/3, newest cycle 2026-10-20T18:05:27Z). 1.2가 요구한 연속 3일 무인 갱신 실행·기록의 local criteria는 충족한다. 실행 binary와 현재 source의 exact build binding, engine 경로/interlock은 1.2의 추가 조건이 아니라 1.4·3.1의 별도 한계다. |
| 1.3 | 기존 M4/M8/M10 기록과 record rate-limit fields | retry matrix·polling SLO 수치 결정은 사람의 측정 해석·승인이 필요하다. |
| 1.4 | 위 artifact 구조, `internal/attest/attest.go:76-114`, `internal/app/engine/interlock.go:1-34` | artifact 존재·형식만 확인했다. resolved account 검증과 실제 engine interlock 기동 검증은 하지 않았다. |
| 2.1 | `measurements.md` M31 status 관측 | `CANCEL_REJECTED`/`REPLACE_REJECTED` 별도 record 형태와 원주문 link는 검토 범위 내 미확인이다. |
| 2.2 | M32 최소 place/cancel/amend 일부 기록 | 부분 매도 수락과 보유 초과 거절은 M30·M33에서 관측됐다. 전량 매도만 미검증이다. |
| 2.3 | 없음 | flatten dry-run 실행·기록 미확인이다. |
| 2.4 | 없음 | 약관·자동화·계정 정지 정책의 승인 기록 미확인이다. |
| 2.5 | 세 local JSONL의 conditional-trigger 항목은 모두 `deferred`, `observed=false`; triggered child 노출은 `unverified` | M13·M34에서 조건주문은 매도가능수량을 예약하지 않고, M29·M33에서 미체결 일반 매도 지정가는 예약한다는 대조가 관측됐다. 발동, child ID 지연, 만료, OCO, 부분체결 및 일반 매도 동시성은 미검증이다. |
| 2.6 | 2.5 전체 ProtectiveCapability 실측 결과 미충족 | 미검증 속성 목록은 아래와 같이 정리할 수 있지만, 보호 유형·허용 시장의 운영 정책 확정은 아직 할 수 없다. |
| 2.7 | M14/M35 replay/conflict·latency 관측 | M14·M35에서 일반·조건주문 동일-key 재생과 conflict, 왕복 지연은 관측됐다. TTL boundary와 key account scope는 미검증이다. 사용자가 TTL edge를 거부하면 task 계약상 이를 명시적으로 미검증으로 남기고 replay를 TTL/2 보수 마진으로 제한하거나 비활성 유지할 수 있다. TTL edge는 conditional trigger의 선행조건이 아니다. |
| 2.8 | M33 resting sell/conditional 관련 일부 기록 | resting sell 예약과 conditional 미예약은 M13·M29·M33·M34에서 관측됐다. collateral·unsettled을 포함한 전체 semantics는 미검증이다. |
| 2.9 | M21: 검증 주문 미체결, `costs.orders_filled=0`, `costs.collected=false` | realised commission/tax 표를 확정할 근거는 없다. |
| 3.1 | `tossctl verify report`는 local record만 읽고 미측정 속성을 `unverified`로 유지한다 (`cmd/tossctl/verify.go:242-268`, `internal/verifylive/report.go:75-165`) | 유효 attestation의 현재 review, 미검증 자동 경로 금지 목록의 확정·문서화가 남았다. |
| 3.2 | 없음 | task가 명시한 `make gate CHANGE=verify-execution-capability`는 전체 범위 실행이다. 이번 감사에서는 사용자가 전체 gate를 승인하지 않았으므로 실행하지 않았다. |

기존 `openspec/changes/verify-execution-capability/measurements.md:153-158,181-206`의 부분 관측과 이 문서는 실계좌 능력·수익성·자동 주문 권한을 승급하지 않는다.

## 후속 판단에 필요한 승인 경계

현재 조사 범위의 미검증 속성은 조건주문 발동·child 연결/지연, 만료,
OCO sibling 취소, 부분체결 잔량, 일반 매도와의 동시성, 전량 매도,
담보·미결제 수량, 멱등키 계좌 scope/TTL 및 실제 비용이다. 이 목록은
해당 속성에 의존하는 자동 경로를 허용할 근거가 없다는 정적 정리다.
실제 설정이나 자동 경로를 변경하지 않았다. TTL 실측 여부는 발동
검증을 막는 선행조건으로 다시 사용하지 않는다.

Manager는 1.2의 기록 충족만 완료로 반영했다. 나머지 열린 task와
기존 측정 원문은 유지한다. 이 change는 전체 완료·아카이브 상태가 아니다.

현재 미해결 항목은 운영자가 선택·승인할 시장, 대상/수량, 손실·노출 상한, 관측 창, 정지 권한, 기록 보관 위치, 선택 binary provenance 및 account/engine interlock 확인이다. 이 문서는 승인 요청이나 실행 명령을 포함하지 않는다.

적용성: research-first는 기존 report·source·receipt를 재사용했다. grill-me는 기존 승인 계약을 바꾸지 않아 N/A다. impeccable은 UI 변경이 없어 N/A다.
