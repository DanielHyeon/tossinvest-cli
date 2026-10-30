# 후속 배선 계약 — 하나의 보수적 예산 그룹 (Manager 승인, 2026-10-31)

## 발견과 최소 처분

`official/reads.go`는 accountSeq를 호출 식별자로, accountNo를 표시값으로 사용한다. `VerifyAuthoritativeAccountIdentity`는 같은 클라이언트의 선택 seq가 계좌 목록에 정확히 한 번 나오는지만 확인한다. 서로 다른 앱 키의 seq가 같은 물리 계좌인지 증명하지 않는다. token/accounts 요청 자체에는 선택 계좌 header가 없다. 기존 engine·verify·verify-reconcile의 선택 규칙도 서로 다르므로 이를 새 추정으로 통일하지 않는다.

첫 배선에서는 키나 seq로 예산을 분할하지 않고, **해당 설치의 모든 참여 profile/계좌를 명시적인 한 예산 그룹으로 묶는 안**을 사용한다. 그 그룹의 모든 요청 합계가 rolling 1초 3개 이하이면 각 계좌의 요청은 그 부분집합이므로 3개 이하이다. 계좌 동등성을 추정할 필요가 없고 인증 bootstrap도 첫 요청부터 같은 예산을 쓴다. 여러 계좌를 함께 쓰면 합산 처리량도 3개 이하로 제한되는 비용이 있다. 계좌마다 독립 3개를 제공한다고 설명하지 않는다.

이 안은 운영 활성화가 아니다. 승인된 strict 상한·보호 대기 정책을 구현하는 보수적 지원 범위이며 독립 설계 검토의 P1 신뢰 기준·P2 객체/프로세스 표현을 보완해 canonical design/spec/tasks에 반영했다. 실제 참여자 범위·group 구성은 배포 전 사람 확인 항목이다.

## 배선 계약

1. required 모드는 root·opaque group scope·instance와 **신뢰된 공통 launch manifest의 예상 root/queue.lock/wire.lock device·inode**를 명시적으로 받는다. 이 manifest는 운영자가 모든 참여자의 launch 설정에 같은 값으로 고정하는 신뢰 기준이며, 열려는 root에서 읽어 그 자리에서 스스로 승인하는 값이 아니다. 필요한 입력 누락/오류·실제 파일 ID 불일치는 bootstrap 네트워크 전 거절한다. manifest 자체를 임의 교체하는 관리자의 행위까지 막는다고 주장하지 않는다. 모든 참여 엔진/CLI/도구/프로필은 같은 group을 사용한다. config-dir·app key·accountSeq별 root 생성, 자동 초기화, 누락 시 OFF 복귀는 없다. OFF는 별도로 명시된 비적용 상태다.
2. 공통 factory가 trusted clock.System 계열로 기존 group을 열어 예상 파일 ID와 비교하고, `NewWithAccountBudget`를 만든 뒤 token/accounts를 호출한다. 같은 프로세스에서 재생성하는 client·a112 clone·replay/probe는 Budget 객체를 재사용한다. 다른 프로세스나 독립 객체는 같은 manifest의 device/inode/instance를 검증한다. Go 객체를 프로세스 사이에 공유한다고 설명하지 않는다. 새 클라이언트에 별도 local bucket을 만들지 않는다.
3. 기존 계좌 선택·권한·원장 계좌 확인은 보존한다. 이 확인을 cross-key 동등성 증거로 쓰지 않는다. 여러 계좌를 반환한 credentials도 첫 bootstrap부터 그룹 총량 안에 있으므로 기존 선택 계약을 임의로 단일 계좌로 제한할 필요가 없다.
4. 다른 root/instance에 나뉜 참여자, 다른 호스트, bare OFF 도구나 외부 앱의 동시 호출은 이 보장의 범위 밖이다. 운영 provisioning/manifest 검사는 모든 알려진 참여자·launch 경로를 열거하고 같은 namespace/device/inode를 검증해야 한다. 확인되지 않으면 운영 적용을 승인하지 않는다. 같은 marker 문자열만 복사한 파일은 같은 예산이라는 증거가 아니다.
5. 그룹 ID에 계좌번호·키·토큰을 넣지 않는다. 새 credential fingerprint·계좌번호 HMAC·별도 계좌 매핑 데이터베이스는 이 방식에서 필요 없으므로 추가하지 않는다. 기존 안전한 파일/instance 계약, runtime mode 및 위 공통 manifest의 예상 파일 ID를 사용한다.
6. critical 우선순위와 이미 선택된 normal의 비선점, auxiliary try-admit, 기존 context/deadline을 유지한다. 그룹에 속한 다른 계좌 요청도 대기를 늘릴 수 있으며 지연 0/유한 wall-time 상한은 주장하지 않는다.

## 수용 증거

- 서로 다른 fake credentials와 서로 다른 계좌 목록을 쓰는 두 실제 official.Client가 같은 group을 통해 token/accounts/계좌별 요청을 보내며 합계 start 상한을 만족한다.
- profile/config-dir이 달라도 같은 root/instance를 사용하고 required 누락/foreign group은 네트워크 전에 거절된다. 같은 marker를 복사한 다른 디렉터리도 공통 manifest의 device/inode와 달라 거절해야 한다. 두 프로세스가 같은 manifest를 사용한 fixture를 포함한다. 알려진 모든 생산 생성자와 별도 전송은 factory 또는 명시 guard를 사용한다.
- 계좌 선택의 기존 engine/verify/reconcile 차이, OFF 동작, 원장 판정은 보존한다. 임의 보호 무대기 우회·미확정 downgrade는 없다.
- 운영 namespace/참여자 확인과 실제 배포는 fixture 통과와 구별한다. 각 계좌의 브로커 도착률·외부 비참여 트래픽까지 보장하지 않는다.

search-first: 기존 계좌 API·선택/authority 함수와 이미 구현된 opaque scope/instance를 재사용한다. 새로운 계좌 동등성 알고리즘·의존성을 만들지 않는다. grill-me/grilling: 사용자 정책을 재질문하지 않고 지원 조건·처리량 비용·운영 전 확인을 분리한다. impeccable: N/A — backend/documentation-only. 새 코드·시험·실계좌 조회 없음.

독립 Terra는 공통 manifest 보완 뒤 P1/P2 해소와 canonical 정합성을 확인했다. scoped OpenSpec strict·diff 검사는 통과했다. 401 문장의 선행 Write는 token이 아니라 동일 주문 attempt의 선행 주문 본체로 한정하는 마지막 P2도 반영했다. Manager가 source 원문과 이 처분을 확인해 후속 구현을 승인한다. fixture 실행·gstack 코드 검토·실제 manifest 검증은 후속 로트이며 이 문서 PASS로 대신하지 않는다.
