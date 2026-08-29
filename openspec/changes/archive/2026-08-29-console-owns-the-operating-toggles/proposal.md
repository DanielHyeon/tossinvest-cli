# Change: console-owns-the-operating-toggles

> 2026-07-30. 사용자 지시: "자꾸 수동으로 뭔가를 설정하라고 하지? 메뉴를 만들어 주던가 해야지."

## Why

엔진을 띄우는 데 필요한 마지막 두 가지 — 거래 정책 토글과 자동화 게이트 — 는 지금
`~/.config/tossctl/config.json` 손편집으로만 켤 수 있다. CLI에도 없고(`tossctl config`는
`init`/`show`뿐), 콘솔에도 없다. 콘솔 편입 설정 화면은 그것을 명시한다:

> "automation gate(운영 게이트)의 ON/OFF와 kill switch는 이 콘솔에서 편집할 수 없다 —
> 게이트 ON은 콘솔 밖 승인 절차이고"

그 근거는 `.claude/CLAUDE.md` §0.7 "운영 토글 flip과 live 검증은 사람이 직접 승인한다"다.
**그 문장은 "콘솔에 두지 말라"고 하지 않는다.** 로컬 콘솔에서 사람이 버튼을 누르는 것이
곧 사람이 직접 승인하는 것이다. 콘솔을 배제한 것은 조항 6과 같은 종류의 과잉 해석이며,
같은 대가를 치르고 있다.

### 대가는 안전이 아니라 반대다

`size-us-guardian-tier` design이 오늘 아침 같은 논증을 이미 적었다 — 상한이 콘솔 쓰기
경로에만 있으므로, 너무 낮은 상한은 운영자를 `config.json` 손편집으로 밀어내고 **거기에는
상한 검사가 없다.** 게이트와 거래 정책은 그보다 심하다. 손편집 경로에는:

- 인터록이 요구하는 다섯 한도가 다 설정됐는지 확인하는 것이 없다
- 게이트를 켜면 무엇이 시작되는지 알려주는 것이 없다 (특히 **편입은 되돌릴 수 없다**)
- JSON 문법 오류를 잡는 것이 없다 — 잘못 고치면 엔진이 설정을 못 읽는다
- 감사 기록이 없다. 콘솔 저장은 audit에 남고, 손편집은 남지 않는다

즉 지금 배치는 가장 위험한 토글을 **가장 검증이 없는 경로로** 밀어내고 있다.

### 추가로 발견된 것

인터록 3절은 `trading.sell`과 `trading.allow_live_order_actions`만 검사한다. 그런데
exit 루프가 매도를 내려면 `trading.place`도 필요하고(`internal/trading/service.go:116` —
매도든 매수든 제출은 place다), 취소도 하므로 `trading.cancel`도 필요하다.
둘만 켜면 **엔진은 뜨고 첫 손절에서 거부된다.** 3절 주석은 "매수는 가능한데 청산이
불가능한 조합으로는 기동할 수 없다"고 쓰는데 정작 청산에 필요한 두 개를 안 본다.

## What Changes

- **`/settings`에 "운영" 섹션 두 개**:
  - **거래 정책** — `place`·`sell`·`cancel`·`allow_live_order_actions` 체크박스.
    exit 루프가 실제로 쓰는 네 개만. `amend`·`conditional`·`fractional`은 화면에 없다.
  - **자동화 게이트** — ON/OFF 하나. 켜기 전에 무엇이 시작되는지 화면이 말한다.
- **인터록 3절 확장**: 청산에 필요한 `place`와 `cancel`을 검사에 넣는다. 3절이
  이미 주장하던 것을 실제로 검사하게 만드는 것이므로 요구를 넓히는 것이 아니라
  주석과 코드를 일치시키는 것이다.
- **게이트 ON 사전 판정**: 화면이 켜기 전에 인터록이 거부할 이유를 보여준다.
  운영자가 켜고 → 엔진이 거부하고 → 로그를 읽는 왕복을 없앤다.
- **`LimitSettings.Save`의 금지 해제**: `settings_limits_test.go`가 "Save는 스위치를
  실을 수 없는 타입을 받아야 한다"를 단언한다. 그 단언이 이 과잉 해석의 코드화이며,
  게이트 스위치는 **자기 seam**으로 분리해 옮긴다 — 한도 저장이 실수로 스위치를
  건드리는 일은 여전히 불가능하다.

### 하지 않는 것

- 타이핑 확인·추가 승인 마찰을 넣지 않는다 (사용자 지시, 기억 `no-typed-confirmation-friction`).
  버튼 하나다.
- kill switch는 이 change의 범위가 아니다.
- 인터록의 다른 절, 한도, attestation, 게이트 OFF 동작을 건드리지 않는다.
- 게이트를 대신 켜지 않는다. 화면을 만들 뿐, 누르는 것은 사람이다.

## Capabilities

### Modified Capabilities

- `operator-console`: Guardian 한도 설정 화면에 운영 토글 두 섹션 추가, 게이트 ON
  사전 판정, 감사 기록.
- `engine-safety`: 인터록 3절이 청산 경로가 실제로 요구하는 네 토글을 검사한다.

## Impact

- Affected code: `internal/console`(신규 섹션·seam·핸들러), `cmd/tossctl`(seam 주입),
  `internal/app/engine/interlock.go`(3절), `internal/config`(거래 정책 surgical write).
- 사용자 영향: 엔진 기동에 필요한 전부가 화면에서 끝난다. `config.json` 손편집 불필요.
- 위험: 게이트를 켜기가 쉬워진다. 그것이 목적이다 — 어려운 쪽이 안전한 것이
  아니라 검증 없는 쪽으로 밀어내고 있었다. 게이트 ON은 여전히 사람이 누르고,
  여전히 audit에 남으며, 이제는 **감사 기록이 남는 경로로** 눌린다.
