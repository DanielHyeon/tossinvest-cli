//go:build !linux

package shadowsession

import "context"

type Session struct{}

func Begin(context.Context, Root, string) (*Session, error) { return nil, ErrInvalid }
func (*Session) Pause() error                               { return ErrClosed }
func (*Session) Abort() error                               { return ErrClosed }
func (*Session) Complete() error                            { return ErrClosed }
func (*Session) Close() error                               { return ErrClosed }
