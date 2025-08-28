package dalec

import (
	"strings"
	"testing"

	"gotest.tools/v3/assert"
)

// Tests for envGetterMap.Keys() function - currently 0% coverage
func TestEnvGetterMap_Keys(t *testing.T) {
	tests := []struct {
		name string
		m    envGetterMap
		want int
	}{
		{
			name: "empty map",
			m:    envGetterMap{},
			want: 0,
		},
		{
			name: "single key",
			m:    envGetterMap{"key1": "value1"},
			want: 1,
		},
		{
			name: "multiple keys",
			m: envGetterMap{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			},
			want: 3,
		},
		{
			name: "keys with empty values",
			m: envGetterMap{
				"key1": "",
				"key2": "value2",
			},
			want: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keys := tt.m.Keys()
			assert.Equal(t, len(keys), tt.want)
			
			// Verify all expected keys are present
			for expectedKey := range tt.m {
				found := false
				for _, k := range keys {
					if k == expectedKey {
						found = true
						break
					}
				}
				assert.Assert(t, found, "expected key %s not found in Keys() result", expectedKey)
			}
		})
	}
}

// Tests for AllowAnyArg function - currently 0% coverage
func TestAllowAnyArg(t *testing.T) {
	tests := []struct {
		name string
		arg  string
	}{
		{
			name: "empty string",
			arg:  "",
		},
		{
			name: "normal arg",
			arg:  "SOME_ARG",
		},
		{
			name: "special characters",
			arg:  "ARG_WITH_SPECIAL_!@#",
		},
		{
			name: "long arg",
			arg:  "VERY_LONG_ARGUMENT_NAME_WITH_LOTS_OF_CHARACTERS",
		},
		{
			name: "numbers",
			arg:  "ARG123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := AllowAnyArg(tt.arg)
			assert.Assert(t, result, "AllowAnyArg should always return true for %q", tt.arg)
		})
	}
}

// Tests for WithAllowAnyArg function - currently 0% coverage  
func TestWithAllowAnyArg(t *testing.T) {
	tests := []struct {
		name        string
		initialFunc func(string) bool
		testArg     string
	}{
		{
			name:        "replace DisallowAllUndeclared",
			initialFunc: DisallowAllUndeclared,
			testArg:     "ANY_ARG",
		},
		{
			name:        "replace nil function",
			initialFunc: nil,
			testArg:     "ANOTHER_ARG",
		},
		{
			name: "replace custom function",
			initialFunc: func(s string) bool {
				return s == "SPECIFIC_ARG"
			},
			testArg: "DIFFERENT_ARG",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &SubstituteConfig{
				AllowArg: tt.initialFunc,
			}

			// Apply WithAllowAnyArg
			WithAllowAnyArg(cfg)

			// Verify the function was replaced with AllowAnyArg
			assert.Assert(t, cfg.AllowArg != nil, "AllowArg function should not be nil")
			
			// Test that it now allows any arg
			result := cfg.AllowArg(tt.testArg)
			assert.Assert(t, result, "AllowArg should now allow any arg after WithAllowAnyArg")
			
			// Test with various args to ensure it's truly AllowAnyArg
			testArgs := []string{"", "TEST", "ANOTHER_TEST", "123", "SPECIAL_!@#"}
			for _, arg := range testArgs {
				assert.Assert(t, cfg.AllowArg(arg), "AllowArg should allow %q", arg)
			}
		})
	}
}

// Tests for rawYAML.UnmarshalYAML - currently 0% coverage
func TestRawYAML_UnmarshalYAML(t *testing.T) {
	tests := []struct {
		name     string
		yamlData string
	}{
		{
			name:     "simple key-value",
			yamlData: "key: value",
		},
		{
			name:     "empty yaml",
			yamlData: "",
		},
		{
			name:     "complex yaml",
			yamlData: "name: test\nversion: 1.0\nconfig:\n  debug: true",
		},
		{
			name:     "yaml with list",
			yamlData: "items:\n  - item1\n  - item2",
		},
		{
			name:     "binary data",
			yamlData: "some binary content \x00\x01\x02",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var raw rawYAML
			err := raw.UnmarshalYAML([]byte(tt.yamlData))
			
			// rawYAML.UnmarshalYAML never returns an error - it just stores the bytes
			assert.NilError(t, err)
			// Verify the raw data was stored exactly as provided
			assert.Equal(t, string(raw), tt.yamlData)
		})
	}
}

// Tests for extensionFields.UnmarshalYAML - currently 0% coverage
func TestExtensionFields_UnmarshalYAML(t *testing.T) {
	tests := []struct {
		name      string
		yamlData  string
		expectErr bool
		expectLen int
	}{
		{
			name:      "empty yaml",
			yamlData:  "",
			expectErr: false,
			expectLen: 0,
		},
		{
			name:      "simple extension fields",
			yamlData:  "x-custom: value1\nx-another: value2",
			expectErr: false,
			expectLen: 2,
		},
		{
			name:      "uppercase extension",
			yamlData:  "X-CUSTOM: uppervalue",
			expectErr: false,
			expectLen: 1,
		},
		{
			name:      "extension with complex value",
			yamlData:  "x-config:\n  debug: true\n  level: info",
			expectErr: false,
			expectLen: 1,
		},
		{
			name:      "regular field should error",
			yamlData:  "name: test",
			expectErr: true, // non-extension fields should cause error
			expectLen: 0,
		},
		{
			name:      "mixed extension and regular fields",
			yamlData:  "x-custom: extension\nname: test",
			expectErr: true, // non-extension field should cause error
			expectLen: 0,
		},
		{
			name:      "invalid yaml",
			yamlData:  "invalid: yaml: [unclosed",
			expectErr: true,
			expectLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ext extensionFields
			err := ext.UnmarshalYAML([]byte(tt.yamlData))
			
			if tt.expectErr {
				assert.Assert(t, err != nil, "expected an error but got nil")
			} else {
				assert.NilError(t, err)
				assert.Equal(t, len(ext), tt.expectLen)
				
				// Verify only x- prefixed fields are present
				for key := range ext {
					assert.Assert(t, strings.HasPrefix(key, "x-") || strings.HasPrefix(key, "X-"), 
						"extension field key %q should start with 'x-' or 'X-'", key)
				}
			}
		})
	}
}

// Tests for Spec.UnmarshalYAML - currently 0% coverage
func TestSpec_UnmarshalYAML(t *testing.T) {
	tests := []struct {
		name      string
		yamlData  string
		expectErr bool
		checkFunc func(*testing.T, *Spec)
	}{
		{
			name: "minimal valid spec",
			yamlData: `name: test-package
version: 1.0.0
revision: "1"`,
			expectErr: false,
			checkFunc: func(t *testing.T, s *Spec) {
				assert.Equal(t, s.Name, "test-package")
				assert.Equal(t, s.Version, "1.0.0")
				assert.Equal(t, s.Revision, "1")
			},
		},
		{
			name: "spec with sources",
			yamlData: `name: test-with-sources
version: 2.0.0
sources:
  main:
    http:
      url: https://example.com/archive.tar.gz`,
			expectErr: false,
			checkFunc: func(t *testing.T, s *Spec) {
				assert.Equal(t, s.Name, "test-with-sources")
				assert.Equal(t, s.Version, "2.0.0")
				assert.Assert(t, s.Sources != nil)
				assert.Assert(t, len(s.Sources) > 0)
				_, exists := s.Sources["main"]
				assert.Assert(t, exists, "should have 'main' source")
			},
		},
		{
			name:      "empty spec",
			yamlData:  "",
			expectErr: true, // Empty YAML results in nil document body
			checkFunc: nil,
		},
		{
			name:      "just whitespace",
			yamlData:  "   \n  \n",
			expectErr: true,
			checkFunc: nil,
		},
		{
			name: "minimal name only",
			yamlData: `name: minimal`,
			expectErr: false,
			checkFunc: func(t *testing.T, s *Spec) {
				assert.Equal(t, s.Name, "minimal")
				assert.Equal(t, s.Version, "") // version is empty but that's valid for unmarshaling
			},
		},
		{
			name:      "invalid yaml syntax",
			yamlData:  "name: test\ninvalid: yaml: [unclosed",
			expectErr: true,
			checkFunc: nil,
		},
		{
			name:      "multiple documents should error",
			yamlData:  "name: doc1\n---\nname: doc2",
			expectErr: true, // Should fail with "expected exactly one yaml document"
			checkFunc: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var spec Spec
			err := spec.UnmarshalYAML([]byte(tt.yamlData))
			
			if tt.expectErr {
				assert.Assert(t, err != nil, "expected an error but got nil")
			} else {
				assert.NilError(t, err)
				if tt.checkFunc != nil {
					tt.checkFunc(t, &spec)
				}
			}
		})
	}
}