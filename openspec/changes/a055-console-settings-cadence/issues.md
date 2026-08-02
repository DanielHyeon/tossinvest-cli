# Issues — a055-console-settings-cadence

## I1. `편입 설정 화면`의 MODIFIED delta 스택이 본문보다 오래됐다

**분류**: blocking-for-archive (구현 차단 아님, 아카이브 순서 결정 필요)
**발견**: 2026-08-02, 이 change의 계약 작성 중

### 사실

`openspec/specs/operator-console/spec.md:155`의 `편입 설정 화면`은 다음 문장을 갖고 있다.

> 상단 navigation은 **"외부 종목 자동관리"** 메뉴를 표시하고 `/settings#adoption`으로 기존
> 편입 설정의 첫 섹션을 직접 열어야 한다(SHALL).

같은 요구사항을 MODIFY하는 미아카이브 change가 3건 있다.

| change | 파일 |
|---|---|
| `console-excludes-in-one-click` | `specs/operator-console/spec.md:41` |
| `console-sets-guardian-limits` | `specs/operator-console/spec.md:45` |
| `console-operator-overview` (계열) | `specs/operator-console/spec.md` |

**그 셋 중 어느 것도 위 문장을 담고 있지 않다** (`grep '외부 종목 자동관리'
openspec/changes/*/specs/operator-console/spec.md` → 0건). MODIFIED는 요구사항 블록 전체를
치환하므로, 지금 그 change들을 아카이브하면 본문이 그 문장을 **잃는다**.

같은 형태의 부채가 `콘솔 안전 불변식`에도 있다 — MODIFY하는 미아카이브 change가 **6건**이다.

### 이 change가 한 것

- delta의 base를 **현재 승인된 본문**으로 명시하고 그 사실을 delta 머리에 적었다
  (WORKFLOW 권위 경계: 의도된 동작의 권위는 `openspec/specs/` + 승인된 change).
- `콘솔 안전 불변식`은 MODIFY하지 않았다. 신설 라우트 4건이 전부 GET이고 폼이 없어
  상태변경 행위 목록을 늘리지 않으므로 MODIFY할 필요가 없다.
- 선행 change `a054-console-status-shell`도 같은 이유로 **ADDED만** 쓴다.

### 남은 결정 (리뷰가 판단할 것)

1. 미아카이브 console 계열 change 중 **이미 구현·배포된 것**이 어느 것인지 확인하고
   아카이브하거나 폐기한다. 코드에는 Guardian 한도·주문 화면·개요가 전부 존재하므로
   최소 세 change는 구현 완료 상태로 보인다.
2. 아카이브 순서를 정한다. 이 change보다 **먼저** 아카이브되는 delta가 있으면 이 delta의
   base가 달라지므로 본문을 다시 맞춰야 한다.
3. 이 change를 그 정리보다 먼저 아카이브할지 결정한다.

이 change 단독으로 결정하지 않는다 — 다른 change 6~9건의 계약에 영향을 준다.

## I2. 죽어 있는 인라인 확인 대화

**분류**: safe local (이 change의 범위 안에서 처리)

`templates_settings.go:111`의 Guardian 프리셋 폼은
`onsubmit="return confirm('… 한도를 기록한다. 다음 엔진 기동부터 반영된다.')"`를 갖고 있다.

배포 CSP는 `default-src 'none'`이고 `script-src`가 없다. 인라인 event handler는
`script-src`(없으면 `default-src`) 지시를 따르므로 **실행되지 않는다.** 이 확인 대화는
한 번도 뜬 적이 없고, 운영자는 프리셋 버튼을 누르는 즉시 다섯 값이 기록되는 것을
확인 없이 겪어 왔다.

기존 `편입 설정 화면` 요구사항이 "같은 template의 다른 legacy confirm handler는 이 change
범위 밖이다"로 남겨 둔 handler가 바로 이것이다. 이 change의 카드 표준이 **적용 전
미리보기**로 그 자리를 대신하므로 handler를 제거하고, 설정 화면 전체에 inline handler가
없음을 정적 검사로 고정한다(delta의 `설정 화면의 인라인 handler 부재` Scenario).
