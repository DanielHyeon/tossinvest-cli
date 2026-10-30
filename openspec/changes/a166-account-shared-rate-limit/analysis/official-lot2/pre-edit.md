# Lot2a Pre-Edit / FLM / BTM

Base `da67f3e1e`, prep `6bc9cee2b`; lot1 draft `abaa22aaa` merged for API reference only, final lot1 integration required. No runtime activation or production caller wiring in this lot. Existing AST pre snapshots accompany this record.

| Function | Current branches and invariant | RED |
|---|---|---|
| New | options applied before configuration sealed; tokenManager and client share hc. | OFF unchanged; enforcing constructor rejects nil and unsupported configuration; same budget for token. |
| hcTransport | nil client false, only direct *http.Transport accepted. | exact constructor-created wrapper preserves authority, arbitrary client/base/option replay never promotes. |
| doRequest | transport error observes raw error then wraps ErrTransport; body error distinct. | enforcing typed prewrite survives wrapping; OFF string and errors.Is unchanged; post-write ambiguous. |
| a112MBUSReadAt | deadline/client/authority checks precede cloned transport and one Do. | enforcing clone uses same budget without changing response/redirect/authority rules. |

Additional existing send/token branches and a165 409 are in `analysis/official-lot2-prep/pre-edit.md` and source-hash-matched AST evidence. Final source modifications determine refreshed FLM/BTM and test scope; no whole test/gate or live operation.

## 구현·검증 기록 — lot2a 한정

Manager 승인 생성자 `NewWithAccountBudget`는 기존 OFF `New`를 유지하고 예산 또는 cleanup 누락·지원 밖 transport를 nil client와 오류로 거절한다. caller의 `http.Client`를 복사하고 새 wrapper를 token manager와 공유한다. private wrapper pointer 및 기존 authorityOrigin·constructor transport 결속을 검사한다. a112 clone도 같은 budget으로 감싼다. enforcing 요청의 기존 ErrTransport 및 원 오류 identity를 함께 보존한다. execgw 분류 코드는 바꾸지 않는다.

core draft `abaa22aaa`에서 시작해 후보 `dc149bc51`, 수정 `cbf53d452`를 병합했다. 입력 transport `Clone`의 lazy HTTP/2 초기화로 a112와 caller alias RED가 실패했고 core 수정 뒤 두 RED가 통과했다. a112 시험은 기존 정적 guard가 허용하는 `a112_mbus_*_test.go`에 위치하며 guard를 수정하지 않았다. local fake dial은 실제 외부 접속을 차단한다. 실제 공식 Client의 local TLS token·401·redirect, token과 주문 prewrite identity, 401 선행 전송 추적, OFF error identity, authority/option replay, a112 같은 budget, a165 409 뒤 재전송 NOT_SENT의 journal IN_DOUBT 경로를 제한적으로 시험했다. 세 변이(a112 wrapper 삭제, token hc binding 삭제, 오류 identity 삭제)는 각각 검출됐다.

테스트 원문은 이 디렉터리의 `a166-lot2a-*.txt`에 보존했다. 최초 normal 정규식의 official `[no tests to run]`은 미실행이며 corrected 파일만 통과 근거다. 중간 a112 guard 실패 원문 두 파일은 실수로 재실행 출력에 덮여 보존하지 못했다. 이후 실행별 고유 파일을 사용했다. 모든 테스트는 local fake 또는 임시 디렉터리이며 live broker 호출 없음.

수동 gstack 범위 검토: 권한/외부 override 경계, 오류 identity, token client alias, a112 clone, body와 기존 OFF 분기·사용자 입력/비밀 출력, 테스트 효과를 검토했다. 전체 mixed-model gstack 절차와 독립 검증은 미실행이다. replay/probe·전 생산 호출자 구조·429 이벤트 및 폴링은 후속 lot2b/3; tasks 2.1–2.5는 완료로 체크하지 않는다. 실제 wire 집계 및 운영 활성화는 검증하지 않았다.
