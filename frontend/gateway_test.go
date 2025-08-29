package frontend

import (
	"context"
	"errors"
	"maps"
	"testing"

	"github.com/moby/buildkit/client/llb"
	"github.com/moby/buildkit/client/llb/sourceresolver"
	"github.com/moby/buildkit/util/apicaps"
	gwclient "github.com/moby/buildkit/frontend/gateway/client"
	"github.com/moby/buildkit/solver/pb"
	"github.com/opencontainers/go-digest"
	"github.com/stretchr/testify/assert"
)

func TestGetBuildArg(t *testing.T) {
	tests := []struct {
		name     string
		opts     map[string]string
		key      string
		expected string
		found    bool
	}{
		{
			name:     "nil opts",
			opts:     nil,
			key:      "MY_ARG",
			expected: "",
			found:    false,
		},
		{
			name:     "empty opts",
			opts:     map[string]string{},
			key:      "MY_ARG",
			expected: "",
			found:    false,
		},
		{
			name: "build arg exists",
			opts: map[string]string{
				"build-arg:MY_ARG": "my_value",
			},
			key:      "MY_ARG",
			expected: "my_value",
			found:    true,
		},
		{
			name: "build arg with empty value",
			opts: map[string]string{
				"build-arg:EMPTY_ARG": "",
			},
			key:      "EMPTY_ARG",
			expected: "",
			found:    true,
		},
		{
			name: "multiple build args",
			opts: map[string]string{
				"build-arg:FIRST":  "first_value",
				"build-arg:SECOND": "second_value",
				"other-opt":        "other_value",
			},
			key:      "SECOND",
			expected: "second_value",
			found:    true,
		},
		{
			name: "key not found",
			opts: map[string]string{
				"build-arg:DIFFERENT": "value",
				"other-opt":           "other",
			},
			key:      "MISSING",
			expected: "",
			found:    false,
		},
		{
			name: "non-build-arg key with same suffix",
			opts: map[string]string{
				"some-prefix:MY_ARG": "wrong_value",
				"build-arg:MY_ARG":   "correct_value",
			},
			key:      "MY_ARG",
			expected: "correct_value",
			found:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := newTestClient(tt.opts)
			value, found := GetBuildArg(client, tt.key)
			assert.Equal(t, tt.expected, value)
			assert.Equal(t, tt.found, found)
		})
	}
}

func TestGetTargetKey(t *testing.T) {
	tests := []struct {
		name     string
		opts     map[string]string
		expected string
	}{
		{
			name:     "nil opts",
			opts:     nil,
			expected: "",
		},
		{
			name:     "empty opts",
			opts:     map[string]string{},
			expected: "",
		},
		{
			name: "target key exists",
			opts: map[string]string{
				keyTopLevelTarget: "my-target",
			},
			expected: "my-target",
		},
		{
			name: "target key empty",
			opts: map[string]string{
				keyTopLevelTarget: "",
			},
			expected: "",
		},
		{
			name: "target key with other options",
			opts: map[string]string{
				keyTopLevelTarget: "special-target",
				"other-option":    "other-value",
				"build-arg:VAR":   "var-value",
			},
			expected: "special-target",
		},
		{
			name: "target key missing",
			opts: map[string]string{
				"other-option": "value",
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := newTestClient(tt.opts)
			result := GetTargetKey(client)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Helper types and functions for testing

// testClient is a minimal implementation of gwclient.Client for testing gateway functions
type testClient struct {
	opts map[string]string
}

var _ gwclient.Client = (*testClient)(nil)

func newTestClient(opts map[string]string) *testClient {
	if opts == nil {
		opts = make(map[string]string)
	}
	return &testClient{
		opts: maps.Clone(opts),
	}
}

func (c *testClient) BuildOpts() gwclient.BuildOpts {
	return gwclient.BuildOpts{
		Opts:    maps.Clone(c.opts),
		LLBCaps: apicaps.CapSet{}, // Zero value for testing
	}
}

// Minimal implementations for gwclient.Client interface
func (c *testClient) Inputs(context.Context) (map[string]llb.State, error) {
	return make(map[string]llb.State), nil
}

func (c *testClient) NewContainer(context.Context, gwclient.NewContainerRequest) (gwclient.Container, error) {
	return nil, errors.New("not implemented")
}

func (c *testClient) ResolveImageConfig(ctx context.Context, ref string, opt sourceresolver.Opt) (string, digest.Digest, []byte, error) {
	return "", "", nil, errors.New("not implemented")
}

func (c *testClient) ResolveSourceMetadata(ctx context.Context, op *pb.SourceOp, opt sourceresolver.Opt) (*sourceresolver.MetaResponse, error) {
	return nil, errors.New("not implemented")
}

func (c *testClient) Solve(ctx context.Context, req gwclient.SolveRequest) (*gwclient.Result, error) {
	return nil, errors.New("not implemented")
}

func (c *testClient) Warn(ctx context.Context, dgst digest.Digest, msg string, opts gwclient.WarnOpts) error {
	return errors.New("not implemented")
}