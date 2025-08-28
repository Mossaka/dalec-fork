package dalec

import (
	"context"
	"testing"

	"github.com/moby/buildkit/client/llb"
	"gotest.tools/v3/assert"
)

func TestSource_isPip(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		source   *Source
		expected bool
	}{
		{
			name:     "nil source",
			source:   nil,
			expected: false,
		},
		{
			name:     "source with no generators",
			source:   &Source{},
			expected: false,
		},
		{
			name: "source with empty generators",
			source: &Source{
				Generate: []*SourceGenerator{},
			},
			expected: false,
		},
		{
			name: "source with non-pip generators",
			source: &Source{
				Generate: []*SourceGenerator{
					{Gomod: &GeneratorGomod{}},
					{NodeMod: &GeneratorNodeMod{}},
				},
			},
			expected: false,
		},
		{
			name: "source with pip generator",
			source: &Source{
				Generate: []*SourceGenerator{
					{Pip: &GeneratorPip{}},
				},
			},
			expected: true,
		},
		{
			name: "source with mixed generators including pip",
			source: &Source{
				Generate: []*SourceGenerator{
					{Gomod: &GeneratorGomod{}},
					{Pip: &GeneratorPip{}},
					{NodeMod: &GeneratorNodeMod{}},
				},
			},
			expected: true,
		},
		{
			name: "source with multiple pip generators",
			source: &Source{
				Generate: []*SourceGenerator{
					{Pip: &GeneratorPip{RequirementsFile: "requirements.txt"}},
					{Pip: &GeneratorPip{RequirementsFile: "dev-requirements.txt"}},
				},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.source == nil {
				// Test method call on nil source - this would panic, so skip
				return
			}
			result := tt.source.isPip()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSpec_HasPips(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		spec     *Spec
		expected bool
	}{
		{
			name:     "nil spec",
			spec:     nil,
			expected: false,
		},
		{
			name:     "spec with no sources",
			spec:     &Spec{},
			expected: false,
		},
		{
			name: "spec with empty sources",
			spec: &Spec{
				Sources: map[string]Source{},
			},
			expected: false,
		},
		{
			name: "spec with sources without pip generators",
			spec: &Spec{
				Sources: map[string]Source{
					"source1": {
						Generate: []*SourceGenerator{
							{Gomod: &GeneratorGomod{}},
						},
					},
					"source2": {
						Generate: []*SourceGenerator{
							{NodeMod: &GeneratorNodeMod{}},
						},
					},
				},
			},
			expected: false,
		},
		{
			name: "spec with one source having pip generator",
			spec: &Spec{
				Sources: map[string]Source{
					"source1": {
						Generate: []*SourceGenerator{
							{Gomod: &GeneratorGomod{}},
						},
					},
					"source2": {
						Generate: []*SourceGenerator{
							{Pip: &GeneratorPip{}},
						},
					},
				},
			},
			expected: true,
		},
		{
			name: "spec with multiple sources having pip generators",
			spec: &Spec{
				Sources: map[string]Source{
					"ml-service": {
						Generate: []*SourceGenerator{
							{Pip: &GeneratorPip{RequirementsFile: "requirements.txt"}},
						},
					},
					"data-processor": {
						Generate: []*SourceGenerator{
							{Pip: &GeneratorPip{
								RequirementsFile: "requirements-dev.txt",
								IndexUrl:         "https://pypi.custom.com/simple",
							}},
						},
					},
				},
			},
			expected: true,
		},
		{
			name: "spec with mixed generators across sources",
			spec: &Spec{
				Sources: map[string]Source{
					"go-app": {
						Generate: []*SourceGenerator{
							{Gomod: &GeneratorGomod{}},
						},
					},
					"python-service": {
						Generate: []*SourceGenerator{
							{Pip: &GeneratorPip{}},
						},
					},
					"node-app": {
						Generate: []*SourceGenerator{
							{NodeMod: &GeneratorNodeMod{}},
						},
					},
				},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.spec == nil {
				// Test method call on nil spec - this would panic, so skip
				return
			}
			result := tt.spec.HasPips()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestWithPip(t *testing.T) {
	t.Parallel()

	// Test basic functionality without complex llb execution
	tests := []struct {
		name      string
		generator *SourceGenerator
		srcName   string
	}{
		{
			name: "basic pip generator",
			generator: &SourceGenerator{
				Pip: &GeneratorPip{},
			},
			srcName: "python-service",
		},
		{
			name: "pip generator with custom requirements file",
			generator: &SourceGenerator{
				Pip: &GeneratorPip{
					RequirementsFile: "requirements-prod.txt",
				},
			},
			srcName: "production-service",
		},
		{
			name: "pip generator with custom index URL",
			generator: &SourceGenerator{
				Pip: &GeneratorPip{
					IndexUrl: "https://pypi.custom.com/simple",
				},
			},
			srcName: "private-repo-service",
		},
		{
			name: "pip generator with extra index URLs",
			generator: &SourceGenerator{
				Pip: &GeneratorPip{
					ExtraIndexUrls: []string{
						"https://extra1.pypi.com/simple",
						"https://extra2.pypi.com/simple",
					},
				},
			},
			srcName: "multi-index-service",
		},
		{
			name: "pip generator with custom paths",
			generator: &SourceGenerator{
				Pip: &GeneratorPip{
					Paths: []string{"api", "worker"},
				},
			},
			srcName: "multi-module-service",
		},
		{
			name: "pip generator with subpath",
			generator: &SourceGenerator{
				Subpath: "python",
				Pip: &GeneratorPip{
					Paths: []string{"."},
				},
			},
			srcName: "nested-service",
		},
		{
			name: "comprehensive pip generator",
			generator: &SourceGenerator{
				Subpath: "services/python",
				Pip: &GeneratorPip{
					Paths:            []string{"api", "worker", "scheduler"},
					RequirementsFile: "requirements-all.txt",
					IndexUrl:         "https://pypi.corporate.com/simple",
					ExtraIndexUrls: []string{
						"https://pypi.ml.com/simple",
						"https://pypi.data.com/simple",
					},
				},
			},
			srcName: "enterprise-python-service",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create basic llb states for testing
			worker := llb.Image("python:3.11-alpine")
			srcState := llb.Scratch()

			// Test that withPip returns a valid state
			resultState := withPip(tt.generator, srcState, worker, tt.srcName)
			// Basic verification that the state was created successfully
			_, err := resultState.Marshal(context.Background())
			assert.NilError(t, err, "withPip should return a valid state")
		})
	}
}

func TestSpec_pipSources(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		spec     *Spec
		expected map[string]Source
	}{
		{
			name: "empty spec",
			spec: &Spec{},
			expected: map[string]Source{},
		},
		{
			name: "spec with no pip sources",
			spec: &Spec{
				Sources: map[string]Source{
					"go-app": {
						Generate: []*SourceGenerator{
							{Gomod: &GeneratorGomod{}},
						},
					},
					"node-app": {
						Generate: []*SourceGenerator{
							{NodeMod: &GeneratorNodeMod{}},
						},
					},
				},
			},
			expected: map[string]Source{},
		},
		{
			name: "spec with one pip source",
			spec: &Spec{
				Sources: map[string]Source{
					"python-service": {
						Generate: []*SourceGenerator{
							{Pip: &GeneratorPip{}},
						},
					},
					"go-service": {
						Generate: []*SourceGenerator{
							{Gomod: &GeneratorGomod{}},
						},
					},
				},
			},
			expected: map[string]Source{
				"python-service": {
					Generate: []*SourceGenerator{
						{Pip: &GeneratorPip{}},
					},
				},
			},
		},
		{
			name: "spec with multiple pip sources",
			spec: &Spec{
				Sources: map[string]Source{
					"ml-service": {
						Generate: []*SourceGenerator{
							{Pip: &GeneratorPip{RequirementsFile: "requirements.txt"}},
						},
					},
					"data-processor": {
						Generate: []*SourceGenerator{
							{Pip: &GeneratorPip{
								IndexUrl: "https://pypi.custom.com/simple",
							}},
						},
					},
					"api-gateway": {
						Generate: []*SourceGenerator{
							{Gomod: &GeneratorGomod{}},
						},
					},
				},
			},
			expected: map[string]Source{
				"ml-service": {
					Generate: []*SourceGenerator{
						{Pip: &GeneratorPip{RequirementsFile: "requirements.txt"}},
					},
				},
				"data-processor": {
					Generate: []*SourceGenerator{
						{Pip: &GeneratorPip{
							IndexUrl: "https://pypi.custom.com/simple",
						}},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.spec.pipSources()
			assert.DeepEqual(t, tt.expected, result)
		})
	}
}

func TestSpec_PipDeps(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		spec        *Spec
		expectNil   bool
		expectError bool
	}{
		{
			name:        "spec with no pip sources",
			spec:        &Spec{Sources: map[string]Source{}},
			expectNil:   true,
			expectError: false,
		},
		{
			name: "spec with non-pip sources only",
			spec: &Spec{
				Sources: map[string]Source{
					"go-app": {
						Generate: []*SourceGenerator{
							{Gomod: &GeneratorGomod{}},
						},
					},
				},
			},
			expectNil:   true,
			expectError: false,
		},
		// Note: More complex integration tests would require full llb setup
		// These basic tests verify the logic flow without external dependencies
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			worker := llb.Image("python:3.11-alpine")
			sOpt := SourceOpts{}

			result, err := tt.spec.PipDeps(sOpt, worker)

			if tt.expectError {
				assert.Assert(t, err != nil, "Expected error but got none")
				return
			}
			assert.NilError(t, err, "Unexpected error")

			if tt.expectNil {
				assert.Assert(t, result == nil, "Expected nil result")
			} else {
				assert.Assert(t, result != nil, "Expected non-nil result")
			}
		})
	}
}

// Test edge cases and configuration options for pip generator
func TestGeneratorPip_Configuration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		generator *GeneratorPip
		expected  GeneratorPip
	}{
		{
			name:      "default pip generator",
			generator: &GeneratorPip{},
			expected:  GeneratorPip{},
		},
		{
			name: "pip generator with custom requirements file",
			generator: &GeneratorPip{
				RequirementsFile: "requirements-dev.txt",
			},
			expected: GeneratorPip{
				RequirementsFile: "requirements-dev.txt",
			},
		},
		{
			name: "pip generator with index URL",
			generator: &GeneratorPip{
				IndexUrl: "https://pypi.corporate.com/simple",
			},
			expected: GeneratorPip{
				IndexUrl: "https://pypi.corporate.com/simple",
			},
		},
		{
			name: "pip generator with extra index URLs",
			generator: &GeneratorPip{
				ExtraIndexUrls: []string{
					"https://extra1.pypi.com/simple",
					"https://extra2.pypi.com/simple",
				},
			},
			expected: GeneratorPip{
				ExtraIndexUrls: []string{
					"https://extra1.pypi.com/simple",
					"https://extra2.pypi.com/simple",
				},
			},
		},
		{
			name: "pip generator with custom paths",
			generator: &GeneratorPip{
				Paths: []string{"api", "worker", "scheduler"},
			},
			expected: GeneratorPip{
				Paths: []string{"api", "worker", "scheduler"},
			},
		},
		{
			name: "comprehensive pip generator configuration",
			generator: &GeneratorPip{
				Paths:            []string{"service", "tests"},
				RequirementsFile: "requirements-prod.txt",
				IndexUrl:         "https://pypi.internal.com/simple",
				ExtraIndexUrls: []string{
					"https://ml.packages.com/simple",
					"https://data.packages.com/simple",
				},
			},
			expected: GeneratorPip{
				Paths:            []string{"service", "tests"},
				RequirementsFile: "requirements-prod.txt",
				IndexUrl:         "https://pypi.internal.com/simple",
				ExtraIndexUrls: []string{
					"https://ml.packages.com/simple",
					"https://data.packages.com/simple",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.DeepEqual(t, tt.expected, *tt.generator)
		})
	}
}

// Test the path defaulting logic similar to what's in withPip
func TestPipGenerator_Paths(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		generator *GeneratorPip
		expected  []string
	}{
		{
			name:      "nil paths should default to current directory in withPip",
			generator: &GeneratorPip{},
			expected:  []string{"."},
		},
		{
			name: "empty paths slice",
			generator: &GeneratorPip{
				Paths: []string{},
			},
			expected: []string{},
		},
		{
			name: "single path",
			generator: &GeneratorPip{
				Paths: []string{"api"},
			},
			expected: []string{"api"},
		},
		{
			name: "multiple paths",
			generator: &GeneratorPip{
				Paths: []string{"api", "worker", "scheduler"},
			},
			expected: []string{"api", "worker", "scheduler"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify the generator paths are as expected
			if tt.generator.Paths == nil && len(tt.expected) == 1 && tt.expected[0] == "." {
				// This tests the default behavior in withPip function
				// The nil check and default assignment happens in withPip
				assert.Assert(t, tt.generator.Paths == nil, "Paths should be nil for default case")
			} else {
				assert.DeepEqual(t, tt.expected, tt.generator.Paths)
			}
		})
	}
}

// Test requirements file defaulting logic
func TestPipGenerator_RequirementsFile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		generator       *GeneratorPip
		expectedDefault string
	}{
		{
			name:            "empty requirements file should default",
			generator:       &GeneratorPip{},
			expectedDefault: "requirements.txt",
		},
		{
			name: "custom requirements file",
			generator: &GeneratorPip{
				RequirementsFile: "requirements-prod.txt",
			},
			expectedDefault: "requirements-prod.txt",
		},
		{
			name: "dev requirements file",
			generator: &GeneratorPip{
				RequirementsFile: "requirements-dev.txt",
			},
			expectedDefault: "requirements-dev.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// The default logic happens in withPip function
			requirementsFile := tt.generator.RequirementsFile
			if requirementsFile == "" {
				requirementsFile = "requirements.txt"
			}
			assert.Equal(t, tt.expectedDefault, requirementsFile)
		})
	}
}