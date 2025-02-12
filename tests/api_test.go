package main

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/tests/environment"
)

func TestAPI(t *testing.T) {
	cases := []struct {
		name string
		run  func(t *testing.T)
	}{}

	env := environment.New(t)

	for _, tt := range cases {
		tc := tt
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()

			compose := env.Clone(t).Up(ctx)
			t.Cleanup(compose.Down)

		})
	}
}
