# offline 도구 수동 검토

기준 cf0f0919e; 사전 분기·시험 지도 pre-edit.md, 독립 설계 검토 independent-design-review.md. 기존 함수 본문은 수정하지 않았다. 기존 core Provision/OpenContext/ProvisionObservation, runtime manifest/Open 및 SystemBootSource, stdlib와 설치된 x/sys를 재사용했다. 신규 의존성 없음; impeccable N/A — CLI/backend-only.

수동 gstack checklist 범위: 경로·소유권·실패 후 파일 보존, context 및 identity 재검사, atomically no-replace 게시·fsync 순서·cleanup 오류, check의 clock/Factory Close와 HTTP/client 미호출을 검토했다. 자동 specialist·업데이트·telemetry는 실행하지 않았고 독립 검증을 대체하지 않는다. 초기 컴파일 RED 이후 provision/check GREEN, root 교체 재검사·RENAME_NOREPLACE·boot snapshot 제거 변이 3개 RED 원문 보존. 변이는 helper 추출 전 동일 핵심 본문에서 수행했으며 합성 오류는 실제 디스크/전원 장애가 아니다.

Go1.25 최종 정확 선택 `go test ./internal/accountbudgetruntime ./tools/a166-budget-state -run '^TestA166Offline' -count=1 -v` normal 0.344초/0.051초, `-race -tags tossos_testseams` 1.385초/1.064초. freeze3 원문은 최종 fixture 전체를 포함한다. 모든 실행은 user/net namespace lo-only/no-route, 임시 root만 사용한다. 파일 source SHA는 fixture-sha256.txt. owner와 unsupported filesystem은 합성 stat/magic값 시험이고 실제 타 UID·다른 filesystem의 실측은 아니다.

check 성공은 local manifest/state/boot 조건 확인만 뜻한다. 다중 namespace 동등성, writer quiescence·30초 정지·배포 권한·실제 운영 경로, powerloss, broker/credential 읽기, 전체 tests/gate는 미검증이다. 실패한 부분 root와 이미 게시된 manifest는 자동 삭제하지 않는다. 동일 UID 악의적 경로 교체는 core API의 지원 밖이며 재검사로 예방 가능한 교체만 거절한다.
