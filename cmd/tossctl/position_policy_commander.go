package main

import (
	"context"
	"errors"

	"github.com/JungHoonGhae/tossinvest-cli/internal/positionpolicy"
	"github.com/JungHoonGhae/tossinvest-cli/internal/positionpolicyrpc"
)

var errPositionPolicyRuntimeUnavailable = errors.New("position policy runtime read endpoint is unavailable")

type positionPolicyLifecycleClient interface {
	List(context.Context) ([]positionpolicy.State, error)
	Preview(context.Context, positionpolicy.Request) (positionpolicy.Preview, error)
	Apply(context.Context, positionpolicy.ApplyRequest) (positionpolicy.State, error)
}

// consolePositionPolicyCommander composes two authorities without widening
// either one: lifecycle commands remain on the engine's loopback client, while
// runtime adoption facts come from the separate read-only Unix endpoint.
type consolePositionPolicyCommander struct {
	lifecycle positionPolicyLifecycleClient
	runtime   positionpolicy.RuntimeReader
}

// positionPolicyRuntimeDescriptorReader intentionally resolves the descriptor
// for every read. The API sidecar may start before the engine, and either reader
// may outlive an engine restart that replaces the socket and bearer token;
// retaining a startup client would then leave the effective policy unknown.
type positionPolicyRuntimeDescriptorReader struct {
	descriptorPath string
}

func (r positionPolicyRuntimeDescriptorReader) Runtime(ctx context.Context) (positionpolicy.ManagementRuntime, error) {
	client, err := positionpolicyrpc.DialRuntime(ctx, r.descriptorPath)
	if err != nil {
		return positionpolicy.ManagementRuntime{}, err
	}
	return client.Runtime(ctx)
}

func (c *consolePositionPolicyCommander) List(ctx context.Context) ([]positionpolicy.State, error) {
	return c.lifecycle.List(ctx)
}

func (c *consolePositionPolicyCommander) Runtime(ctx context.Context) (positionpolicy.ManagementRuntime, error) {
	if c.runtime == nil {
		return positionpolicy.ManagementRuntime{}, errPositionPolicyRuntimeUnavailable
	}
	return c.runtime.Runtime(ctx)
}

func (c *consolePositionPolicyCommander) Preview(ctx context.Context,
	req positionpolicy.Request) (positionpolicy.Preview, error) {
	return c.lifecycle.Preview(ctx, req)
}

func (c *consolePositionPolicyCommander) Apply(ctx context.Context,
	req positionpolicy.ApplyRequest) (positionpolicy.State, error) {
	return c.lifecycle.Apply(ctx, req)
}
