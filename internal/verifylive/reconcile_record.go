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

// readRecordStrictNoTail 은 개행 꼬리 검사 없이 읽고 엄격 해독함(추가 직전: 해독 → 지문 → 개행 순서용).
func readRecordStrictNoTail(path string) ([]byte, []Entry, error) {
	raw, err := readRecordRaw(path)
	if err != nil {
		return nil, nil, refuse(RefuseRecordUnreadable, "the record cannot be read: "+err.Error())
	}
	var entries []Entry
	for n, line := range bytes.Split(raw, []byte{'\n'}) {
		if len(bytes.TrimSpace(line)) == 0 {
			continue
		}
		var e Entry
		if err := json.Unmarshal(line, &e); err != nil {
			return nil, nil, refuse(RefuseRecordUndecodable, fmt.Sprintf("record line %d does not decode: %v", n+1, err))
		}
		if e.FormatVersion > RecordFormatVersion {
			return nil, nil, refuse(RefuseRecordFormat,
				fmt.Sprintf("record line %d is format %d; this build understands %d", n+1, e.FormatVersion, RecordFormatVersion))
		}
		entries = append(entries, e)
	}
	return raw, entries, nil
}

// appendReconcileLine 은 대사 줄 하나를 Recorder 로 추가함 — Calls 없음, 근거는 관측 하나, 계좌는 기록의 마스크.
func appendReconcileLine(path, accountMask string, art Artifact, basis string, start, at time.Time) error {
	rec, err := OpenRecorder(path)
	if err != nil {
		return err
	}
	defer rec.Close()
	runID, err := reconcileToken("reconcile")
	if err != nil {
		return err
	}
	return rec.Append(Entry{
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
