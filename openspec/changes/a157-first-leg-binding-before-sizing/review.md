# a157 review

## proposal-freeze 리뷰 (2026-10-18)

- 보이스: 분리 컨텍스트 둘(적대 Eng).
  - Opus 서브에이전트: 저장소 밖 사본에 A1 을 실제로 적용하고 probe 시험을 HEAD · A1 양쪽에서 돌렸다.
  - codex `exec -s read-only`.
- 원문: `analysis/review/freeze-opus.md` · `analysis/review/freeze-codex.md`.
- 판정:
  - Opus **APPROVE-WITH-FIXES**(P0 0 · P1 3 · P2 4 · P3 6). 받아들이는 집합 불변을 실측했고, 주문을 늘리는 갈래는 없다.
  - codex **REJECT as written**(P1 1 · P2 3). D2-c 에서 a137 강화(`ENTRY_BLOCKED` 영속)가 사라진다는 지적이다.
- 사람 결정(2026-10-18, 「1-1, 2-1」):
  - 1-1: 일일 손실 체인 거절을 결속 결함보다 먼저 보고한다(codex P1 수용 — 강화 보존).
  - 2-1: 조립 감독자에 `EntryGate` 를 nil 가드와 함께 배선한다.

| 지적 | 처분 | 반영 자리 |
|---|---|---|
| codex P1 · Opus P2-4 D2-c 강화 상실 | 수용(사람 1-1) — 보고 우선순위 「일일 손실 > 결속 > 그 밖」, 판별 함수 하나를 `admit` 과 공유 | design D1 · D2(c)(d) · spec ADDED · tasks 0.4/1.2 · M13/M14 |
| Opus P1-1 「결함 하나면 그대로」 거짓(실측 2) | 수용 — 「한쪽 판정만 거는 요청」으로 다시 씀, D2(f) 추가 | spec 시나리오 · design D2(f) · D3 · tasks 0.4 |
| Opus P1-2 `EntryGate: c.Entry` 타입 있는 nil → panic | 수용(사람 2-1) — `if c.Entry != nil` | design F10 · D4 · D5 · spec MODIFIED |
| Opus P1-3 부분 이동 · 복제를 못 잡음 | 수용 — 상한 0 + Bn 일곱(B5 둘), 구조 AST 시험, M2~M8 부분 이동 · M15 복제 | design D6 · tasks 0.4/1.1 |
| Opus P2-1 sha 대조 불가(파일 sha) | 수용 — qfinal.go 는 파일 sha, 첫 다리 발급은 함수 본문 sha | design D3 · tasks 1.3 |
| Opus P2-2 dedup 수명 · Run 시험 닿지 않음 | 수용 — 시나리오를 「같은 함수 · 같은 게이트」로, 시험은 칸 단언 + 기록자 직접 호출, dedup 상향은 후속 | design D4 · spec · tasks 후속 |
| Opus P2-3 spec 자기모순(일일 손실 결과 · a137) | 수용 — MODIFIED 에 단서, 1-1 로 a137 충돌 소멸 | spec |
| codex P2 D2 열거 불완전(강화 실패 · 모인 범위 오류 · 측정 불가 기록) | 수용 — 1-1 로 (c)(d) 는 Y 그대로가 되어 모두 보존, 표에 명시 | design D2 |
| codex P2-3 · Opus P3-2 시계 주장 과장 | 수용 — 「같은 입력 · 같은 시계 관측」으로 한정, 지터 명시 | design F1 · D2 · spec |
| codex P2 시험 공백(엔진 경로 · 강화 성공/실패) | 수용 — 엔진 행 B2 · B8 × 일일 손실 둘 | tasks 0.4 · design D3 |
| Opus P3-1 :392 → :394 | 수용 | proposal · design F9 |
| Opus P3-3 F3/F4 는 원문 읽기 | 수용 — 근거 표기 정정 | design 머리 · F3 · F4 |
| Opus P3-4 SHALL NOT 범위 | 수용 — 「일곱 결속 판정 각각, precheck 안에서」 | spec |
| Opus P3-5 「모든 생산 조립」 모호 | 수용 — 두 `Context` 생성자 이름 명시 | spec MODIFIED |
| Opus P3-6 컴파일 실패는 변이 아님 | 수용 — 빌드 메모로 | design D6 |

- 거절: 없음.
- 사람 결정 1-1 로 범위가 넓어졌다. `admit` 의 일일 손실 두 조건이 판별 함수로 옮겨 가므로 `admit` 이 편집 대상에 들어간다(FLM/BTM 필요).
  이 범위 확장이 다른 보고 결과를 바꾸지 않는지는 구현 뒤 독립 리뷰(2.3)에서 다시 본다.

## Pre-Edit Gate

```text
Pre-Edit Gate:
- change id / task id: a157-first-leg-binding-before-sizing / 1.1~1.3
- 대상 심볼(편집하는 기존 함수 — 번들 analysis/function-logic/, ast-base.json = base 7b1748481 소스):
    internal/execgw/riskguardian_first_leg.go           RiskGuardian.PrecheckQFinalCampaignFirstLeg (본문 순서 재배치 — design D1)
    internal/app/engine/strategy_first_leg_admission.go strategyFirstLegAdmissionBridge.admit (일일 손실 두 조건 → 판별 함수 호출, 분기 결과 동일)
    internal/app/engine/strategy_entry_supervisor.go    Context.NewPairedStrategyEntryProductionAssembly (394 행 옵션 — 알림 둘 + nil 가드 게이트)
  새 파일(새 leaf): internal/execgw/a157_daily_loss_kind.go — 판별 함수 `DailyLossChainRefusalOf`(2.3 수리 b8d6d4cee 에서 `DailyLossRefusalKindOf` 로 개명) 와 결과 타입(기존 파일 무편집 — Logic Map 범위 고정)
  무편집(확인 방법):
    internal/execgw/riskguardian_qfinal.go — 파일 sha 6fcace26fd7b2ddc…(편집 전) 그대로여야 함(PrecheckQFinalEntry · IssueQFinalEntry)
    RiskGuardian.IssuePrecheckedQFinalCampaignFirstLeg — 함수 본문 sha 28c40c8ab49a1cb6…(편집 전, analysis/harness/_funcsha gofmt 정규화 FuncDecl 구간)
    사이징(internal/risk) · 체인 · 버킷(internal/riskbucket) · 원장(internal/journal) · Gateway · 매도/손절/청산/대사 경로 전부
- CodeGraph: analysis/code-context/ — codegraph 1.6.0 callers/impact, affected --filter '*_test.go' 1,153(기본 판별식 1).
  불일치 둘(인터페이스 선언 :65 대 실제 호출 :105, admit 필드 경유 호출 누락)은 HEAD grep 으로 해소. CGC 는 색인 범위가 옛 사본이라 advisory.
- 기존 동작 근거: HEAD 읽기 + analysis/ast-evidence/ 8 + FLM 번들 3 + 특성 시험 13 함수(편집 전 HEAD 통과 — analysis/red/characterization-head.txt)
  + freeze 리뷰 실측(analysis/review/freeze-opus.md)
- Function Logic Map / Branch Test Map: analysis/function-logic/{internal-execgw--riskguardian.precheckqfinalcampaignfirstleg,
  internal-app-engine--strategyfirstlegadmissionbridge.admit, internal-app-engine--context.newpairedstrategyentryproductionassembly}
  (생성: analysis/harness/a157_flm.py base · check_analysis.py 통과)
- upstream 상속 시험 영향: no — 게이트 OFF 는 Guardian · 전략 감독자 경로를 만들지 않음. 편집 함수는 첫 다리(매수) 진입 판정 · 그 분류 · 감독자 조립뿐
- 실패 시험 선행: yes — 1.1 에서 특성 시험의 뒤집을 행(상한 0 + Bn 9 · 버킷/표지 없는 입력 불가 + B2/B8 · (f) 2 · 엔진 상한 0 + B2/B8)을
  새 기대로 바꾸고, 구조 AST 시험 · 판별 표 시험 · 조립 감독자 시험 셋을 더해 RED 를 확인한 뒤(analysis/red/red.txt) 편집
- 설정 · DB · journal 변경: 없음(새 키 0 · 스키마 0 · 원장 쓰기 0 · 서명 매니페스트 0). rollback = 이전 이미지
- 안전 불변식 §0 검토: 통과 —
    받아들이는 요청 집합 불변(design D2: B1~B8 은 부작용 없는 판정의 논리곱, B2~B8 은 시계 · 원장을 읽지 않음, PrecheckQFinalEntry 한 번),
    바뀌는 것은 둘 이상 걸린 요청의 보고뿐이고 어느 갈래도 발급 · 주문을 늘리지 않음,
    일일 손실 강화 · 계좌 단위 진입 차단 보존(사람 결정 1-1 — 판별 함수 하나를 precheck 와 admit 이 공유),
    조립 감독자 EntryGate 는 nil 가드(F10 — 타입 있는 nil panic 방지, 오늘은 Run 안 함),
    사이징 무편집(§0.9), 매도 · 손절 · 비상 청산 경로 무관(§0.3), 토글 OFF 무관
```

## 구현 독립 적대 리뷰 (task 2.3, 2026-10-18~19)

- 보이스: 분리 컨텍스트 둘.
  - Opus 서브에이전트: 저장소 밖 사본에서 편집 전 생산 코드에 새 시험을 돌려 특성 행이 정직한지 대조했고, 직접 변이를 넣었다.
  - codex `exec -s read-only`.
- 대상: `7b1748481..10ab9f87a`.
- 원문: `analysis/review/impl-opus.md` · `analysis/review/impl-codex.md`.
- 판정: 둘 다 **APPROVE-WITH-FIXES**(P0 0 · P1 0). 생산 동작 결함은 없다.

| 지적 | 처분 | 반영 |
|---|---|---|
| Opus P2-1 · codex P2 구조 시험이 글자 위치만 봄 — 이름 붙인 closure 두 번 호출이 SURVIVED(실측), 사이징 뒤 호출도 통과 | 수용 — FuncLit 하나 · 즉시 호출 · 최상위 대입 · 사이징 문장보다 앞을 단언, M15b/c/d 추가 | b8d6d4cee · round2 |
| codex P3 · Opus P3-4 통과 봉인 값 시험이 머리말만큼 단언하지 않음 | 수용 — strategy · campaign · weekly · issuedAt 를 고정 시계 기대값과 정확 대조 | b8d6d4cee |
| Opus P3-1 「판별은 여기 하나뿐」이 저장소 전체로는 거짓(delivery:46 표지만 봄) | 수용 — 주석 범위를 좁히고 tasks 「후속」에 이전 메모 | b8d6d4cee |
| Opus P3-2 `DailyLossChainRefusal` 과 `dailyLossChainRefusal` 대소문자만 다름 | 수용 — `DailyLossRefusalKind(Of)` 로 개명 | b8d6d4cee |
| Opus P3-3 강화 실패 + 결속 결함 엔진 행 없음 | 수용 — 엔진 시험 추가(무결함 · B2 · B8, 같은 joined 결함) | b8d6d4cee |
| Opus P3-5 「동등 변이 없음」과 M14c 긴장 | 수용 — M14c 는 생산 경로에서 동등, 합성 표 행으로만 잡힘으로 표기 | 98d6bd526 |
| Opus P3-6 주석이 편집 시점을 가리킴 | 수용 — 불변식 서술로 | b8d6d4cee |

- 거절: 없음.
- 수리 뒤 검증(Teammate, HEAD b8d6d4cee — `analysis/verify/receipts.md`):
  - 변이 round2 **36/36 CAUGHT**(무변이 대조군 GREEN · 시작/복원 sha 단언 · 전부 대상 도달). M14c 는 생산 경로 동등.
  - `make lint` rc 0.
  - execgw · app/engine 무태그 · `-tags tossos_testseams` 전체 ok, `-race` 대상 ok.
  - `openspec validate --strict` valid, `check_analysis` evidence complete.
  - `riskguardian_qfinal.go` 파일 sha · `IssuePrecheckedQFinalCampaignFirstLeg` 본문 sha 불변.
- Manager 판정: 수리 diff 확인 뒤 2.3 닫음. 남은 것은 3.1(격리 워크트리 게이트 · 착지 기록 · archive)이다.

