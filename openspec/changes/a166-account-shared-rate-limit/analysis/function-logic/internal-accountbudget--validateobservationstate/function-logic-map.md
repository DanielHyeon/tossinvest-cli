# validateObservationState FLM — retained gap/Last

후보 22784a1 이전 AST는 `analysis/429-store-lot/gap-last-validation-pre.json`, 이후 ast.json. Version·scope/instance·pending·ring 크기 및 정확 잔존 개수·counter를 검사하고, event와 gap 각각 sequence/attempt/종류·순서를 검사한다. event와 gap attempt 집합을 축적하여 중복 및 Last.NoResponse/Response와의 모순을 거절한다. Last.RateLimited는 최신 event와 일치해야 하며 slot key/poison/hint/anchor는 별도 검증한다. 유효 상태는 nil, 모순은 error로 read/Snapshot failclosed. 삭제된 ring 밖 이력은 증명하지 않는다. 기존 호출·파일 형식·API 변경 없음.
