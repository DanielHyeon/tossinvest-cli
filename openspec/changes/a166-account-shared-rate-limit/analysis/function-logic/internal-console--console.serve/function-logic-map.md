# Function Logic Map: internal-console--console.serve

기준 bbe5066ec; ast.json은 수정 전 source SHA. CodeGraph 격리 worktree 인덱스 없음, 원본 손읽기.

listener 경계 확인 후 Serve. 기존 server 오류는 정리 전 반환, signal/relaunch 시 settle→5초 Shutdown→relaunch. 변경: 모든 종료 경로에서 shutdown과 worker 실제 done을 확인하고 오류 결합; handler quiesce 선행으로 새 run 경합 차단.
