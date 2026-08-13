# Branch Test Map: `readDescriptor`

- Source: `internal/strategyprojectionrpc/transport.go` (159-183) — revision `current`
- Fix 라운드가 형식 검사를 `openVerifiedDescriptor`로 떼어 냈다(design D1-2).

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 형식 검사·열기가 실패했다 | `TestDescriptorPublicationIsAtomic`, `TestStartRefusesUnsafeLeftoverShapes` | no (편집 전 GREEN) | yes |
| B2 | 4096바이트를 넘거나 읽지 못한다 | 없음(무변경·기존 무테스트) | no | no |
| B3 | JSON이 아니다 | `TestStartRecoversFromTruncatedDescriptorLeftover` | yes (`stale descriptor is unsafe`) | yes |
| B4 | 값이 둘 이상이다 | 없음(무변경·기존 무테스트) | no | no |
| B5 | 필드가 유효하지 않다 | 없음(무변경·기존 무테스트) | no | no |

## B3의 방향에 주의

`TestStartRecoversFromTruncatedDescriptorLeftover`는 B3이 **일어나는 것**을 재지 않는다.
잘린 JSON을 잔재로 놓고 기동이 성공하는지를 재는 테스트이고, 그것이 성공한다는 것은
**회수 경로가 이 함수를 부르지 않는다**는 뜻이다. 회수는 `openVerifiedDescriptor`만
부른다. 뮤테이션 M16(회수를 `readDescriptor`로 되돌림)이 이 테스트로 죽는다.

B3이 실제로 발동하는 곳은 `Dial` 경로다 — 소비자는 내용까지 봐야 하므로 잘린 JSON을
받으면 거부하고, 호출부가 그것을 dormant 강등으로 흡수한다(design D4-2).

## B1이 두 방향으로 쓰인다

- `TestStartRefusesUnsafeLeftoverShapes` — 형식이 틀린 잔재(0644 descriptor)는 회수도
  거부한다. 완화된 것은 내용이지 형식이 아니다.
- `TestDescriptorPublicationIsAtomic` — 재발행 중의 `errDescriptorChanged`는 손상이
  아니라 rename 원자성의 관측이다. 그 구분이 없으면 정상 경합이 손상으로 읽힌다.

## 미테스트 분기에 대하여

B2·B4·B5는 이 change가 **바꾸지 않은 줄**이고 편집 전에도 무테스트였다. 커버리지
개선은 이 change의 scope가 아니라 **범위 밖**이다 — `not-applicable`이 아니다.
