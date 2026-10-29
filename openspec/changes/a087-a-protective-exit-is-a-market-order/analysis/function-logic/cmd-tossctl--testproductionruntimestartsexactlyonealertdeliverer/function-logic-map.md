# Function Logic Map: `TestProductionRuntimeStartsExactlyOneAlertDeliverer`

- Source: `cmd/tossctl/a098_the_engine_registers_a_sender_test.go:18-31`
- Pre-Edit HEAD: `20ef25ef5e64fa5ff179c62fbeed7c3d5f176052`
- AST evidence: 기존 테스트 함수 단일 runtime 조립·집합 동일성 비교

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| runtime auxiliary 이름 | 생산 조립의 전체 집합 | `engineRuntime` | 누락·중복·무단 추가 시 실패함 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `engineRuntime` 오류 | 없음 | `t.Fatalf` | 기존 runtime 오류 시험 |
| B2 | 이름 집합 불일치 | 없음 | `t.Fatalf` | 이 함수 |
| B3 | 전체 일치 | 없음 | 통과 | 이 함수 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `engineRuntime` | 실제 생산 보조 실행자 집합 조립 | 오류면 B1 | `ast.json` |
| `AuxiliaryNames` | 전체 이름 집합 읽기 | 오류 없음 | B2 |

## State mutations and fallbacks

- 시험은 상태를 바꾸지 않고 runtime 조립 결과만 읽음.
- admission nil 보조 실행자는 취소 전 대기하므로 이 집합 포함이 broker GET을 뜻하지 않음.

## Safety conclusion

- 하한가 갱신기는 감독 루프가 아닌 휴면 보조 실행자 하나로만 추가함.
- 이름 집합 비교를 넓혀 무단 보조 실행자·중복 등록을 막음.
