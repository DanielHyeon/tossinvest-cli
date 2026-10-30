# Function Logic Map: internal-console--console.settle

기준 bbe5066ec; ast.json은 수정 전 source SHA. CodeGraph 격리 worktree 인덱스 없음, 원본 손읽기.

현재 run nil 또는 finishedFlag면 반환, 그 외 취소 후 최대 2분 대기. finishedFlag는 실제 done보다 먼저 세워지므로 done 닫힘만 완료다. timeout은 오류.
