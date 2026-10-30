package execgw_test

import (
	"context"
	"fmt"
	"net"
	"net/http/httptest"
)

// a166FixtureDial은 테스트 서버 외 주소의 연결을 거부함.
func a166FixtureDial(s *httptest.Server) func(context.Context, string, string) (net.Conn, error) {
	want := s.Listener.Addr().String()
	host, _, err := net.SplitHostPort(want)
	if err != nil || net.ParseIP(host) == nil || !net.ParseIP(host).IsLoopback() {
		panic("fixture listener must use loopback")
	}
	return func(ctx context.Context, network, address string) (net.Conn, error) {
		if address != want {
			return nil, fmt.Errorf("fixture refused nonlocal destination")
		}
		return (&net.Dialer{}).DialContext(ctx, network, want)
	}
}
