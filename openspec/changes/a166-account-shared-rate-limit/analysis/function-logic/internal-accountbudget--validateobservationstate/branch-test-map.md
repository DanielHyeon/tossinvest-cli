# validateObservationState BTM — retained gap/Last

canonical JSON에 Attempt=2, Gap=1, gap Attempt=2, Last Attempt=2/NoResponse를 넣은 `TestA166ObservationRejectsRetainedGapAsNoResponse`는 baseline RED→GREEN 및 한 줄 제거 변이 RED. 정상 no-response, gap ring, event/slot 한도, child crash, 저장 fault는 `gap-last-normal.log`/`gap-last-race.log`의 정확 선택 범위에서 검증한다. ring 밖 이전 이력·전체 gate는 미검증.
