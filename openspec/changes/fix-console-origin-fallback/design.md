## Context

원격 콘솔의 `remoteRuntime.sameOrigin`은 상태 변경 POST보다 먼저 실행된다. 현재
우선순위는 `Origin`, 그 값이 비면 `Referer`이며 둘 다 없으면 무조건 거부한다.
동시에 원격 응답은 개인정보 보호를 위해 `Referrer-Policy: no-referrer`를 보낸다.
일부 브라우저는 same-origin 폼 POST에 `Origin`도 생략하므로 canonical
`https://127.0.0.1:37085`에서 제출한 모든 설정이 handler 전에 막힌다.

외곽 `remoteRuntime.security`는 이미 실제 peer CIDR과 exact canonical Host를
검사하고 TLS listener만 사용한다. 상태 변경 경로에는 별도의 무작위 CSRF 토큰도
필수다.

## Goals / Non-Goals

**Goals:**

- privacy header 때문에 `Origin`과 `Referer`가 모두 없는 same-origin 브라우저 POST를
  정상 처리한다.
- scheme·host·port의 canonical boundary와 CSRF를 함께 유지한다.
- 명시적으로 제공된 잘못된 `Origin`/`Referer`는 계속 fail-closed한다.
- 상태 변경 POST에만 headerless fallback을 적용하고 compatibility login은 기존 strict
  header 판정을 유지한다.

**Non-Goals:**

- 여러 public origin, wildcard host, forwarding header 신뢰를 추가하지 않는다.
- application login/token을 되살리거나 trusted-network 범위를 바꾸지 않는다.
- handler별 저장 동작, 운영 토글 승인, 엔진 interlock 또는 주문 경로를 바꾸지 않는다.
- URL path를 origin identity에 포함하지 않는다.

## Decisions

1. strict header 판정과 mutation-only fallback 판정을 분리한다.
   `sameOrigin`은 compatibility `/login`과 mutation의 명시적 Origin/Referer를
   검증하고, 새 `sameOriginForMutation`만 두 헤더 key가 아예 없을 때 TLS+Host를
   사용한다. 따라서 login origin 정책은 변경되지 않는다.

2. 헤더 존재 여부는 `Header.Get`의 trim 결과가 아니라 canonical header map의 key
   존재로 판정한다. Origin/Referer key가 존재하는데 값이 비었거나 공백이거나 여러
   값이면 명시적 오류로 거부한다. Origin key가 있으면 Referer보다 우선하며,
   Referer가 malformed이더라도 읽거나 평가하지 않는다. 반대로 Origin이 잘못됐을 때
   canonical Referer로 fall through하지 않는다.

3. headerless fallback은 `r.TLS != nil`이고 `"https://"+r.Host`가 configured
   origin과 정확히 같을 때만 허용한다. `X-Forwarded-Host`,
   `X-Forwarded-Proto`는 사용하지 않는다. TLS 종료 proxy를 지원하기 위한 예외도
   추가하지 않는다.

4. CSRF는 기존 `Console.mutating`의 다음 gate로 유지한다. headerless 요청이
   TLS+Host를 통과해도 올바른 페이지의 process-local CSRF가 없으면 handler에
   도달하지 않는다.

5. `Referer` fallback은 URL을 parse한 뒤 scheme과 host만 사용한다. 하위 path,
   query, fragment는 origin 비교에 관여하지 않는다.

## Risks / Trade-offs

- [브라우저가 origin 계열 헤더를 모두 숨겨도 요청을 허용함] → direct TLS, exact Host,
  peer CIDR, process-local CSRF가 모두 독립적으로 통과해야 한다.
- [TLS 종료 reverse proxy에서는 fallback이 거부됨] → forwarding header를 신뢰하지
  않는 현재 위협 모델을 유지한다. VPN/mobile 배포도 TossOS가 TLS를 직접 종료한다.
- [공통 함수 변경이 모든 POST에 영향] → handler를 호출하는 무해한 test seam으로
  gate 순서와 explicit mismatch 거부를 회귀 테스트한다.
- [compatibility login의 anti-forgery 계약 약화] → login은 strict `sameOrigin`만
  호출하며 Origin/Referer key가 모두 없는 요청을 계속 거부하는 회귀 테스트를 둔다.

## Migration Plan

1. 회귀 테스트와 Function Logic Map을 먼저 추가한다.
2. strict `sameOrigin`과 새 mutation-only predicate를 구현하고 `mutating`의 한
   호출부만 새 predicate로 전환한다.
3. 이미지를 재빌드해 기존 Compose service를 recreate한다.
4. rollback은 이전 이미지 digest 또는 직전 commit의 `sameOrigin` 구현으로 되돌린다.
   config·DB·journal migration은 없다.

## Open Questions

없음. 사용자가 제공한 실패 URL로 canonical host와 POST path가 확인됐고, fallback
범위는 CSRF가 뒤따르는 상태 변경 gate로 한정된다.
