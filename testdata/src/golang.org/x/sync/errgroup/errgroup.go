// Stub package for testing
package errgroup

import "context"

type Group struct{}

func (g *Group) Go(f func() error)         {}
func (g *Group) TryGo(f func() error) bool { return true }
func (g *Group) Wait() error               { return nil }
func (g *Group) SetLimit(n int)            {}

func WithContext(ctx context.Context) (*Group, context.Context) {
	return &Group{}, ctx
}
