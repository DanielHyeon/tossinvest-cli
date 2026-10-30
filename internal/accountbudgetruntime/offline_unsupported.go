//go:build !linux

package accountbudgetruntime

import (
	"context"
	"github.com/JungHoonGhae/tossinvest-cli/internal/accountbudget"
)

func ProvisionOffline(context.Context, string, string, string, string) error {
	return accountbudget.ErrUnavailable
}
func CheckOffline(context.Context, string) error { return accountbudget.ErrUnavailable }
