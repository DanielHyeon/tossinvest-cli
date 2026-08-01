package httpapi

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync/atomic"
	"time"
)

type Options struct {
	Reader         Reader
	Stream         http.Handler
	MutationRoutes map[string]http.Handler
	Now            func() time.Time
}

type router struct {
	reader         Reader
	stream         http.Handler
	mutationRoutes map[string]http.Handler
	now            func() time.Time
}

var allowedMutationRoutes = map[string]struct{}{
	"/api/v1/optimization/previews":     {},
	"/api/v1/optimization/applications": {},
}

func NewRouter(options Options) (http.Handler, error) {
	if options.Reader == nil {
		return nil, errors.New("httpapi: read service is required")
	}
	if options.Now == nil {
		options.Now = time.Now
	}
	routes := make(map[string]http.Handler, len(options.MutationRoutes))
	for path, handler := range options.MutationRoutes {
		if _, allowed := allowedMutationRoutes[path]; !allowed || handler == nil {
			return nil, fmt.Errorf("httpapi: mutation route %q is not an exact approved resource", path)
		}
		routes[path] = handler
	}
	return &router{reader: options.Reader, stream: options.Stream, mutationRoutes: routes, now: options.Now}, nil
}

func (r *router) ServeHTTP(w http.ResponseWriter, request *http.Request) {
	setResponseHeaders(w)

	if mutation, ok := r.mutationRoutes[request.URL.Path]; ok {
		if request.Method != http.MethodPost {
			w.Header().Set("Allow", http.MethodPost)
			r.writeError(w, request, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "이 resource는 POST만 허용합니다.", nil)
			return
		}
		mutation.ServeHTTP(w, request)
		return
	}

	if request.URL.Path == "/api/v1/stream" {
		r.serveStream(w, request)
		return
	}

	resource, known := readResourceForPath(request.URL.Path)
	if !known {
		r.writeError(w, request, http.StatusNotFound, "RESOURCE_NOT_FOUND", "요청한 API resource가 없습니다.", nil)
		return
	}
	if request.Method != http.MethodGet && request.Method != http.MethodHead {
		w.Header().Set("Allow", http.MethodGet+", "+http.MethodHead)
		r.writeError(w, request, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "읽기 resource는 GET 또는 HEAD만 허용합니다.", nil)
		return
	}
	if request.Body != nil && request.Body != http.NoBody {
		r.writeError(w, request, http.StatusBadRequest, "BODY_NOT_SUPPORTED", "읽기 resource는 request body를 받지 않습니다.", []ErrorDetail{{
			Field: "body", Code: "UNEXPECTED_VALUE", Message: "request body를 제거하세요.",
		}})
		return
	}
	if request.URL.RawQuery != "" {
		r.writeError(w, request, http.StatusBadRequest, "QUERY_NOT_SUPPORTED", "이 resource는 서버가 정한 고정 조회만 제공합니다.", []ErrorDetail{{
			Field: "query", Code: "UNEXPECTED_VALUE", Message: "query parameter를 제거하세요.",
		}})
		return
	}

	data, err := r.read(request, resource)
	if err != nil {
		r.writeError(w, request, http.StatusServiceUnavailable, "RESOURCE_UNAVAILABLE", "현재 resource를 안전하게 읽을 수 없습니다.", nil)
		return
	}
	r.writeEnvelope(w, request, resource, data)
}

func readResourceForPath(path string) (string, bool) {
	const prefix = "/api/v1/"
	if !strings.HasPrefix(path, prefix) || strings.Contains(strings.TrimPrefix(path, prefix), "/") {
		return "", false
	}
	resource := strings.TrimPrefix(path, prefix)
	for _, known := range readResourceNames {
		if resource == known {
			return resource, true
		}
	}
	return "", false
}

func (r *router) read(request *http.Request, resource string) (any, error) {
	ctx := request.Context()
	switch resource {
	case "engine":
		return r.reader.Engine(ctx)
	case "positions":
		value, err := r.reader.Positions(ctx)
		if value.Items == nil {
			value.Items = []Position{}
		}
		return value, err
	case "orders":
		value, err := r.reader.Orders(ctx)
		if value.Items == nil {
			value.Items = []Order{}
		}
		return value, err
	case "candidates":
		value, err := r.reader.Candidates(ctx)
		if value.Items == nil {
			value.Items = []Candidate{}
		}
		return value, err
	case "performance":
		value, err := r.reader.Performance(ctx)
		return PerformanceFrom(value), err
	case "settings":
		value, err := r.reader.Settings(ctx)
		if value.Items == nil {
			value.Items = []Setting{}
		}
		return value, err
	case "optimization":
		value, err := r.reader.Optimization(ctx)
		return OptimizationFrom(value), err
	default:
		return nil, errors.New("httpapi: unknown read resource")
	}
}

func (r *router) serveStream(w http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodGet && request.Method != http.MethodHead {
		w.Header().Set("Allow", http.MethodGet+", "+http.MethodHead)
		r.writeError(w, request, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "stream은 GET 또는 HEAD만 허용합니다.", nil)
		return
	}
	if request.Body != nil && request.Body != http.NoBody {
		r.writeError(w, request, http.StatusBadRequest, "BODY_NOT_SUPPORTED", "stream은 request body를 받지 않습니다.", nil)
		return
	}
	if request.URL.RawQuery != "" {
		r.writeError(w, request, http.StatusBadRequest, "QUERY_NOT_SUPPORTED", "stream filter는 서버가 정하며 query를 받지 않습니다.", nil)
		return
	}
	if r.stream == nil {
		r.writeError(w, request, http.StatusServiceUnavailable, "STREAM_UNAVAILABLE", "현재 stream을 안전하게 열 수 없습니다.", nil)
		return
	}
	r.stream.ServeHTTP(w, request)
}

func setResponseHeaders(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Referrer-Policy", "no-referrer")
}

func (r *router) writeEnvelope(w http.ResponseWriter, request *http.Request, resource string, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	if request.Method == http.MethodHead {
		return
	}
	_ = json.NewEncoder(w).Encode(Envelope{SchemaVersion: SchemaVersion, Resource: resource,
		GeneratedAt: r.now().UTC(), Data: data})
}

func (r *router) writeError(w http.ResponseWriter, request *http.Request, status int, code, message string, details []ErrorDetail) {
	if details == nil {
		details = []ErrorDetail{}
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if request.Method == http.MethodHead {
		return
	}
	_, _ = w.Write(errorResponseBody(code, message, details, r.now()))
}

var requestIDFallback atomic.Uint64

func newRequestID() string {
	var value [12]byte
	if _, err := rand.Read(value[:]); err == nil {
		return "req_" + hex.EncodeToString(value[:])
	}
	return fmt.Sprintf("req_fallback_%d", requestIDFallback.Add(1))
}
