# a109 구현 중 발견한 계약 결함

분류 규칙: ① blocking(안전·동작 모순) → 구현 중단 + 보고 종료 · ② safe local → 구현하며
사후 기록 · ③ editorial → 즉시 수정.

## T2-1 (safe local) — D4 가 열거한 nil 검사 2곳 밖에 production REST 경로가 하나 더 있다

- **발견**: design D4 와 freeze P1-4 는 소비자측 nil 검사를 두 곳으로 열거했다 —
  `cmd/tossctl/httpapi_reader.go:566` 과 `internal/httpapi/strategy_runtime.go:18`.
  실제로는 **세 곳**이고, 셋 중 SSE helper 쪽(`strategy_runtime.go:18`)은
  `StrategyRuntimeSnapshotFunc` 의 것인데 **production 호출자가 없다**
  (`rg StrategyRuntimeSnapshotFunc` = 정의 1 + contract test 1).
  열거에서 빠진 `internal/httpapi/router.go:154` 가 REST `/api/v1/strategy-runtime`
  경로의 진짜 부재 판정이다:

  ```go
  case "strategy-runtime":
      if r.strategyRuntime == nil {
          return strategyprojection.DormantSnapshot(r.now().UTC()), nil
      }
  ```

- **왜 결함인가**: 재부착 wrapper 는 정의상 non-nil 이므로 이 검사가 영원히 거짓이 된다.
  전략 화면을 안 쓰는 배포에서 REST 응답이 **dormant 스냅샷 → 오류**로 바뀐다. 그것이
  http-api-service delta 의 「재부착 전의 응답 값은 기존 부재·unavailable 구분을 그대로
  유지해야 한다(SHALL)」를 REST 경로에서 거짓으로 만든다. a108 D4-2 가 금지한 접힘의
  같은 모양이다.
- **분류**: safe local. 설계 의도(부재/unavailable 구분 보존)를 **바꾸지 않고 완성**하는
  기계적 동일 수정이다 — 세 자리 모두 같은 공유 판정 하나로 교체한다.
- **처리**: `internal/httpapi/router.go` 를 T2 표면에 더하고(1줄), 판정은
  `internal/httpapi/strategy_runtime.go` 에 **한 벌만** 둔다
  (`StrategyRuntimeAbsent` + `StrategyRuntimePresence`). 복사한 검사는 어긋나기 시작한
  검사다(a098 D7.1) — 세 곳에 세 벌을 두지 않는다.
- **Manager 확인 요청**: T2 표면 목록에 `internal/httpapi/router.go` 추가.

## T2-2 (editorial) — 격리 해제 오귀속 문구는 한 곳이 아니라 세 곳이다

- **발견**: design D3a-2 는 `internal/console/exit_quarantine.go:227–229` 한 자리를
  인용했다. 같은 문자열이 **세 벌** 있다 — `:161`(release preview),
  `:196`(release apply), `:229`(`writeQuarantineError` 의 `ErrUnwired` 가지).
- **왜 결함인가**: 인용된 file:line 만 고치면 같은 값의 사본 둘이 살아남고, 운영자가
  실제로 먼저 만나는 것은 preview 경로(:161)다.
- **분류**: editorial — 정정의 단위는 줄이 아니라 **값**이다. 즉시 세 곳 모두 고친다.
- **처리**: 세 문자열을 같은 문구로 바꾸고, 핀 테스트가 **세 경로 전부**를 확인한다.

## T2-3 (editorial) — proposal 의 fatal 줄번호는 freeze 판결대로 :274/:279/:315 다

- proposal 본문의 `:274/:279/:315` 는 현재 HEAD 와 일치함을 AST 로 재확인했다
  (`analysis/function-logic/cmd-tossctl--runenginerun/ast.json` returns[8]=274,
  returns[9]=279, returns[11]=315). freeze P1-9 가 정정한 :294 표기는 남아 있지 않다.
  **수정 불요** — 확인만 기록한다.

## T2-4 (기록) — D5b 실측은 편집 전에 끝냈고 결과는 "사전 wipe 불요"다

- tasks 3.3 은 §3 소속이라 T2 가 체크하지 않는다. 측정 결과는 완료 보고에 담고
  Manager 가 배포 절차 prose 에 반영한다.
- 요약: 구버전(HEAD, pre-a109) 세 Start × 잔재 6모양 = 18/18 START OK, 잔재는 그대로
  생존. 코드 근거: `internal/positionpolicyrpc`·`internal/app/engine` production 코드에
  `os.ReadDir` 호출 0건(테스트 3건뿐) — 구버전 회수는 디렉터리를 열거하지 않으므로
  `.s-*` 잔재를 **볼 수 없다.**
