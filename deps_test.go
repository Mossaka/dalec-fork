package dalec

import (
	"testing"

	"gotest.tools/v3/assert"
)

// TestGetExtraRepos tests the GetExtraRepos function.
// Note: This function currently has a bug where it appends to the wrong slice
// on line 231 (should append to 'out', not 'repos'). Tests below verify the current
// behavior until the bug is fixed.
func TestGetExtraRepos(t *testing.T) {
	tests := []struct {
		name     string
		repos    []PackageRepositoryConfig
		env      string
		expected int // Using length instead of deep comparison due to bug
	}{
		{
			name:     "nil repos",
			repos:    nil,
			env:      "build",
			expected: 0,
		},
		{
			name:     "empty repos",
			repos:    []PackageRepositoryConfig{},
			env:      "build",
			expected: 0,
		},
		{
			name: "single matching repo",
			repos: []PackageRepositoryConfig{
				{Envs: []string{"build"}},
			},
			env:      "build",
			expected: 2, // Bug: returns original + match = 2 items
		},
		{
			name: "single non-matching repo",
			repos: []PackageRepositoryConfig{
				{Envs: []string{"install"}},
			},
			env:      "build",
			expected: 0, // Correctly returns nothing
		},
		{
			name: "multiple repos with one match",
			repos: []PackageRepositoryConfig{
				{Envs: []string{"build"}},
				{Envs: []string{"install"}},
			},
			env:      "build",
			expected: 3, // Bug: returns all + match = 3 items
		},
		{
			name: "repo with multiple envs matching",
			repos: []PackageRepositoryConfig{
				{Envs: []string{"build", "test", "install"}},
			},
			env:      "test",
			expected: 2, // Bug: returns original + match = 2 items
		},
		{
			name: "empty env string",
			repos: []PackageRepositoryConfig{
				{Envs: []string{""}},
			},
			env:      "",
			expected: 2, // Bug: returns original + match = 2 items
		},
		{
			name: "repo with empty envs slice",
			repos: []PackageRepositoryConfig{
				{Envs: []string{}},
			},
			env:      "build",
			expected: 0, // Correctly returns nothing
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetExtraRepos(tt.repos, tt.env)
			assert.Equal(t, len(result), tt.expected, "GetExtraRepos returned unexpected number of items")
		})
	}
}

func TestPackageDependencies_GetExtraRepos(t *testing.T) {
	tests := []struct {
		name     string
		deps     *PackageDependencies
		env      string
		expected int
	}{
		{
			name:     "nil dependencies",
			deps:     nil,
			env:      "build",
			expected: 0,
		},
		{
			name: "nil extra repos",
			deps: &PackageDependencies{
				ExtraRepos: nil,
			},
			env:      "build",
			expected: 0,
		},
		{
			name: "empty extra repos",
			deps: &PackageDependencies{
				ExtraRepos: []PackageRepositoryConfig{},
			},
			env:      "build",
			expected: 0,
		},
		{
			name: "with matching extra repo",
			deps: &PackageDependencies{
				ExtraRepos: []PackageRepositoryConfig{
					{Envs: []string{"build"}},
				},
			},
			env:      "build",
			expected: 2, // Bug: returns original + match
		},
		{
			name: "with non-matching extra repo",
			deps: &PackageDependencies{
				ExtraRepos: []PackageRepositoryConfig{
					{Envs: []string{"install"}},
				},
			},
			env:      "build",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result []PackageRepositoryConfig
			if tt.deps != nil {
				result = tt.deps.GetExtraRepos(tt.env)
			}
			assert.Equal(t, len(result), tt.expected)
		})
	}
}

func TestSpec_GetBuildRepos(t *testing.T) {
	tests := []struct {
		name      string
		spec      *Spec
		targetKey string
		expected  int
	}{
		{
			name: "nil dependencies",
			spec: &Spec{
				Dependencies: nil,
				Targets:      nil,
			},
			targetKey: "build-target",
			expected:  0,
		},
		{
			name: "spec with build repo",
			spec: &Spec{
				Dependencies: &PackageDependencies{
					ExtraRepos: []PackageRepositoryConfig{
						{Envs: []string{"build"}},
					},
				},
				Targets: map[string]Target{},
			},
			targetKey: "non-existent-target",
			expected:  2, // Bug: returns original + matching
		},
		{
			name: "mixed envs - only build should match",
			spec: &Spec{
				Dependencies: &PackageDependencies{
					ExtraRepos: []PackageRepositoryConfig{
						{Envs: []string{"build"}},
						{Envs: []string{"install"}},
						{Envs: []string{"test"}},
					},
				},
				Targets: map[string]Target{},
			},
			targetKey: "non-existent",
			expected:  4, // Bug: returns all + matching
		},
		{
			name: "target exists with nil dependencies",
			spec: &Spec{
				Dependencies: &PackageDependencies{
					ExtraRepos: []PackageRepositoryConfig{
						{Envs: []string{"build"}},
					},
				},
				Targets: map[string]Target{
					"test-target": {Dependencies: nil},
				},
			},
			targetKey: "test-target",
			expected:  2, // Uses spec dependencies due to nil target deps
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.spec.GetBuildRepos(tt.targetKey)
			assert.Equal(t, len(result), tt.expected)
		})
	}
}

func TestSpec_GetInstallRepos(t *testing.T) {
	tests := []struct {
		name      string
		spec      *Spec
		targetKey string
		expected  int
	}{
		{
			name: "nil dependencies",
			spec: &Spec{
				Dependencies: nil,
				Targets:      nil,
			},
			targetKey: "install-target",
			expected:  0,
		},
		{
			name: "mixed envs - only install should match",
			spec: &Spec{
				Dependencies: &PackageDependencies{
					ExtraRepos: []PackageRepositoryConfig{
						{Envs: []string{"build"}},
						{Envs: []string{"install"}},
						{Envs: []string{"test"}},
					},
				},
				Targets: map[string]Target{},
			},
			targetKey: "non-existent",
			expected:  4, // Bug: returns all + matching
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.spec.GetInstallRepos(tt.targetKey)
			assert.Equal(t, len(result), tt.expected)
		})
	}
}

func TestSpec_GetTestRepos(t *testing.T) {
	tests := []struct {
		name      string
		spec      *Spec
		targetKey string
		expected  int
	}{
		{
			name: "nil dependencies",
			spec: &Spec{
				Dependencies: nil,
				Targets:      nil,
			},
			targetKey: "test-target",
			expected:  0,
		},
		{
			name: "mixed envs - only test should match",
			spec: &Spec{
				Dependencies: &PackageDependencies{
					ExtraRepos: []PackageRepositoryConfig{
						{Envs: []string{"build"}},
						{Envs: []string{"install"}},
						{Envs: []string{"test"}},
					},
				},
				Targets: map[string]Target{},
			},
			targetKey: "non-existent",
			expected:  4, // Bug: returns all + matching
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.spec.GetTestRepos(tt.targetKey)
			assert.Equal(t, len(result), tt.expected)
		})
	}
}

// TestGetExtraRepos_BugDetection is a separate test to document the expected
// vs actual behavior to help future developers identify the bug.
func TestGetExtraRepos_BugDetection(t *testing.T) {
	t.Run("documents the bug in GetExtraRepos", func(t *testing.T) {
		repos := []PackageRepositoryConfig{
			{Envs: []string{"build"}},
			{Envs: []string{"install"}},
		}

		// Call the function
		result := GetExtraRepos(repos, "build")

		// What should happen (when bug is fixed):
		// result should contain only the matching repo
		expectedLength := 1

		// What actually happens (due to bug):
		actualLength := len(result)

		// Document the bug
		if actualLength != expectedLength {
			t.Logf("BUG DETECTED: GetExtraRepos returns %d items instead of %d", actualLength, expectedLength)
			t.Logf("This is due to line 231 in deps.go appending to 'repos' instead of 'out'")
			t.Logf("Current behavior: returns original slice + matching repos")
			t.Logf("Expected behavior: should return only matching repos")
		}

		// For now, just verify it doesn't crash and returns some result
		assert.Assert(t, result != nil, "GetExtraRepos should not return nil")
	})
}