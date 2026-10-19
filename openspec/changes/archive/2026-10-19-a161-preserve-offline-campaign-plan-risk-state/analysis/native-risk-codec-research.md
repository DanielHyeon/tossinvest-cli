# 오프라인 plan + native risk inspection codec 조사

기준 `67810038f21dffd519a3d44edb9e255238b5b74a`; 소스 읽기만 수행, 테스트/소스/명세/운영 변경 없음. 사용자 범위는 초기 계획·예산 및 현재 native RiskState의 filled/held/latches/fills/cancels를 함께 보존하되 복원 결과는 detached inspection 자료이며 실행 권한 아님.

## 실제 native 표현과 경계

- continuation `internal/continuationlane/risk.go:15-34`는 plan digest/budget, filled/held, latches, fill의 applied/known/risk/fingerprint, cancel의 applied/fingerprint를 보유. `:270-272` scope 검증은 plan digest/budget 일치뿐으로 전체 map/금액의 진위나 생산성 검사 아님. exported struct 및 map은 caller가 직접 조작 가능. `:135-212`는 unknown fill 시 latch를 설정하고 기존 숫자를 남기며, 초과도 실제 수치를 유지한다.
- reversal `internal/reversallane/types.go:302-315`는 plan digest, filled/held, latches, fills(applied/risk/fingerprint), cancel ID→fingerprint를 보유. `risk.go:86-118,188-243`의 검사들은 연산별 부분 검사이지 공개된 완전한 RiskState 진위 검증기가 아니다. unknown 시 빈 risk 값과 unknown latch가 남을 수 있다. `evaluate.go:188-194`의 lineage riskBudgetDigest는 plan digest와 예산 등의 **별도 결합 해시**다.
- weekly `internal/weeklyvaluelane/risk.go:14-25,246-274`는 private plan digest/filled/held/latches/fill fingerprint·Known과 integrity seal 보유. public getter는 fill count만 공개하며 fill 기록을 내보내지 않는다. **RiskState에 cancels 필드는 없다**. release와 중복영수증은 별도 sealed ReservationState (`reservation.go:96-100,132-150,228-247`)에 존재한다. weekly RiskState export에서 cancels는 `not_applicable`로 표시하고 별도 ReservationState/취소 이력이 복원되었다고 주장하지 않는다.
- reversal plan (`reversallane/plan.go:28-32`), weekly plan (`weeklyvaluelane/plan.go:30-42`), weekly risk는 private data를 가지므로 외부 generic codec만으로 전체 입력을 얻을 수 없다. 최소한 lane 소유 package 내 단방향 detached export가 필요하며 Go production 소스가 추가되더라도 실행 경로가 호출하지 않게 import/wiring guard가 필요하다. 검증 helper 재사용은 내부 범위에서 수행하고 decoder는 어떤 native plan/risk/FX type도 반환하지 않는다.
- continuation `validatePlan` 및 reversal `plan.valid`는 plan 구조·digest 검사, weekly `validRiskState`는 seal·파싱 검사다. 이 검사는 데이터 출처 인증, 현재 원장 반영, RiskState transition 이력의 완전성을 증명하지 않는다. continuation/reversal state는 검증 후에도 caller-owned 변조 가능. export는 deep-copy해 시점 값만 캡처하고 concurrent mutation에 대해 호출자 동기화가 필요하다; concurrent access가 안전하다는 주장 금지.

## 제안 contract

1. lane별 export는 검증 가능한 native plan과 현재 risk snapshot을 입력받아 immutable detached DTO를 생성. plan binding/자료형/범위/map/key/known 상태를 검사하고 불완전하거나 모순된 입력은 명시적 unavailable/error. private sealed FX 값은 관찰·표시용 원문/provenance로만 취급하고 어떠한 import/mint 경로도 만들지 않는다. 무결성 seal 검사와 source 인증을 구별한다.
2. known zero와 unknown을 구분한다. 금액 문자열이 `0`이라고 unknown latch가 지워지지 않는다. 일부 이벤트의 미확인/충돌/불완전 결산에서 stored filled/held는 **snapshot 숫자**이나 실제 current risk가 known이라고 결론내리지 않는다. `stored_amount`와 `risk_known`/reason 분리, 숫자를 지우거나 0으로 바꾸지 않는다. reversal은 명시적 Known 필드가 없으므로 빈 RiskMinor, Applied 및 unknown latch를 보수적으로 해석하고 불확실할 때 unknown; 추정으로 원본 Known을 생성하지 않는다.
3. over-budget는 부패나 자동 거절 대상으로 분류하지 않고 예산 초과 당시의 실제 stored amount와 latch 그대로 내보낸다. 다만 형식/음수/overflow·digest mismatch 등은 오류. overage latch 유무로 정상·비정상 여부를 임의 정정하지 말고 native semantics를 보존한다. 날짜/통화 단위·scope/generation·plan/budget digest를 포함하고, 각 lane의 서로 다른 preimage·lineage digest 관계를 확실히 식별한다.
4. versioned canonical bytes는 크기/문자열/컬렉션 제한, duplicate JSON key·unknown fields·trailing data·잘못된 값/알 수 없는 버전 거부, exact deterministic encoding과 content digest 검증. digest는 우발적 손상 확인이지 서명/인증이 아니다. export 이후 byte를 deep copy하고 재open은 inspection DTO만 반환. 누락된 과거 원장 사실을 DTO나 0으로 채우지 않는다. 원장 migration·실거래·후속 qFinal·UI/API/CLI 연결 없음.
5. 파일은 owner-only 별도 synthetic 경로, 경로·symlink/일반파일 제한, exclusive writer와 partial write/short write 에러, temp 파일 sync→atomic rename→directory sync. 충돌은 overwrite 정책을 명시하고 crash 시 기존 증거 보존; 기존 운영 파일 경로 사용 금지. fsync 실패 후 success를 반환하지 않고 이후 파일이 실제 보이더라도 성공했던 저장으로 소급 주장하지 않는다. 원장과 동일 시점 원자성은 없다.

## reuse 조사

- `internal/strategyobserve/observe.go:211-255`는 bounded canonical envelope 및 strict JSON/단일 개행을 구현하나 `Diagnostic` 전용 검증; 일반 codec으로 직접 재사용 불가, 중복 JSON 키 거부도 별도 확인 필요. `internal/reversallane/schema.go:35-64`는 size, duplicate keys, unknown fields 및 trailing value 거부하지만 private helper라 타 package에서 직접 호출 불가. 검증 방식 참고; public 공용화 시 기존 caller 영향·FLM 별도 검토.
- `internal/strategyobserve/store_linux.go:46-103,118-133,199-210,371-402`는 nofollow 경로/파일, 디렉터리 flock, 제한 읽기, file fsync→rename→directory fsync를 이미 구현한다. 그러나 `Store`는 a155 diagnostic ledger·자체 marker/schema/checkpoint에 묶여 있으므로 새 arbitrary artifact store로 그대로 쓰면 계약 위반; private helper도 직접 호출 불가. 최소 독립 sidecar 전용 파일 처리 또는 작은 helper 추출을 비교하되 기존 a155 의미 변경/guard 우회 금지.
- `internal/attest/attest.go:181-202`의 공개 Save는 rename이 atomic이지만 fsync/no-follow 경계가 없어 durable 안전 요건의 완전한 대체가 아니다. 기존 atomic이라는 이름만으로 crash durability를 주장하지 않는다. 새 의존성 필요 증거 없음.

## 필수 검증

각 lane 정상 plan+현재 nonzero risk deep-copy roundtrip 및 0/unknown/overbudget·latch·fill/cancel 보존; weekly cancel N/A, ReservationState 미포함; native invalid plan/seal·조작 가능한 state 표시 한계; unknown version/duplicate key/trailing/size·금액 overflow·digest/scope tamper; 원본 바이트·입력 맵 변경없음. 실패 주입 write/sync/rename/dirsync 및 충돌·경로/심링크·restart. 검사 결과에 native plan/risk/FX object, 거래 권한, qFinal 또는 live 복원 claims가 없어야 한다. existing old ledgers의 증거 누락은 unavailable. 실행 시험은 아직 하지 않았다.
