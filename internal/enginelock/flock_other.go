//go:build !unix

package enginelock

// flock_other.go is the honest refusal.
//
// soakproc_other.go degrades a *convenience* (a detached survey keeps the
// parent's process group — worse, still useful). This is not that: the thing
// being degraded is the guarantee that one journal has one writer, and an
// exclusion that does not exclude is worse than no engine at all. So a
// non-Unix build refuses to start the runtime and says why, rather than starting
// it with the door open.
//
// Adding Windows support means adding a LockFileEx implementation here, not
// relaxing the refusal.

func acquireLock(string) (lockHandle, error) { return nil, ErrLockUnsupported }
