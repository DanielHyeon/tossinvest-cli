//go:build !unix

package engine

// The alert operator endpoint is a Unix socket because design D7.1 chose
// filesystem permissions as its access control. On other targets the engine runs
// unchanged and the endpoint simply does not exist — the CLI then fails to find a
// descriptor, which is the same answer it gives when the engine is not running.
//
// ⛔ This is not a degraded acknowledge path. There is none: acknowledging has to
// happen inside the process that holds the entry gate, and a build with no
// endpoint has no way in. Offering a ledger-only acknowledge here would be exactly
// the "operator acknowledged but entry stays shut" state D7.1 exists to prevent.

type AlertControlServer struct{}

func StartAlertControlServer(string, *AlertOperations) (*AlertControlServer, error) {
	return &AlertControlServer{}, nil
}

func (*AlertControlServer) Close() error { return nil }
