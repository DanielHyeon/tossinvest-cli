package verifylive

// account_seal.go — a131: 검증 기록의 봉인된 계좌(마스크)를 읽는 쪽에서 엶(design D4). 봉인은 Recorder.Append ·
// lockedRecord.appendLine 의 redact.SealRecordLine.

import (
	"errors"

	"github.com/JungHoonGhae/tossinvest-cli/internal/redact"
)

// errSealUnopenable 는 줄의 봉인 계좌를 이 기록의 키로 열 수 없음임. 해독 오류로 다룸(빈 계좌로 읽으면 그 줄이 연속일수·endpoint
// 증거에 섞여 다른 줄의 계좌로 셈해짐 — a131 리뷰 R-2). 마지막 줄 찢김 관용도 받지 않음.
var errSealUnopenable = errors.New("the line's sealed account cannot be opened with this record's seal key")

// openEntryAccount 는 봉인된 계좌가 있으면 열어 AccountRef 를 채움. 열 수 없으면 오류. 평문만 있는 a131 이전 줄은 그대로 둠(전환기 호환).
func openEntryAccount(e Entry, opener *redact.Opener) (Entry, error) {
	if e.AccountRefSealed == "" {
		return e, nil
	}
	ref, ok := opener.Open(e.AccountRefSealed)
	if !ok {
		return e, errSealUnopenable
	}
	e.AccountRef = ref
	// 봉인 값은 저장 형식일 뿐 — 메모리에서는 원래 값만 남김(다시 직렬화하면 같은 규칙이 새로 봉인함).
	e.AccountRefSealed = ""
	return e, nil
}
