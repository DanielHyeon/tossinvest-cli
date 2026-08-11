# a101 High-risk Pre-Edit 선언

`docs/WORKFLOW.md:387-401`의 형식을 따른다.

이 change가 닿는 경로는 **인증(attestation)**이다 — 배선의 산출물이 automation gate가 읽는
capability attestation을 갱신하는 프로세스이기 때문이다(`.claude/CLAUDE.md` §0-5).
주문·손절 경로는 건드리지 않는다.

---

## 1. `cmd/tossctl.runConsole` — 기동 배선과 재시작 seam

```text
Pre-Edit Gate:
- change id / task id: a101 / tasks 2.3, 3.1
- 대상 심볼(패키지.함수): cmd/tossctl.runConsole (console.go:211)
- 기존 동작 파악 근거:
    FLM analysis/function-logic/cmd-tossctl--runconsole/
    측정: cmd/tossctl 49.6% — **runConsole 자체는 0.0%, 분기 41개 전부 미실행**
    호출부: `tossctl console` 하나
- upstream 상속 테스트 영향: no (TossOS 고유 콘솔)
- 실패 테스트 선행 작성: **yes** — cmd/tossctl/soakautostart_test.go.
    RED은 컴파일 실패(`undefined: runConfiguredSoakAutostart`)로 확인했다.
    **단, internal/config 쪽(soak_io.go)은 구현을 먼저 썼다. RED을 관측하지 않았다** —
    review.md에 이탈로 기록한다.
- 안전 불변식 §0 위반 여부 검토: **조건부 통과.** 통과 조건 넷:
    (1) **엔진 autostart보다 뒤에 둔다.** 같은 계좌의 rate budget을 쓰고, 엔진의 기동
        인터록이 서베이의 산출물을 읽는 쪽이다.
    (2) **서베이 기동 실패로 return하지 않는다.** 이 함수의 관통 불변식은 「선택 기능의
        부재는 출력 한 줄이고 기동 실패가 아니다」이며, 그것을 깨면 조회 전용 도구 하나가
        운영자 화면 전체를 없앤다.
    (3) **판정을 runConsole 안에 두지 않는다.** 0.0% 측정이 그 이유다 — 거기 둔 판정은
        어떤 테스트도 부르지 못한다. 판정은 인자를 받는 별도 함수에 두고 여기에는 호출과
        출력만 남긴다.
    (4) **`RestartSoak`의 반환 계약을 바꾸지 않는다.** 승인 기록 실패를 에러로 승격시키면
        대시보드가 "재시작 실패"를 표시하고, 운영자의 자연스러운 반응(다시 누르기)이
        방금 선 서베이를 죽인다.
```

### 측정으로 보장되지 않는 부분 (숨기지 않고 적는다)

조건 (1)과 (2)는 **테스트로 고정하지 않았다.** `runConsole`이 0.0%이므로 그 안의 순서와
early-return 부재를 관측할 테스트를 만들 수 없다. 조건 (3)이 그 사실에 대한 대응이며,
(1)·(2)는 **코드 리뷰 조건**으로만 존재한다.

(4)는 `rememberSoakApproval`이 별도 함수이므로 측정된다
(`TestSoakApprovalFailureDoesNotUndoTheRestart`).
