# a115 issues — 표면 판단과 잔존 기록

## S1 — 파일 표면 해석 (safe-local)

등록 문서(proposal Impact)는 「`cmd/tossctl/console.go` 부팅 경로 + 테스트」라 적었지만, tasks 0.3 이
소비 page 의 FLM 을, Manager 지시가 「화면이 dormant/unavailable 구분」을 범위로 적었다. 부팅 경로만
고치면 wrapper 가 non-nil 이라 두 소비자의 `== nil` 판정이 부재를 「읽지 못했다」로 오귀속한다 —
spec 시나리오 「미구성은 그대로 미구성이다」위반. 따라서 표면에 넣는다: `internal/console` 두 함수의
nil 판정 한 줄씩, 그리고 부재 판정·presence 인터페이스의 `internal/strategyprojection` 이동(freeze 리뷰
P1-1 — 판정은 한 벌, `internal/httpapi` 는 type alias + 위임으로 기존 표면 유지). 등록 문서보다 넓지만
스펙 시나리오가 요구하는 최소이며, 화면 구조·문구는 무변경이다. design D2·「파일 표면 해석」 참조.

## S2 — 열한 번째 FLM 번들: `httpapi.StrategyRuntimeAbsent` (구현 착수 시 발견)

design D2 는 판정을 `internal/strategyprojection` 으로 옮기고 httpapi 이름을 위임으로 남긴다. 위임으로
바꾸는 것은 **기존 함수 본문 편집**이다(분기 2 → 0). freeze 의 열 벌에는 이 함수가 없었다 — Pre-Edit 로
`analysis/function-logic/internal-httpapi--strategyruntimeabsent/` 를 편집 **전에** base 에서 만들었다
(1.1–1.2 커밋에 함께 착지). 옮겨 간 세 갈래는 새 잎 함수 `strategyprojection.StrategyRuntimeAbsent`
(새 파일 `presence.go` — 새 leaf, FLM not-applicable)에 그대로 있다.

## R1 — wrapper 전이 로그의 「데몬」 문구 (표면 밖)

`httpapi_strategy_attach.go` 의 `observe` 전이 로그는 「데몬은 그대로 돈다」를 말한다(a109 코드).
콘솔이 같은 wrapper 를 재사용하면 콘솔 stderr 에서 「데몬」은 이 콘솔 프로세스를 뜻하게 된다.
문구 일반화는 httpapi 표면 편집이라 이 change 밖 — 오독 위험은 낮고(같은 재부착 의미), 필요 시 후속.
첫 부착에도 「다시 붙었다」가 찍히는 표기 문제(freeze 리뷰 P2-9)도 같은 자리다.

## R2 — 규칙 12/13 적용 범위

화면 구조·색·문구 무변경이라 `ui-skills-root` 규칙 선택은 not-applicable(design). 규칙 13 실측은
tasks 1.6: 엔진 없이 콘솔을 띄워 전략 화면 본문을 확인한다(버튼 누르지 않음, 안전 불변식 1·7).
freeze 리뷰 P2-7 반영 — dormant 절반만이 아니라, 격리 config 에 잔재 descriptor 와 죽은 socket 을 두어
unavailable 절반도 잰다(엔진 불필요·부작용 없음).

## R4 — 물려받은 깜빡임 병 (표면 밖, freeze 리뷰 P2-1)

wrapper 의 `observe` 는 **답한** 실패(코드 붙은 rpcError — 엔진 Validate 실패의 503, decode·Validate
거절)도 탈착으로 읽는다. 콘솔에서는 렌더마다 「실패」 로그와 펌프 재-dial 의 「다시 붙었다」 로그 짝이
간격당 최대 한 번 반복될 수 있다. 화면 값은 맞다. httpapi 에서 물려받은 병이고 고치는 일은 wrapper
편집(a114 P1-1 의 전략판)이라 표면 밖 — 후속 후보.

## R5 — http-api spec 의 「반쪽 잔재」 문장과 코드 불일치 (표면 밖, freeze 리뷰 P2-2)

http-api-service spec :206–208 은 「반쪽 잔재 → descriptor 부재 기동과 동일」을 말하는데 현 코드는
dial 실패 시 sentinel(도달 불가)이다. a115 이전부터 있던 spec-코드 불일치 — 후속에서 spec 정정.

## R6 — 붙어 있던 콘솔의 깨끗한 정지는 도달 불가로 보인다 (선언된 한계, 구현 후 리뷰 P2-1)

design 「구성」의 뜻 문단은 깨끗한 정지 뒤 dormant 를 말하지만, 그것은 정지 **뒤에 부팅한** 콘솔 이야기다. live 로
붙어 있던 콘솔은 wrapper 가 live 자리를 부재로 격하하지 않으므로(a109 `attempt` B1 — 「붙어 있다가 잠깐 못 읽는 중」을
「안 쓴다」로 바꾸지 않는다는 의도된 규칙) 엔진이 돌아올 때까지 「runtime projection을 읽지 못했다」를 그린다. 문구는
엔진 부재를 단정하지 않고 회복은 저절로 온다. 고치려면 wrapper 편집(a109 표면)이 필요해 이 change 밖이다. 동작은
`TestAnAttachedConsoleShowsACleanStopAsUnreachable` 이 고정한다 — 바뀌면 시험이 알린다.
