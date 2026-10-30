# a166 4판 Manager 검토 — 2026-10-31

## 범위와 판정

사용자 확정 정책은 계좌 공유 3건/초 우선·보호 대기, 이후 개발 진행·검증 뒤 종결이다. Manager는 OpenSpec 작성·검토만 수행하고 모든 코딩·시험을 Terra에 위임했다. 운영·실계좌·배포·전체 시험·전체 gate는 수행하지 않았다.

현재 판정은 **설계 freeze 조건부 PASS**다. 현재 코드의 전체 계좌 상한 또는 실제 활성화 PASS가 아니다. 연구 핵심은 제품 구현의 기술 선택 근거로 수락하며, 제품 RED/GREEN·기존 함수의 사전 FLM/BTM·배선·권한·환경 검증을 건너뛰지 않는다.

## 원문 검증

- TLS 연구 `d0485309a`: 실제 `official.Client` 로컬 fixture의 token→401→token→200, 정상 분할 body, request/response 취소, TLS 검증, 다중 프로세스 동일 flock, 소유자 종료, full-response 소유권을 시험했다. Manager는 `tls_test.go`의 Write/revoke/owner 수명과 source diff를 읽었다. 연구 측정점은 plaintext 첫 Write이며 서버 도착률은 아니다.
- 독립 Terra는 핵심 `b90524af6`을 Go 1.26.5 `-race` 20.537초, 지원 버전 Go 1.25.0 20.598초로 재실행했다. `1b5f09589`의 새 responseOwner 동시 Read/Close/cancel 및 Close 오류 계수 시험은 Go 1.25.0에서 1.014초 통과했다. 최종 `d0485309a`는 Linux build constraints와 기록 변경이며 별도 diff 확인했다. Alpine 이미지에서 실행한 증거는 아니다.
- 큐 연구 `41c8c16a7`: priority·liveness·취소·dirty·fault 주입·aux 정리를 확인했다. Manager는 445419bc1→최신 diff와 원문을 읽었고, 독립 Terra의 최신 소형 패키지 `-race`는 6.152초 통과했다. fsync 오류는 callback 주입이며 실제 kernel EIO·전원 장애·reboot 실측이 아니다.
- RED/변이에는 응답 조기 release, revoke 생략, 취소 오류 무시, dirty 무시, aux dead 정리 생략, 불확실한 state write 뒤 liveness 삭제가 포함된다. 일부 초기 연구 실패 원문을 덮어쓴 한계는 해당 README에 그대로 남겼으며 재현한 척 보완하지 않았다.
- source/AST: 기준 e32606d4b의 기존 a166 AST 23개 중 21개 파일 hash 일치, 2개는 재추출했다. 파일 hash 변화와 의미 변화는 구별한다. CodeGraph는 up to date를 표시했지만 commit SHA를 출력하지 않아 원본 source/AST와 교차 확인했다. 쿼리를 완전한 전체 호출자 증명으로 보고하지 않는다.

## 독립 설계 리뷰 처분

독립 Terra가 4판 원문을 검토했다. 이는 문서 손 읽기이며 source 분기 AST 판정이 아니다.

- P1: dirty 제거 후 마지막 directory sync 실패와 state rename 내구화 실패가 spec에서 합쳐져 있었다. 두 scenario로 분리하여 design과 일치시켰다. 재검토에서 해소됐다.
- P3: 「모든 429 이벤트 및 cooldown」을 「429 이벤트·해당 RequiredQuery의 유효 hint 제한·350ms 간격」으로 바꿔 새 계좌 전역 패널티로 읽히지 않게 했다.
- 일반 유한 wall-time 보장을 하지 않는다는 한계를 승인된 보호 대기 정책에 대한 새 blocker로 취급하지 않았다. 지원하는 표준 HTTP/1 TLS 경로의 구체 위반은 제품 검증에서 별도로 찾는다.
- 큐 Register 실패의 파일 정리는 확정 미공개와 공개 불확실을 분리했다. 불확실할 때 liveness를 남기는 보수 동작을 무조건 orphan 결함으로 바꾸지 않았다. aux는 죽은 ticket을 공유 절차로 정리한다.

## 연구 선택과 skill 적용

search-first: 기존 `internal/enginelock`, `internal/config` flock, `internal/officialfx/production_owner_unix.go` 및 production 파일 갱신 선례, Go 표준 `net/http`/`crypto/tls`/context를 우선했다. 기존 `golang.org/x/sys v0.46.0`은 go.mod와 로컬 module LICENSE를 확인했다. Go·x/sys의 BSD 계열 라이선스 원문을 확인했으며 런타임 의존성 추가는 없다. 프로세스별 rate limiter는 합산 상한을 해결하지 않아 채택하지 않는다. 기존 native 잠금과 stdlib를 확장하는 안을 택한다.

공식 문서: https://pkg.go.dev/net/http , https://pkg.go.dev/net/http/httptrace . 연구는 설치된 Go 1.26.5 source의 transport retry와 TLS close를 손 읽기로 대조했고 지원 Go 1.25.0으로 fixture를 재실행했다. 특정 Go 버전 소스와 표본의 한계를 일반 플랫폼 증명으로 확대하지 않는다. 새 패키지 도입이 필요하지 않아 외부 패키지 registry 비교는 수행하지 않았다.

grill-me/grilling: 확정 정책을 반복 질문하지 않고 scope·실패·요청 경계·수용 기준을 구분했다. impeccable: N/A — backend/documentation-only. ai-memory 프로젝트 CLI 검색은 빈 결과였고 과거 로컬 기억은 현재 source로 재검증했다. Hindsight/CodeGraphContext 신규 조회는 수행하지 못했으며 advisory 공백으로 남긴다. gstack은 각 연구·독립 검토의 체크리스트를 수동 적용했다. 외부 모델 CLI를 포함한 전체 다중 보이스 절차는 실행하지 않았고, 제품 High-risk 구현 리뷰를 대체하지 않는다.

OpenSpec 검증은 과거 프로젝트 사용 근거가 있는 1.4.1을 임시 전용 경로에 exact pin했다. 삭제 직전 설치 버전이 확인된 것은 아니다. npm package integrity·MIT·Node >=20.19.0 및 로컬 v22.14.0을 확인하고 install scripts/telemetry를 끈 채 scoped validate만 실행했다. 전역 링크와 repo dependency를 바꾸지 않았다. 전체 dependency CVE 감사는 하지 않았다.

## 남은 조건

최신 원문 증거를 현재 change에 보존하고, 연구 branch와 Manager 계약을 통합한 commit에서 scoped strict·PM·diff check를 통과한 뒤 Manager READY를 기록한다. 이후 미사용 내부 모듈 로트부터 제품 구현을 시작한다. namespace/계좌 scope/provisioning, clock provider와 hint 복원, authority/OFF·모든 호출자 배선은 각 제품 로트의 수용 시험이며 연구 완료로 체크하지 않는다. a087 admission nil, a169 리허설·측정 보류는 유지한다.
