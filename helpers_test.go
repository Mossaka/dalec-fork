package dalec

import (
	"errors"
	"testing"

	"github.com/moby/buildkit/client/llb"
	ocispecs "github.com/opencontainers/image-spec/specs-go/v1"
	"gotest.tools/v3/assert"
	"gotest.tools/v3/assert/cmp"
)

func TestMergeDependencies(t *testing.T) {
	tests := []struct {
		name     string
		base     *PackageDependencies
		target   *PackageDependencies
		expected *PackageDependencies
	}{
		{
			name:     "both nil",
			base:     nil,
			target:   nil,
			expected: nil,
		},
		{
			name: "base nil",
			base: nil,
			target: &PackageDependencies{
				Build: map[string]PackageConstraints{
					"pkg1": {},
				},
			},
			expected: &PackageDependencies{
				Build: map[string]PackageConstraints{
					"pkg1": {},
				},
			},
		},
		{
			name: "target nil",
			base: &PackageDependencies{
				Runtime: map[string]PackageConstraints{
					"pkg2": {},
				},
			},
			target: nil,
			expected: &PackageDependencies{
				Runtime: map[string]PackageConstraints{
					"pkg2": {},
				},
			},
		},
		{
			name: "merge dependencies",
			base: &PackageDependencies{
				Build: map[string]PackageConstraints{
					"pkg1": {},
				},
				Runtime: map[string]PackageConstraints{
					"pkg2": {},
				},
			},
			target: &PackageDependencies{
				Build: map[string]PackageConstraints{
					"pkg3": {},
				},
				Test: []string{"test1"},
			},
			expected: &PackageDependencies{
				Build: map[string]PackageConstraints{
					"pkg3": {},
				},
				Runtime: map[string]PackageConstraints{
					"pkg2": {},
				},
				Test: []string{"test1"},
			},
		},
		{
			name: "custom repo in target",
			base: &PackageDependencies{
				Build: map[string]PackageConstraints{
					"pkg1": {},
				},
				Runtime: map[string]PackageConstraints{
					"pkg2": {},
				},
			},
			target: &PackageDependencies{
				ExtraRepos: []PackageRepositoryConfig{
					{
						Config: map[string]Source{
							"custom.repo": {
								HTTP: &SourceHTTP{
									URL: "my.repo.com/custom.repo",
								},
							},
						},
					},
				},
			},
			expected: &PackageDependencies{
				Build: map[string]PackageConstraints{
					"pkg1": {},
				},
				Runtime: map[string]PackageConstraints{
					"pkg2": {},
				},
				ExtraRepos: []PackageRepositoryConfig{
					{
						Config: map[string]Source{
							"custom.repo": {
								HTTP: &SourceHTTP{
									URL: "my.repo.com/custom.repo",
								},
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MergeDependencies(tt.base, tt.target)
			assert.Check(t, cmp.DeepEqual(tt.expected, result))
		})
	}
}

func TestDisableDiffMerge(t *testing.T) {
	// Test setting to true
	DisableDiffMerge(true)
	assert.Check(t, disableDiffMerge.Load() == true)
	
	// Test setting to false
	DisableDiffMerge(false)
	assert.Check(t, disableDiffMerge.Load() == false)
}

func TestWithDirContentsOnly(t *testing.T) {
	opt := WithDirContentsOnly()
	copyInfo := &llb.CopyInfo{}
	opt.SetCopyOption(copyInfo)
	assert.Check(t, copyInfo.CopyDirContentsOnly == true)
}

func TestWithRunOptions(t *testing.T) {
	// Test with multiple options - basic functionality test
	opt1 := llb.Args([]string{"echo", "test"})
	opt2 := llb.Dir("/tmp")
	
	combined := WithRunOptions(opt1, opt2)
	execInfo := &llb.ExecInfo{}
	
	// Should not panic and should apply options
	combined.SetRunOption(execInfo)
}

func TestWithConstraints(t *testing.T) {
	// Test with constraint options
	platform := ocispecs.Platform{Architecture: "amd64", OS: "linux"}
	opt := WithConstraints(llb.Platform(platform))
	
	constraints := &llb.Constraints{}
	opt.SetConstraintsOption(constraints)
	
	assert.Check(t, constraints.Platform != nil)
	assert.Check(t, constraints.Platform.Architecture == "amd64")
	assert.Check(t, constraints.Platform.OS == "linux")
}

func TestSortedMapValues(t *testing.T) {
	testMap := map[string]int{
		"zebra": 3,
		"apple": 1, 
		"banana": 2,
	}
	
	result := SortedMapValues(testMap)
	expected := []int{1, 2, 3} // sorted by keys: apple, banana, zebra
	
	assert.Check(t, cmp.DeepEqual(result, expected))
}

func TestShArgsf(t *testing.T) {
	opt := ShArgsf("echo %s %d", "test", 123)
	execInfo := &llb.ExecInfo{}
	
	// Should not panic - basic functionality test
	opt.SetRunOption(execInfo)
}

func TestHasValidSigner(t *testing.T) {
	// Test nil package config
	assert.Check(t, hasValidSigner(nil) == false)
	
	// Test package config with nil signer
	pc1 := &PackageConfig{}
	assert.Check(t, hasValidSigner(pc1) == false)
	
	// Test package config with signer but no image
	pc2 := &PackageConfig{
		Signer: &PackageSigner{
			Frontend: &Frontend{}, // Empty frontend with no image
		},
	}
	assert.Check(t, hasValidSigner(pc2) == false)
	
	// Test package config with valid signer (Image promoted from embedded Frontend)
	pc3 := &PackageConfig{
		Signer: &PackageSigner{
			Frontend: &Frontend{
				Image: "signerimage:latest",
			},
		},
	}
	assert.Check(t, hasValidSigner(pc3) == true)
}

func TestGetSigner(t *testing.T) {
	tests := []struct {
		name              string
		spec              *Spec
		targetKey         string
		expectedSigner    *PackageSigner
		expectedOverrides bool
	}{
		{
			name: "no signer anywhere",
			spec: &Spec{},
			targetKey: "test-target",
			expectedSigner: nil,
			expectedOverrides: false,
		},
		{
			name: "root level signer only",
			spec: &Spec{
				PackageConfig: &PackageConfig{
					Signer: &PackageSigner{
						Frontend: &Frontend{
							Image: "root-signer:latest",
						},
					},
				},
			},
			targetKey: "test-target",
			expectedSigner: &PackageSigner{
				Frontend: &Frontend{
					Image: "root-signer:latest",
				},
			},
			expectedOverrides: false,
		},
		{
			name: "target level signer overrides root",
			spec: &Spec{
				PackageConfig: &PackageConfig{
					Signer: &PackageSigner{
						Frontend: &Frontend{
							Image: "root-signer:latest",
						},
					},
				},
				Targets: map[string]Target{
					"test-target": {
						PackageConfig: &PackageConfig{
							Signer: &PackageSigner{
								Frontend: &Frontend{
									Image: "target-signer:latest",
								},
							},
						},
					},
				},
			},
			targetKey: "test-target",
			expectedSigner: &PackageSigner{
				Frontend: &Frontend{
					Image: "target-signer:latest",
				},
			},
			expectedOverrides: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			signer, overrides := tt.spec.GetSigner(tt.targetKey)
			assert.Check(t, cmp.DeepEqual(signer, tt.expectedSigner))
			assert.Check(t, overrides == tt.expectedOverrides)
		})
	}
}

func TestSpecGetRuntimeDeps(t *testing.T) {
	spec := &Spec{
		Dependencies: &PackageDependencies{
			Runtime: map[string]PackageConstraints{
				"zlib": {},
				"openssl": {},
				"apache": {},
			},
		},
	}
	
	deps := spec.GetRuntimeDeps("test-target")
	expected := []string{"apache", "openssl", "zlib"} // sorted
	assert.Check(t, cmp.DeepEqual(deps, expected))
}

func TestSpecGetTestDeps(t *testing.T) {
	spec := &Spec{
		Dependencies: &PackageDependencies{
			Test: []string{"pytest", "coverage", "mock"},
		},
	}
	
	deps := spec.GetTestDeps("test-target")
	expected := []string{"coverage", "mock", "pytest"} // sorted
	assert.Check(t, cmp.DeepEqual(deps, expected))
}

func TestWithRepoData(t *testing.T) {
	// Test with empty repos
	opt1 := WithRepoData([]PackageRepositoryConfig{}, SourceOpts{})
	execInfo1 := &llb.ExecInfo{}
	opt1.SetRunOption(execInfo1)
	// Should not panic and should be a no-op
	
	// Test with repos that have no data
	repos := []PackageRepositoryConfig{
		{
			Config: map[string]Source{
				"test.repo": {
					HTTP: &SourceHTTP{URL: "http://example.com/test.repo"},
				},
			},
		},
	}
	opt2 := WithRepoData(repos, SourceOpts{})
	execInfo2 := &llb.ExecInfo{}
	opt2.SetRunOption(execInfo2)
	// Should not panic
}

func TestHasGolang(t *testing.T) {
	tests := []struct {
		name      string
		spec      *Spec
		targetKey string
		expected  bool
	}{
		{
			name: "has golang",
			spec: &Spec{
				Dependencies: &PackageDependencies{
					Build: map[string]PackageConstraints{
						"golang": {},
					},
				},
			},
			targetKey: "test-target",
			expected:  true,
		},
		{
			name: "has msft-golang",
			spec: &Spec{
				Dependencies: &PackageDependencies{
					Build: map[string]PackageConstraints{
						"msft-golang": {},
					},
				},
			},
			targetKey: "test-target",
			expected:  true,
		},
		{
			name: "has golang-*",
			spec: &Spec{
				Dependencies: &PackageDependencies{
					Build: map[string]PackageConstraints{
						"golang-1.19": {},
					},
				},
			},
			targetKey: "test-target",
			expected:  true,
		},
		{
			name: "no golang",
			spec: &Spec{
				Dependencies: &PackageDependencies{
					Build: map[string]PackageConstraints{
						"python": {},
						"gcc": {},
					},
				},
			},
			targetKey: "test-target",
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HasGolang(tt.spec, tt.targetKey)
			assert.Check(t, result == tt.expected)
		})
	}
}

func TestHasNpm(t *testing.T) {
	tests := []struct {
		name      string
		spec      *Spec
		targetKey string
		expected  bool
	}{
		{
			name: "has npm",
			spec: &Spec{
				Dependencies: &PackageDependencies{
					Build: map[string]PackageConstraints{
						"npm": {},
					},
				},
			},
			targetKey: "test-target",
			expected:  true,
		},
		{
			name: "no npm",
			spec: &Spec{
				Dependencies: &PackageDependencies{
					Build: map[string]PackageConstraints{
						"nodejs": {},
						"yarn": {},
					},
				},
			},
			targetKey: "test-target",
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HasNpm(tt.spec, tt.targetKey)
			assert.Check(t, result == tt.expected)
		})
	}
}

func TestErrorState(t *testing.T) {
	state := llb.Scratch()
	
	// Test with nil error - should not panic
	ErrorState(state, nil)
	
	// Test with error - should not panic
	testErr := errors.New("test error")
	ErrorState(state, testErr)
}

func TestErrorStateOption(t *testing.T) {
	state := llb.Scratch()
	
	// Test with nil error - should not panic
	opt1 := ErrorStateOption(nil)
	opt1(state)
	
	// Test with error - should not panic
	testErr := errors.New("test error")  
	opt2 := ErrorStateOption(testErr)
	opt2(state)
}

func TestNoopStateOption(t *testing.T) {
	state := llb.Scratch()
	// Should not panic
	NoopStateOption(state)
}

func TestSpecGetProvides(t *testing.T) {
	spec := &Spec{
		Provides: map[string]PackageConstraints{
			"global-provide": {},
		},
		Targets: map[string]Target{
			"test-target": {
				Provides: map[string]PackageConstraints{
					"target-provide": {},
				},
			},
			"other-target": {},
		},
	}
	
	// Test target with specific provides
	result1 := spec.GetProvides("test-target")
	expected1 := map[string]PackageConstraints{"target-provide": {}}
	assert.Check(t, cmp.DeepEqual(result1, expected1))
	
	// Test target without specific provides - should get global
	result2 := spec.GetProvides("other-target")
	expected2 := map[string]PackageConstraints{"global-provide": {}}
	assert.Check(t, cmp.DeepEqual(result2, expected2))
}

func TestSpecGetReplaces(t *testing.T) {
	spec := &Spec{
		Replaces: map[string]PackageConstraints{
			"global-replace": {},
		},
		Targets: map[string]Target{
			"test-target": {
				Replaces: map[string]PackageConstraints{
					"target-replace": {},
				},
			},
			"other-target": {},
		},
	}
	
	// Test target with specific replaces
	result1 := spec.GetReplaces("test-target")
	expected1 := map[string]PackageConstraints{"target-replace": {}}
	assert.Check(t, cmp.DeepEqual(result1, expected1))
	
	// Test target without specific replaces - should get global
	result2 := spec.GetReplaces("other-target")
	expected2 := map[string]PackageConstraints{"global-replace": {}}
	assert.Check(t, cmp.DeepEqual(result2, expected2))
}

func TestSpecGetConflicts(t *testing.T) {
	spec := &Spec{
		Conflicts: map[string]PackageConstraints{
			"global-conflict": {},
		},
		Targets: map[string]Target{
			"test-target": {
				Conflicts: map[string]PackageConstraints{
					"target-conflict": {},
				},
			},
			"other-target": {},
		},
	}
	
	// Test target with specific conflicts
	result1 := spec.GetConflicts("test-target")
	expected1 := map[string]PackageConstraints{"target-conflict": {}}
	assert.Check(t, cmp.DeepEqual(result1, expected1))
	
	// Test target without specific conflicts - should get global
	result2 := spec.GetConflicts("other-target")
	expected2 := map[string]PackageConstraints{"global-conflict": {}}
	assert.Check(t, cmp.DeepEqual(result2, expected2))
}
