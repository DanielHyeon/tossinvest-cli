# Issues — a073-operate-multi-market-strategy-lanes

아카이브 시점(2026-09-08)에 **측정해서** 남기는 부채다. 각 항목은 본선 spec 의
좁힌 문장과 짝을 이룬다. 근거는 비테스트 import·호출자 수이며, 세는 방법은
`rg -l '<import path>' --glob '*.go' | grep -v _test.go` 다.

## D-1 배포 가드가 교체 경로에 없다

`internal/deployguard` 의 **비테스트 import 는 0** 이다. 유일한 import 는 자기 시험
파일이다. 즉 preimage 동결·frozen order·5분 상한·`ROLLBACK_INCOMPATIBLE` 은 코드로
구현돼 있고 시험이 지키지만, **실제 service 교체가 그 코드를 부르지 않는다.**

`docs/operations.md` 도 같은 말을 한다: 이 라이브러리는 "plain evidence 를 검증하고
다음 한 단계의 action **값**만 반환"하며 Docker/process 실행기, engine control,
config/journal/protection writer, broker capability 를 갖지 않는다. 교체는 사람이
`docs/operations.md` 절차로 한다.

- 위험: 사람이 절차를 건너뛰면 아무것도 막지 않는다. 시험이 초록인 것이 교체가
  안전했다는 증거가 되지 못한다.
- 범위: 교체 경로가 이 라이브러리를 실제로 부르게 만드는 change.
- 본선 spec: `http-api-service` — "dormant deployment 교체 규칙은 실행 가능한 계획
  라이브러리로 동결된다" 안에 명시했다.

## D-2 파생 성과 저장소를 채우는 것은 자동이 아니다

`internal/performance` 의 읽는 쪽은 배선돼 있다(`internal/httpapi/read.go`,
`internal/console/performance_history.go`). 그러나 **채우는 쪽**은
`tossctl performance project-attribution` 이라는 별도 CLI 하위명령
(`cmd/tossctl/performance_project.go`, `cmd/tossctl/root.go` 에 등록)이고 엔진
사이클은 그것을 부르지 않는다.

- 위험: 투영을 안 돌린 상태의 빈 표본을 "성과 0" 으로 읽을 수 있다.
- 범위: 투영을 운영 루프에 넣는 change. 넣을 때 D-1 과 같은 질문을 해야 한다 —
  누가 언제 부르는가.
- 본선 spec: `lane-performance` Purpose 와 `operator-console` 성과 Requirement.

## D-3 빌린 증거의 인용은 빌려준 쪽이 바뀌면 낡는다

a073 은 FLM 증거를 a072 에서 통째로 빌린다(231 번들). a073 이 테스트 둘의 이름을
바꿨지만 빌린 번들의 인용은 옛 이름에 남아 있었다. 이번에 고쳤다(review.md 참조).

- 구조적 위험: 공유 증거 풀은 **빌리는 쪽의 편집**으로 낡는다. 빌리는 change 가
  코드를 바꾸면 그 change 가 풀을 갱신할 책임이 있는데, 그 책임을 아무 검사도
  이름으로 부르지 않는다.
- 지금 막는 것: 인용된 테스트가 현재 트리에 실재하는지 보는 검사(a073 보다 나중에
  생겼다)가 이 종류를 잡는다. 좌표의 **역할**까지는 base revision 번들에서 보지 않는다.

## D-4 시장 활성화는 여전히 사람 결정이다

이 change 는 두 시장을 **휴면으로** 배포했다. KR/US 카드가 값을 갖는지는 활성화의
함수이고, 활성화는 서명된 사람 결정이다(a112 8.7.1). 화면·API 계약이 참이라는 것과
두 시장이 돌고 있다는 것은 다른 문장이다.
