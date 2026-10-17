package redact

// sinks.go — 줄 단위 싱크 둘: slog 핸들러(구조화 로그)와 io.Writer(프로세스의 표준 출력·오류).

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"strconv"
)

// Unloggable 은 치환 중 패닉한 값을 대신하는 문자열임.
const Unloggable = "[unloggable]"

// Handler 는 h 앞에 계좌 규칙을 세운 slog 핸들러를 돌려줌.
//
// 핵심: 계좌 키 속성은 버리고(그룹 안까지), 나머지 값은 **Handle 시점에** String 으로 치환함. With 로 묶은 속성도
// 안쪽 핸들러에 미리 넘기지 않고 여기 들고 있다가 Handle 때 치환함 — 계좌가 등록되기 전에 묶인 값도 새지 않게.
func Handler(h slog.Handler) slog.Handler {
	if h == nil {
		return nil
	}
	return &handler{inner: h}
}

type handler struct {
	inner slog.Handler
	// bound 는 그룹을 열기 전에 With 로 묶은 속성임. 그룹이 열린 뒤에 묶는 속성은 그 시점에 치환해 안쪽에 넘김
	// (생산 코드에 WithGroup 사용 0 — 그 경로의 잔재는 design 잔재 목록).
	bound   []slog.Attr
	grouped bool
}

func (h *handler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.inner.Enabled(ctx, level)
}

func (h *handler) Handle(ctx context.Context, r slog.Record) error {
	out := slog.NewRecord(r.Time, r.Level, String(r.Message), r.PC)
	for _, a := range h.bound {
		if kept, ok := scrubAttr(a); ok {
			out.AddAttrs(kept)
		}
	}
	r.Attrs(func(a slog.Attr) bool {
		if kept, ok := scrubAttr(a); ok {
			out.AddAttrs(kept)
		}
		return true
	})
	return h.inner.Handle(ctx, out)
}

func (h *handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	if h.grouped {
		var kept []slog.Attr
		for _, a := range attrs {
			if s, ok := scrubAttr(a); ok {
				kept = append(kept, s)
			}
		}
		return &handler{inner: h.inner.WithAttrs(kept), grouped: true}
	}
	bound := make([]slog.Attr, 0, len(h.bound)+len(attrs))
	bound = append(append(bound, h.bound...), attrs...)
	return &handler{inner: h.inner, bound: bound}
}

func (h *handler) WithGroup(name string) slog.Handler {
	// 그룹을 열기 전에 묶은 속성은 지금 치환해 안쪽에 넘김 — 그룹 밖 속성이 그룹 안으로 들어가지 않게.
	var kept []slog.Attr
	for _, a := range h.bound {
		if s, ok := scrubAttr(a); ok {
			kept = append(kept, s)
		}
	}
	inner := h.inner
	if len(kept) > 0 {
		inner = inner.WithAttrs(kept)
	}
	return &handler{inner: inner.WithGroup(name), grouped: true}
}

// scrubAttr 는 속성 하나를 규칙에 통과시킴. 두 번째 반환이 false 면 버림(계좌 키).
//
// 값 종류별: 문자열은 치환, 그룹은 재귀, error·Stringer 는 문자열로 만들어 치환, 그 밖의 Any 는 JSON 으로 만들어 치환이
// 생길 때만 치환된 문자열로 바꿈(안 생기면 원래 값 — 출력 모양 보존). 숫자 등도 문자열 꼴이 치환되면 문자열로 바꿈.
func scrubAttr(a slog.Attr) (out slog.Attr, keep bool) {
	if IsAccountKey(a.Key) {
		return slog.Attr{}, false
	}
	// 값의 MarshalJSON · String · Error 가 패닉해도 줄과 루프는 살아야 함(손절 루프가 이 로거를 씀) — 그 속성만 대체.
	defer func() {
		if recover() != nil {
			out, keep = slog.String(a.Key, Unloggable), true
		}
	}()
	v := a.Value.Resolve()
	switch v.Kind() {
	case slog.KindString:
		return slog.String(a.Key, String(v.String())), true
	case slog.KindGroup:
		var kept []any
		for _, g := range v.Group() {
			if s, ok := scrubAttr(g); ok {
				kept = append(kept, s)
			}
		}
		return slog.Group(a.Key, kept...), true
	case slog.KindAny:
		switch x := v.Any().(type) {
		case error:
			return slog.String(a.Key, String(x.Error())), true
		case fmt.Stringer:
			return slog.String(a.Key, String(x.String())), true
		case []byte:
			return slog.String(a.Key, String(string(x))), true
		default:
			// JSON 과 %+v 둘 다 봄 — 내보내지 않은 필드는 JSON 에 안 나오지만 텍스트 핸들러는 %+v 로 찍음(구현 리뷰 I-6).
			text, err := json.Marshal(x)
			if err != nil {
				text = []byte(fmt.Sprint(x))
			}
			plain := fmt.Sprintf("%+v", x)
			if scrubbed := String(string(text)); scrubbed != string(text) || String(plain) != plain {
				return slog.String(a.Key, String(plain)), true
			}
			return slog.Attr{Key: a.Key, Value: v}, true
		}
	case slog.KindFloat64:
		// 실수는 지수 꼴로 찍히면 숫자열 규칙이 못 봄 — 정수 꼴로도 대조.
		f := v.Float64()
		if text := strconv.FormatFloat(f, 'f', -1, 64); String(text) != text {
			return slog.String(a.Key, String(text)), true
		}
		return slog.Attr{Key: a.Key, Value: v}, true
	default:
		text := v.String()
		if scrubbed := String(text); scrubbed != text {
			return slog.String(a.Key, scrubbed), true
		}
		return slog.Attr{Key: a.Key, Value: v}, true
	}
}

// Writer 는 쓰기마다 String 을 적용하는 io.Writer 를 돌려줌. 줄 단위 쓰기를 전제함(fmt 의 한 호출 · slog 의 한 줄 =
// 한 번의 Write) — 계좌가 두 쓰기에 걸쳐 쪼개지는 출력은 이 프로세스에 없음.
func Writer(w io.Writer) io.Writer {
	if w == nil {
		return nil
	}
	if already, ok := w.(*writer); ok {
		return already
	}
	return &writer{w: w}
}

type writer struct{ w io.Writer }

func (x *writer) Write(p []byte) (int, error) {
	if _, err := io.WriteString(x.w, String(string(p))); err != nil {
		return 0, err
	}
	// 호출자에게는 받은 길이를 돌려줌 — 치환으로 길이가 달라도 「다 썼음」의 뜻은 같음.
	return len(p), nil
}
