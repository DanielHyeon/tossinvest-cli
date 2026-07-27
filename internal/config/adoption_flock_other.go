//go:build !unix

package config

// adoption_flock_other.go is the honest refusal, mirroring
// internal/enginelock/flock_other.go: a save that cannot be serialized could
// install a stale read over another writer's rename, and a lock that does not
// lock is worse than a refusal. Windows support means a LockFileEx
// implementation here, not a relaxation.

import "errors"

type configLock struct{}

func (configLock) release() {}

func acquireConfigLock(string) (configLock, error) {
	return configLock{}, errors.New("config: saving from the console is not supported on this platform")
}
