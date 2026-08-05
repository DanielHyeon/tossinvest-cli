package execgw

// export_status_test.go is a new file on purpose (change a082): a new function
// appended to an existing one pulls the function above it into the Function Logic
// Map requirement set, and export_test.go's neighbour is an unrelated init.
// StatusOfForTest exposes the sentinel→status inverse mapping.
//
// statusOf decides whether a mutation failure is a definitive refusal or an
// ambiguous one, and that decides whether a symbol gets blocked. It reaches the
// authentication sentinels only through errors.Is, and change a082 made those
// sentinels arrive wrapped — so the wrapped shape needs a test of its own rather
// than an assumption that it behaves like the bare one.
func StatusOfForTest(err error) (int, bool) { return statusOf(err) }
