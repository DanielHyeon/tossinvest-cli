package httpapi

const (
	AuthModeBrowser    = "browser"
	AuthModeMTLS       = "mtls"
	AuthModeCapability = "capability"
)

type MutationIdentity struct {
	Actor  string
	Client string
	Mode   string
}

func (i MutationIdentity) valid() bool {
	return identityComponent.MatchString(i.Actor) && identityComponent.MatchString(i.Client) &&
		(i.Mode == AuthModeBrowser || i.Mode == AuthModeMTLS || i.Mode == AuthModeCapability)
}
