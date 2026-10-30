# a166 lot3a runtime factory — 설계·RED 준비

기준 `f44143788b068f19569ca2dc2cf46dbb299e34bc`. 후속 group-binding 문서 착지와 Manager 검토 후 그 SHA를 확인한다. 제품 callsite·설정·운영 변경 없음. 기존 `internal/accountbudget.OpenContext`는 scope·instance, root 및 queue/wire inode를 내부 pin하고 매 획득에서 재검증한다. `files_linux.go`의 openat NOFOLLOW·owner/mode·파일 유형/링크 수·ext4/tmpfs 제한을 재사용한다. manifest의 기대 ID는 예산 root 자체에서 읽어 자기 승인하지 않고 신뢰된 별도 launch 파일에서 받는다.

## 제안

OFF/required는 호출자의 명시 mode로 구분한다. env 편의 해석에서 변수 부재를 OFF로 해석하더라도 required 배포의 상위 launch 경로가 이 부재를 오류로 판정할 의무가 없으면 required 누락 검출이 불가능하다. required mode에서 manifest 미지정·빈 값은 오류다. 새로운 전역 registry나 hot reload 없이 프로세스 소유 factory 인스턴스 하나가 `clock.System`을 사용해 기존 group을 `OpenContext`로 연 뒤, 고정된 root/queue/wire device·inode와 외부 manifest 기대값을 비교한다. 검증 실패는 닫고 client 0이다. 모든 재생성 client는 factory의 동일 budget을 사용한다. factory는 Close와 client 생성의 동시성, 활성 소유자 있을 때 budget Close 오류를 보존한다. close 후 생성 금지. 실제 official client는 기존 `NewWithAccountBudget`를 사용한다.

manifest는 version·scope·instance·정규 절대 root 및 세 객체의 device/inode만 갖는 크기 제한 strict JSON이다. symlink, 잘못된 owner/mode, trailing/unknown 필드, 변조·읽기 실패는 fail-closed. 모호한 path·파일 교체 검사는 안전한 fd 기반 열기와 열린 Budget이 pin한 원본 식별자의 비교를 사용한다. profile/계좌별 새 group 생성·자동 provisioning·키 동등성 추정은 없다. 프로세스 간 Go 객체를 공유한다고 주장하지 않는다.

## RED 계획

1. OFF vs required 미설정·빈 manifest, 잘못된 schema·버전·범위·symlink·owner/mode·크기·후행 JSON은 모두 client 생성 전 거절. OFF의 기존 official.New 동작은 무변경.
2. 정상 required의 여러 client가 같은 budget으로 token/accounts와 각 호출을 공유하고, local fixture가 client binding과 time window를 관찰한다. 취소된 OpenContext와 종료 중 생성/Close 실패를 명시 검증.
3. 두 자식 프로세스가 같은 trusted manifest를 사용: 정상 root 수락, scope/instance 마커까지 동일한 복사 root는 root/queue/wire dev+inode 불일치로 bootstrap wire 0. 다른 expected inode 각 필드 변이도 거절.
4. manifest와 actual root에 같은 값을 각각 self-read하는 unsafe 변이, expected inode 비교 생략 변이, OFF fallback 변이를 RED로 검출한다. no credentials 또는 raw account data 사용.

## 검증·범위

새 leaf는 기존 함수 FLM N/A; 기존 `Budget` 신원 leaf를 수정한다면 사전 Go AST·FLM/BTM 및 변이 계획을 추가한다. search-first: stdlib JSON·기존 x/sys/unix 및 기존 accountbudget 검증/clock 사용. 기존 도구와 자료의 버전은 design/manager-review 기록을 재확인한다. grill-me/grilling: 사용자 strict3·보호 대기 결정 재질문 없음; 구현 전 신뢰된 launch mode·manifest 경계만 명확히 한다. impeccable N/A — backend/documentation-only. 이 준비는 구현·시험·운영 PASS가 아니다.

## 결과와 수동 검토 — 후보, 미승인

Manager는 `Mode`/manifest·borrowed Budget API를 승인했고 canonical group 계약 `649fce7e8c0d17c76d0deb5b4f11ae0730448558`을 병합했다. 새 leaf만 추가해 기존 함수 FLM/BTM N/A다. AST post JSON의 hash는 후보 source를 가리킨다. `ModeOff`는 1, `ModeRequired`는 2이며 zero/unknown은 오류다. 환경변수 미설정은 명시 OFF로 변환될 뿐, 실제 운영의 required 참여자 누락을 검출하지 않는다. 운영 launch가 모든 참여자를 열거·검증해야 한다.

`Factory`는 explicit mode와 단일 Budget을 보유한다. `Budget()`은 신뢰된 내부 caller의 대여 포인터이며 닫기는 factory만 소유한다. OFF는 nil budget, 닫힘·불완전 factory는 오류다. NewClient·Budget·Close는 mutex로 직렬화하고 활성 lease에 대한 Close 오류를 보존한다. 이 포인터와 identity 확인은 금전 authority·상태 readiness·실제 lease 또는 다른 프로세스의 동등성을 단독으로 증명하지 않는다.

manifest는 root 외부의 신뢰된 launch 파일이어야 한다. 정규 절대 경로와 NOFOLLOW 각 부모, 파일 0600·소유자·단일 링크·최대 4096 bytes, strict 대소문자/중복/알 수 없는 키·후행 JSON 거절을 적용한다. 문자열 관계와 열린 부모들의 dev/inode를 기대 root와 대조한다. 관리자에 의한 manifest 교체, 예산 root 하위 디렉터리를 다른 mount에 bind하는 적대적 alias까지 이 비교로 막는다고 주장하지 않는다. 실제 namespace와 참여자는 운영 승인 전 확인 대상이다.

실제 자식 프로세스가 같은 manifest를 읽어 원래 root를 수락한 뒤 그 root 경로에 동일 marker의 복제 root를 배치하면 두 번째 자식은 inode 불일치로 거절한다. 두 실제 official.Client와 서로 다른 fake credentials·계좌 seq를 local TLS fixture에 연결했고 보유 중인 예산 lease가 token bootstrap을 wire 0으로 막는 것을 확인했다. 두 client가 동일 Budget 객체를 쓰는 것은 검사했으나 이 fixture는 transport 내부 첫 plaintext Write 시각을 직접 계측하지 않는다. 그룹 rolling 시작률의 end-to-end 직접 증거는 후속 로트에서 필요하며 core 단위 증거를 그 대체물로 과장하지 않는다.

선택 normal·race·tagged race, 기존 accountbudget 선택 race, 다른 root/각 ID·invalid mode·symlink·strict schema·active close·동시 생성 및 오류 경로 시험은 동봉 원문에 기록했다. identity 검증 제거, 빈 env OFF 허용, manifest self-approval 허용, zero mode OFF 허용 변이는 각각 검출됐다. Darwin 대상은 빌드만 확인했고 `[no test files]`이므로 실행 PASS가 아니다. Local TLS와 임시 디렉터리 외 운영/브로커 호출은 없다.

수동 gstack 범위 검토: required 누락·권한·symlink·identity·lock lifetime·오류·자격증명 노출·OFF 전이·테스트 변이를 점검했다. 전체 mixed-model gstack·독립 리뷰·전체 시험/gate·실제 배포는 수행하지 않았다. 이 후보는 production callsites에 연결되지 않았고 a087/a169 활성화 권한이 아니다.
