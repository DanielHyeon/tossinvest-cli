# a091 설계 — 0주는 캡이 아니라 실패다

## D1 — 새 이벤트 종류를 만든다

`SeverityOf`는 `criticalEvents` map만 보는 순수 함수다(`event.go:309-314`, AST branches 1).
**등급은 이벤트 종류에 붙어 있다.** 따라서 같은 종류 안에서 등급을 나눌 수 없다.

선택지는 둘이다.

| 안 | 방법 | 문제 |
| --- | --- | --- |
| A | `EventExitProposalCapped`를 `criticalEvents`에 추가 | **부분 캡까지 critical**이 된다. 8/2가 보여준 결함은 0주이고 부분 캡은 축소된 노출이다 |
| B | **0주 전용 이벤트 종류 신설** + 그것만 `criticalEvents`에 등록 | 소비자(콘솔 필터·로그)를 확인해야 한다 |

**B를 택한다.** A는 등급을 사실보다 넓게 올려 critical 채널을 희석한다 — a089 리뷰가
반복해서 지적한 실패 형태다.

기존 `EventExitProposalCapped`는 **부분 캡 전용**으로 좁아지고 등급·문구가 그대로 남는다.

## D2 — 보호/익절 구분은 호출자가 넘긴다

`applyFloor`는 제안이 손절인지 익절인지 모른다(FLM Inputs 표 — 맥락이 인자에 없다).
`submit`은 `proposal exitpolicy.Proposal`을 갖고 있고(`:1237`) `isProtective`가 이미 있다
(`:1217-1219`). 인자를 하나 더 넘긴다.

**판정기를 건드리지 않는다.** `isProtective`는 기존 술어를 그대로 쓴다.

## D3 — 0주가 되는 두 경로를 같게 다룬다

| 경로 | 조건 | 현재 |
| --- | --- | --- |
| B2 `:1408→:1414` | 확정 하한을 계산할 수 없다 | `logErr` 한 줄, **알림 없음** |
| `:1446` | 확정 하한이 0을 허용한다 | `EventExitProposalCapped` (normal) |

운영자에게 둘의 차이는 없다 — **손절이 한 주도 안 나갔다.** 원인은 `detail`에 담고
등급과 종류는 같게 한다.

B2는 지금 알림이 아예 없으므로 이 change가 알림을 **추가**한다. `logErr`은 유지한다
(오류 객체를 담는 유일한 자리).

## D4 — 문구

0주일 때 "일부만 나갔다"는 거짓이다. 제목과 본문을 결과에 맞춘다.
부분 캡의 문구는 **건드리지 않는다** — 그 경우엔 참이다.

## 범위 완결성 — exit 루프의 알림 7종을 전수 확인했다

승격 대상이 하나뿐인지 확인하기 위해 exit 관측 루프가 `o.alert`로 올리는 `obs.Event`를
전부 세고 각각의 등급을 `criticalEvents`와 대조했다.

> **정정(a092 작업 중 발견)**: 처음 이 절은 `exitloop.go` **파일 안**의
> `o.alert(ctx, obs.Event{`만 세어 **6종**이라고 썼다. 일곱 번째
> `EventExitSnapshotQuarantined`는 같은 `o.alert`를 부르지만
> `exit_quarantine_announce.go:72`에 있어서 파일 단위 열거에서 빠졌다.
> **이미 critical이므로 아래 승격 결론은 바뀌지 않는다** — 틀린 것은 완전성 주장이다.
> 열거의 단위는 파일이 아니라 **호출자**여야 했다.

| 줄 | 이벤트 | 등급 | 함수 |
| --- | --- | --- | --- |
| `:781` | `EventExitObservationOutage` | CRITICAL | `checkOutage` |
| **`:1431`** | **`EventExitProposalCapped`** | **normal** | `applyFloor` ← **이 change** |
| `:1501` | `EventExitPositionUnmanaged` | normal | `alertUnmanaged` |
| `:1527` | `EventExitJudgementRefused` | CRITICAL | `alertRefused` |
| `:1551` | `EventExitProposalRefused` | CRITICAL | `alertProposalRefused` |
| `:1581` | `EventExitLiquidationDelayed` | CRITICAL | `noteDelay` |
| `exit_quarantine_announce.go:72` | `EventExitSnapshotQuarantined` | CRITICAL | `announceQuarantine` |

normal은 둘이고, 나머지 하나(`EventExitPositionUnmanaged`)는 **승격하지 않는다.**

그 본문은 "손절·익절이 자동으로 걸려 있지 않다"라고 말하므로 표면적으로는 같은 종류로
보인다. 다르다 — 그것은 **운영자가 선택한 정상 상태**다. 편입되지 않은 보유는 엔진이
관리하지 않기로 한 것이고 편입은 §0.7 사람 항목이다(`engine.adoption`). 승격하면
편입하지 않은 모든 보유가 영구적으로 게이트를 잠근다.

`applyFloor`의 0주는 반대다 — **엔진이 이미 보호하기로 한 포지션의 보호가 실패한 것**이다.
경계는 "엔진이 그 포지션을 보호하기로 했는가"이고, 그 술어는 이미 코드에 있다.

## 건드리지 않는 것

- **제출 수량 계산** — `applyFloor`의 반환값(B1~B6·`:1446`)은 한 글자도 바꾸지 않는다.
  §0.3·§0.9가 걸리는 자리는 여기이고, 이 change는 보고만 바꾼다
- **0주가 되는 원인** — RECONCILE 확정 하한은 대사 영역
- **부분 캡의 등급·문구**
- **`exitpolicy` 판정기**
- **스키마**

## 검증

- 보호 + 0주(두 경로 각각) → critical 등급, outbox 행 생성
- 보호 + 부분 캡 → **종전 등급·문구 무변화**
- 익절 + 0주 → **종전 등급 무변화**
- 미등록 이벤트 종류는 여전히 normal (`SeverityOf` 기본값 보존)
- **반환값 회귀**: 위 전부에서 `applyFloor`의 (수량, capped, err)가 무변화
- 기존 18종 critical 등급 무변화
- 8/2 재생: 13회 시퀀스를 fixture로 돌려 outbox 행이 생기는지 확인
- `go test ./... -count=1 -race` 회귀 0
