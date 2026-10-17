package main

// receipt.go 는 전송 한 번의 영수증을 씀 — 성공·거절·불명 전부.
//
// 순서가 계약임: 전송 **전에** 키 이름의 파일을 O_EXCL 로 만들어 "sending" 을 적고, 답을 받은 뒤
// 같은 파일을 임시 파일 + rename 으로 덮어씀. 그래서
//   - 같은 키(= 같은 confirm token)로 두 번째 전송은 파일 생성에서 막힘(1회용 토큰).
//   - 전송 도중 프로세스가 죽어도 "sending" 영수증이 남아 결과 불명임을 알림.
//   - 영수증을 못 쓰는 상태면 아예 보내지 않음(기록 없는 실주문 금지).
//
// 싣지 않는 것: 토큰·API 키·시크릿·Authorization/계좌 헤더 값·계좌 목록·잔고. 보낸 헤더는 이름만 적음.

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

const receiptSchema = "a087-market-sell-receipt/v1"

// 영수증 상태.
const (
	stateSending  = "sending"   // 전송 직전 기록 — 이 상태로 남아 있으면 결과 불명
	stateAnswered = "answered"  // HTTP 응답을 받음
	stateNoAnswer = "no-answer" // 요청은 넘겼으나 응답 없음(타임아웃·연결 끊김)
)

type receipt struct {
	Schema         string          `json:"schema"`
	Tool           string          `json:"tool"`
	State          string          `json:"state"`
	Outcome        string          `json:"outcome,omitempty"`
	ClientOrderID  string          `json:"client_order_id"`
	ConfirmToken   string          `json:"confirm_token"`
	IssuedAtKST    string          `json:"issued_at_kst"`
	SentAtKST      string          `json:"sent_at_kst"`
	AnsweredAtKST  string          `json:"answered_at_kst,omitempty"`
	Session        sessionContext  `json:"session_context"`
	Request        requestRecord   `json:"request"`
	Response       *responseRecord `json:"response,omitempty"`
	TransportError string          `json:"transport_error,omitempty"`
	Guidance       string          `json:"guidance"`
	Note           string          `json:"note,omitempty"`
}

// sessionContext 는 5.2(세션 밖 응답) 판독용 시각 문맥임. 달력을 조회하지 않으므로 휴장일은 모름 —
// 그 한계를 basis 칸에 그대로 적음.
type sessionContext struct {
	KSTWeekday string `json:"kst_weekday"`
	ClockPhase string `json:"clock_phase"`
	Basis      string `json:"basis"`
}

type requestRecord struct {
	Method         string   `json:"method"`
	Path           string   `json:"path"`
	WireBody       string   `json:"wire_body"`
	WireBodySHA256 string   `json:"wire_body_sha256"`
	HeaderNames    []string `json:"header_names"`
}

type responseRecord struct {
	HTTPStatus        int             `json:"http_status"`
	Body              string          `json:"body"`
	BodySHA256        string          `json:"body_sha256"`
	BodyJSON          json.RawMessage `json:"body_json,omitempty"`
	ErrorCode         string          `json:"error_code,omitempty"`
	ErrorMessage      string          `json:"error_message,omitempty"`
	OrderID           string          `json:"order_id,omitempty"`
	ClientOrderIDEcho *string         `json:"client_order_id_echo,omitempty"`
	EchoMatchesSent   *bool           `json:"echo_matches_sent,omitempty"`
}

// KRX 정규장(09:00~15:30 KST) 을 벽시계로만 가른 값.
const (
	phaseRegular = "krx-regular-by-clock"
	phaseOutside = "outside-regular-by-clock"
	phaseWeekend = "weekend"
)

func sessionAt(t time.Time) sessionContext {
	local := t.In(kst)
	ctx := sessionContext{
		KSTWeekday: local.Weekday().String(),
		Basis:      "KST wall clock against KRX regular 09:00-15:30 only; no market calendar was read, so holidays are not detected",
	}
	minutes := local.Hour()*60 + local.Minute()
	switch {
	case local.Weekday() == time.Saturday || local.Weekday() == time.Sunday:
		ctx.ClockPhase = phaseWeekend
	case minutes >= 9*60 && minutes < 15*60+30:
		ctx.ClockPhase = phaseRegular
	default:
		ctx.ClockPhase = phaseOutside
	}
	return ctx
}

func kstStamp(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.In(kst).Format(time.RFC3339Nano)
}

func sha256Hex(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

// receiptPath 는 키 하나당 파일 하나임 — 키가 파일 이름이라 O_EXCL 이 1회 전송을 강제함.
func receiptPath(dir, clientOrderID string) string {
	return filepath.Join(dir, clientOrderID+".json")
}

// newPendingReceipt 는 전송 직전 영수증임.
func newPendingReceipt(order assembledOrder, token string, sentAt time.Time, headerNames []string, note string) receipt {
	return receipt{
		Schema:        receiptSchema,
		Tool:          "a087-market-sell-probe",
		State:         stateSending,
		ClientOrderID: order.Spec.ClientOrderID,
		ConfirmToken:  token,
		IssuedAtKST:   kstStamp(order.IssuedAt),
		SentAtKST:     kstStamp(sentAt),
		Session:       sessionAt(sentAt),
		Request: requestRecord{
			Method:         "POST",
			Path:           orderPath,
			WireBody:       order.WireBody,
			WireBodySHA256: sha256Hex([]byte(order.WireBody)),
			HeaderNames:    headerNames,
		},
		Guidance: "the process stopped after handing the request over; the result is UNKNOWN — " +
			"check `tossctl orders list` / `tossctl orders completed` and do not resend",
		Note: note,
	}
}

// completeReceipt 는 전송 결과를 영수증에 채움.
func completeReceipt(pending receipt, result sendResult, sentKey string) receipt {
	r := pending
	r.SentAtKST = kstStamp(result.SentAt)
	r.AnsweredAtKST = kstStamp(result.AnsweredAt)
	r.TransportError = result.TransportError
	answer := readAnswer(result.Body, sentKey)
	if result.HTTPStatus != 0 {
		r.State = stateAnswered
		record := &responseRecord{
			HTTPStatus:        result.HTTPStatus,
			Body:              string(result.Body),
			BodySHA256:        sha256Hex(result.Body),
			ErrorCode:         answer.ErrorCode,
			ErrorMessage:      answer.ErrorMessage,
			OrderID:           answer.OrderID,
			ClientOrderIDEcho: answer.ClientOrderEcho,
			EchoMatchesSent:   answer.EchoMatchesSent,
		}
		if json.Valid(result.Body) {
			record.BodyJSON = json.RawMessage(result.Body)
		}
		r.Response = record
	} else {
		r.State = stateNoAnswer
	}
	r.Outcome = classify(result, answer)
	r.Guidance = guidanceFor(r)
	return r
}

func guidanceFor(r receipt) string {
	switch r.Outcome {
	case outcomeAccepted:
		// 2xx 는 접수일 뿐 a087 §5 의 「수용」이 아님 — 비동기 REJECTED 가 있음(3차 재리뷰 A-P1-2).
		return "the broker took the request (HTTP 200, orderId recorded) — this is NOT yet the a087 §5 'accepted' row; " +
			"read this orderId's status with `tossctl orders list` / `tossctl orders completed`: " +
			"PENDING / PARTIAL_FILLED / FILLED = accepted, REJECTED = refused (fill quality is NOT measured here)"
	case outcomeRefused:
		// refused 는 HTTP 4xx 라는 뜻뿐 — 인증·수량·한도 사유는 MARKET 반증이 아님(3차 재리뷰 B-P3-2).
		return "the broker refused the request (HTTP " + strconv.Itoa(r.Response.HTTPStatus) +
			"); no order is expected — confirm with `tossctl orders list` before any new attempt. " +
			"refused is NOT by itself a MARKET disproof: match the error code against the a087 tasks §5 table " +
			"(auth, quantity, rate-limit or symbol-state codes are 'undecidable')"
	default:
		return "the result is UNKNOWN — an order may exist. Do NOT resend with this tool; " +
			"check `tossctl orders list` / `tossctl orders completed` for this symbol first"
	}
}

func marshalReceipt(r receipt) ([]byte, error) {
	encoded, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(encoded, '\n'), nil
}

var errAlreadySent = errors.New("a receipt for this client order id already exists — this key was already used; do not resend")

// createPendingReceipt 는 키 파일을 배타적으로 만들고 "sending" 을 적음. 이미 있으면 errAlreadySent.
func createPendingReceipt(path string, r receipt) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("creating the receipt directory: %w", err)
	}
	encoded, err := marshalReceipt(r)
	if err != nil {
		return err
	}
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if errors.Is(err, os.ErrExist) {
		return errAlreadySent
	}
	if err != nil {
		return fmt.Errorf("creating the receipt: %w", err)
	}
	if _, err := file.Write(encoded); err != nil {
		_ = file.Close()
		return fmt.Errorf("writing the receipt: %w", err)
	}
	if err := file.Sync(); err != nil {
		_ = file.Close()
		return fmt.Errorf("syncing the receipt: %w", err)
	}
	return file.Close()
}

// finalizeReceipt 는 같은 경로를 임시 파일 + rename 으로 덮어씀 — 중간 상태의 반쪽 파일이 남지 않게 함.
func finalizeReceipt(path string, r receipt) error {
	encoded, err := marshalReceipt(r)
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), "."+filepath.Base(path)+".tmp-*")
	if err != nil {
		return fmt.Errorf("creating the receipt temp file: %w", err)
	}
	tmpName := tmp.Name()
	if _, err := tmp.Write(encoded); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpName)
		return fmt.Errorf("writing the receipt temp file: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpName)
		return fmt.Errorf("syncing the receipt temp file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpName)
		return err
	}
	if err := os.Chmod(tmpName, 0o644); err != nil {
		_ = os.Remove(tmpName)
		return err
	}
	return os.Rename(tmpName, path)
}
