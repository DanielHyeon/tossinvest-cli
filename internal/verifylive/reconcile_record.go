package verifylive

// reconcile_record.go — a121 대사의 기록 쪽: 엄격 해독(F3·R2-3), 대사 줄 추가(Recorder 경유 1회), 근거 지문,
// 대사 줄의 직렬화 모양, 보고 투영.

import (
	"bytes"
	"crypto/rand"
	"encoding/base32"
	"encoding/json"
	"fmt"
	"time"
)

// readRecordStrict 는 기록을 한 번 읽어 모든 비공백 줄을 엄격 해독하고 개행 꼬리를 봄(선택 전 사전 검사).
// LoadEntries 와 달리 해독 불능 마지막 줄을 버리지 않음.
func readRecordStrict(path string) ([]byte, []Entry, error) {
	raw, entries, err := readRecordStrictNoTail(path)
	if err != nil {
		return nil, nil, err
	}
	if len(raw) > 0 && raw[len(raw)-1] != '\n' {
		return nil, nil, refuse(RefuseRecordTailNoNewline,
			"the record does not end with a newline — an append would fuse with its last line")
	}
	return raw, entries, nil
}

// readRecordStrictNoTail 은 개행 꼬리 검사 없이 읽고 엄격 해독함.
func readRecordStrictNoTail(path string) ([]byte, []Entry, error) {
	raw, err := readRecordRaw(path)
	if err != nil {
		return nil, nil, refuse(RefuseRecordUnreadable, "the record cannot be read: "+err.Error())
	}
	entries, err := decodeRecordStrict(raw)
	if err != nil {
		return nil, nil, err
	}
	return raw, entries, nil
}

// decodeRecordStrict 는 모든 비공백 줄을 엄격 해독함 — LoadEntries 와 달리 해독 불능 마지막 줄을 버리지 않음.
func decodeRecordStrict(raw []byte) ([]Entry, error) {
	var entries []Entry
	for n, line := range bytes.Split(raw, []byte{'\n'}) {
		if len(bytes.TrimSpace(line)) == 0 {
			continue
		}
		var e Entry
		if err := json.Unmarshal(line, &e); err != nil {
			return nil, refuse(RefuseRecordUndecodable, fmt.Sprintf("record line %d does not decode: %v", n+1, err))
		}
		if e.FormatVersion > RecordFormatVersion {
			return nil, refuse(RefuseRecordFormat,
				fmt.Sprintf("record line %d is format %d; this build understands %d", n+1, e.FormatVersion, RecordFormatVersion))
		}
		entries = append(entries, e)
	}
	return entries, nil
}

// encodeReconcileLine 은 대사 줄 하나를 개행까지 직렬화함 — Calls 없음, 근거는 관측 하나, 계좌는 기록의 마스크.
func encodeReconcileLine(accountMask string, art Artifact, basis string, start, at time.Time) ([]byte, error) {
	runID, err := reconcileToken("reconcile")
	if err != nil {
		return nil, err
	}
	b, err := json.Marshal(Entry{
		FormatVersion: RecordFormatVersion,
		Kind:          KindReconcile,
		RunID:         runID,
		StepID:        StepReconcile,
		Title:         "reconciled absent",
		StartedAt:     start.UTC(),
		FinishedAt:    at.UTC(),
		AccountRef:    accountMask,
		Observations: []Observation{{Key: ObservationReconcileBasis, Value: basis,
			Detail: "two complete official list reads narrowed the absence window; not a cancel, fill or endpoint proof"}},
		Artifacts: []Artifact{art},
	})
	if err != nil {
		return nil, fmt.Errorf("verify reconcile: encoding the reconcile line: %w", err)
	}
	return append(b, '\n'), nil
}

func reconcileToken(prefix string) (string, error) {
	var buf [10]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return "", fmt.Errorf("verify reconcile: crypto/rand is unavailable: %w", err)
	}
	return prefix + "-" + base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(buf[:]), nil
}

// ReconcileBasisDigest 는 읽기 근거를 버전·도메인 태그와 함께 정규 순서로 지문화함 — 순서 무관, multiset(중복 유지).
func ReconcileBasisDigest(rows []ReconcileBasisRow) string {
	return ReconcileBasisDomain + ":" + Digest(struct {
		Domain string              `json:"domain"`
		Rows   []ReconcileBasisRow `json:"rows"`
	}{Domain: ReconcileBasisDomain, Rows: sortedReconcileRows(rows)})
}

// MarshalJSON 은 대사된 artifact 만 design 확정형(P2-6)으로 씀 — 취소·체결·보유 필드(영 시각 포함)를 싣지 않음.
// 대사되지 않은 artifact 는 별칭 타입으로 기존과 바이트 단위로 같게 씀(reconcile_marshal_test.go 가 핀).
func (a Artifact) MarshalJSON() ([]byte, error) {
	type plainArtifact Artifact
	if !a.ReconciledAbsent {
		return json.Marshal(plainArtifact(a))
	}
	return json.Marshal(struct {
		Kind             string    `json:"kind"`
		ID               string    `json:"id"`
		Symbol           string    `json:"symbol"`
		CreatedAt        time.Time `json:"created_at"`
		Cancelled        bool      `json:"cancelled"`
		Deliberate       bool      `json:"deliberate,omitempty"`
		ChainID          string    `json:"chain_id,omitempty"`
		Note             string    `json:"note,omitempty"`
		ReconciledAbsent bool      `json:"reconciled_absent"`
		ReconciledAt     time.Time `json:"reconciled_at,omitzero"`
	}{a.Kind, a.ID, a.Symbol, a.CreatedAt, false, a.Deliberate, a.ChainID, a.Note, true, a.ReconciledAt})
}

// --- 기록 잠금·추가의 오류 표지(record.go 의 lockedRecord 가 돌려주고 대사가 가름) ----------------

// errRecordLocked 는 다른 대사가 같은 기록 파일(같은 inode — 하드링크 별칭 포함)을 잡고 있음을 뜻함.
var errRecordLocked = fmt.Errorf("another reconciliation holds this record file")

// errRecordAppendIncomplete 는 대사 줄 쓰기가 끝까지 가지 못했음을 뜻함 — 기록 끝에 찢긴 줄이 남았을 수 있음.
var errRecordAppendIncomplete = fmt.Errorf("the reconciliation line was not completely written; the record may now end in a torn line — " +
	"inspect its tail before any `tossctl verify run --resume`")

// errRecordAppendUnverified 는 쓴 뒤 다시 읽은 바이트가 쓴 줄과 다름을 뜻함(교차 쓰기 등).
var errRecordAppendUnverified = fmt.Errorf("the reconciliation line read back differently from what was written; " +
	"another writer may have interleaved — inspect the record before any further run")

// errRecordSizeMoved 는 잠근 뒤 판정한 길이와 쓰기 직전 파일 길이가 다름 — 잠그지 않는 작성자가 끼어들었음.
var errRecordSizeMoved = fmt.Errorf("the record grew between the locked read and the append")
