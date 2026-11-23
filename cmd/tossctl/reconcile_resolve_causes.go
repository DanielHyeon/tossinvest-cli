package main

import (
	"errors"

	"github.com/JungHoonGhae/tossinvest-cli/internal/reconcile"
)

// reconcileResolveRefusedCause 는 RED 단계용 빈 껍데기임(a195 3.1) — 모든 입력을 거절함.
func reconcileResolveRefusedCause(blocks []reconcile.Block) error {
	return errors.New("engine reconcile-resolve: a195 cause table not implemented")
}
