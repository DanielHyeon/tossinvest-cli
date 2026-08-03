//go:build !unix

package strategyprojectionrpc

import (
	"context"
	"errors"
	"os"

	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyprojection"
)

type Server struct{}

func Start(string, strategyprojection.Reader) (*Server, error) {
	return nil, errors.New("strategy projection runtime: Unix transport unavailable")
}

func Dial(context.Context, string) (*Client, error) {
	return nil, errors.New("strategy projection runtime: Unix transport unavailable")
}

func (s *Server) Close() error { return nil }

func openDescriptorNoFollow(string) (*os.File, error) {
	return nil, errors.New("strategy projection runtime: Unix transport unavailable")
}
