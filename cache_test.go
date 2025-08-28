package dalec

import (
	"errors"
	"testing"

	"github.com/moby/buildkit/client/llb"
	ocispecs "github.com/opencontainers/image-spec/specs-go/v1"
	"gotest.tools/v3/assert"
)

func TestCacheConfigOptionFunc(t *testing.T) {
	var called bool
	opt := CacheConfigOptionFunc(func(info *CacheInfo) {
		called = true
		info.DirInfo.Platform = &ocispecs.Platform{OS: "linux"}
	})
	
	var info CacheInfo
	opt.SetCacheConfigOption(&info)
	
	assert.Assert(t, called, "option function should be called")
	assert.Equal(t, info.DirInfo.Platform.OS, "linux")
}

func TestCacheDirOptionFunc(t *testing.T) {
	var called bool
	opt := CacheDirOptionFunc(func(info *CacheDirInfo) {
		called = true
		info.Platform = &ocispecs.Platform{OS: "windows"}
	})
	
	var info CacheDirInfo
	opt.SetCacheDirOption(&info)
	
	assert.Assert(t, called, "option function should be called")
	assert.Equal(t, info.Platform.OS, "windows")
}

func TestWithCacheDirConstraints(t *testing.T) {
	platform := &ocispecs.Platform{OS: "linux", Architecture: "amd64"}
	constraintOpt := llb.Platform(*platform)
	
	opt := WithCacheDirConstraints(constraintOpt)
	
	var info CacheInfo
	opt.SetCacheConfigOption(&info)
	
	assert.Equal(t, info.DirInfo.Platform.OS, "linux")
	assert.Equal(t, info.DirInfo.Platform.Architecture, "amd64")
}

func TestCacheConfig_ToRunOption_Panic(t *testing.T) {
	// Test that empty CacheConfig panics as expected
	c := &CacheConfig{}
	
	assert.Assert(t, panics(func() {
		c.ToRunOption(llb.Scratch(), "test-distro")
	}), "empty cache config should panic")
}

func TestCacheConfig_ToRunOption_Dir(t *testing.T) {
	c := &CacheConfig{
		Dir: &CacheDir{
			Dest:    "/tmp/test-cache",
			Sharing: "shared",
		},
	}
	
	// Should not panic and return a valid run option
	opt := c.ToRunOption(llb.Scratch(), "test-distro")
	assert.Assert(t, opt != nil, "should return non-nil run option")
}

func TestCacheConfig_ToRunOption_GoBuild(t *testing.T) {
	c := &CacheConfig{
		GoBuild: &GoBuildCache{
			Scope: "test-scope",
		},
	}
	
	// Should not panic and return a valid run option
	opt := c.ToRunOption(llb.Scratch(), "test-distro")
	assert.Assert(t, opt != nil, "should return non-nil run option")
}

func TestCacheConfig_ToRunOption_Bazel(t *testing.T) {
	c := &CacheConfig{
		Bazel: &BazelCache{
			Scope: "test-scope",
		},
	}
	
	// Should not panic and return a valid run option
	opt := c.ToRunOption(llb.Scratch(), "test-distro")
	assert.Assert(t, opt != nil, "should return non-nil run option")
}

func TestCacheDir_ToRunOption_SharingModes(t *testing.T) {
	tests := []struct {
		name    string
		sharing string
	}{
		{"unset", ""},
		{"shared", "shared"},
		{"locked", "locked"},
		{"private", "private"},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &CacheDir{
				Dest:    "/tmp/test",
				Sharing: tt.sharing,
			}
			
			opt := c.ToRunOption("test-distro")
			assert.Assert(t, opt != nil, "should return non-nil run option")
		})
	}
}

func TestCacheDir_ToRunOption_InvalidSharing_Panic(t *testing.T) {
	c := &CacheDir{
		Dest:    "/tmp/test",
		Sharing: "invalid",
	}
	
	assert.Assert(t, panics(func() {
		opt := c.ToRunOption("test-distro")
		// Execute the run option to trigger the panic
		var ei llb.ExecInfo
		opt.SetRunOption(&ei)
	}), "invalid sharing mode should panic")
}

func TestCacheDir_ToRunOption_KeyHandling(t *testing.T) {
	// Test with explicit key
	c1 := &CacheDir{
		Key:  "explicit-key",
		Dest: "/tmp/test",
	}
	
	opt1 := c1.ToRunOption("test-distro")
	assert.Assert(t, opt1 != nil, "should return non-nil run option with explicit key")
	
	// Test with key derived from dest
	c2 := &CacheDir{
		Dest: "/tmp/test",
	}
	
	opt2 := c2.ToRunOption("test-distro")
	assert.Assert(t, opt2 != nil, "should return non-nil run option with derived key")
}

func TestCacheDir_ToRunOption_NoAutoNamespace(t *testing.T) {
	c := &CacheDir{
		Key:             "test-key",
		Dest:            "/tmp/test",
		NoAutoNamespace: true,
	}
	
	opt := c.ToRunOption("test-distro")
	assert.Assert(t, opt != nil, "should return non-nil run option with no auto namespace")
}

func TestGoBuildCache_validate(t *testing.T) {
	c := &GoBuildCache{
		Scope:    "test-scope",
		Disabled: false,
	}
	
	err := c.validate()
	assert.NilError(t, err, "GoBuildCache validation should always pass")
}

func TestGoBuildCacheOptionFunc(t *testing.T) {
	var called bool
	opt := GoBuildCacheOptionFunc(func(info *GoBuildCacheInfo) {
		called = true
		info.Platform = &ocispecs.Platform{OS: "darwin"}
	})
	
	var info GoBuildCacheInfo
	opt.SetGoBuildCacheOption(&info)
	
	assert.Assert(t, called, "option function should be called")
	assert.Equal(t, info.Platform.OS, "darwin")
}

func TestWithGoCacheConstraints(t *testing.T) {
	platform := &ocispecs.Platform{OS: "linux", Architecture: "arm64"}
	constraintOpt := llb.Platform(*platform)
	
	opt := WithGoCacheConstraints(constraintOpt)
	
	var info CacheInfo
	opt.SetCacheConfigOption(&info)
	
	assert.Equal(t, info.GoBuild.Platform.OS, "linux")
	assert.Equal(t, info.GoBuild.Platform.Architecture, "arm64")
}

func TestGoBuildCache_ToRunOption_Disabled(t *testing.T) {
	c := &GoBuildCache{
		Disabled: true,
	}
	
	opt := c.ToRunOption("test-distro")
	assert.Assert(t, opt != nil, "should return non-nil run option even when disabled")
	
	// When disabled, the run option should be a no-op
	var ei llb.ExecInfo
	opt.SetRunOption(&ei)
	
	// The exec info should remain unchanged when disabled (this is a basic smoke test)
	// This just tests that the function doesn't crash when disabled
}

func TestGoBuildCache_ToRunOption_WithScope(t *testing.T) {
	c := &GoBuildCache{
		Scope:    "custom-scope",
		Disabled: false,
	}
	
	opt := c.ToRunOption("test-distro")
	assert.Assert(t, opt != nil, "should return non-nil run option")
}

func TestBazelCache_validate(t *testing.T) {
	c := &BazelCache{
		Scope: "test-scope",
	}
	
	err := c.validate()
	assert.NilError(t, err, "BazelCache validation should always pass")
}

func TestBazelCacheOptionFunc(t *testing.T) {
	var called bool
	opt := BazelCacheOptionFunc(func(info *BazelCacheInfo) {
		called = true
		info.Platform = &ocispecs.Platform{OS: "linux"}
	})
	
	var info BazelCacheInfo
	opt.SetBazelCacheOption(&info)
	
	assert.Assert(t, called, "option function should be called")
	assert.Equal(t, info.Platform.OS, "linux")
}

func TestWithBazelCacheConstraints(t *testing.T) {
	platform := &ocispecs.Platform{OS: "windows", Architecture: "amd64"}
	constraintOpt := llb.Platform(*platform)
	
	opt := WithBazelCacheConstraints(constraintOpt)
	
	var info CacheInfo
	opt.SetCacheConfigOption(&info)
	
	assert.Equal(t, info.Bazel.Platform.OS, "windows")
	assert.Equal(t, info.Bazel.Platform.Architecture, "amd64")
}

func TestBazelCache_ToRunOption_WithScope(t *testing.T) {
	c := &BazelCache{
		Scope: "custom-scope",
	}
	
	opt := c.ToRunOption(llb.Scratch(), "test-distro")
	assert.Assert(t, opt != nil, "should return non-nil run option")
}

func TestBazelCache_ToRunOption_EmptyScope(t *testing.T) {
	c := &BazelCache{}
	
	opt := c.ToRunOption(llb.Scratch(), "test-distro")
	assert.Assert(t, opt != nil, "should return non-nil run option with empty scope")
}

// Helper function to test if a function panics
func panics(f func()) (panicked bool) {
	defer func() {
		if r := recover(); r != nil {
			panicked = true
		}
	}()
	f()
	return false
}

func TestCacheConfig_validate_NilConfig(t *testing.T) {
	var c *CacheConfig
	err := c.validate()
	assert.NilError(t, err, "nil cache config should be valid")
}

func TestCacheConfig_validate_ExactlyOne(t *testing.T) {
	tests := []struct {
		name        string
		config      CacheConfig
		expectError bool
	}{
		{
			name: "only dir",
			config: CacheConfig{
				Dir: &CacheDir{Dest: "/tmp/test"},
			},
			expectError: false,
		},
		{
			name: "only gobuild",
			config: CacheConfig{
				GoBuild: &GoBuildCache{},
			},
			expectError: false,
		},
		{
			name: "only bazel",
			config: CacheConfig{
				Bazel: &BazelCache{},
			},
			expectError: false,
		},
		{
			name: "none set",
			config: CacheConfig{},
			expectError: true,
		},
		{
			name: "multiple set",
			config: CacheConfig{
				Dir:     &CacheDir{Dest: "/tmp/test"},
				GoBuild: &GoBuildCache{},
			},
			expectError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.validate()
			if tt.expectError {
				assert.Assert(t, err != nil, "expected validation error")
			} else {
				assert.NilError(t, err, "expected no validation error")
			}
		})
	}
}

func TestCacheConfig_validate_SubConfigValidation(t *testing.T) {
	// Test that validation errors from sub-configs are propagated
	config := CacheConfig{
		Dir: &CacheDir{
			Dest:    "", // Invalid: empty dest
			Sharing: "invalid", // Invalid: bad sharing mode
		},
	}
	
	err := config.validate()
	assert.Assert(t, err != nil, "should return validation error")
	
	// Should be a joined error containing multiple validation failures
	var joinedErr interface{ Unwrap() []error }
	if errors.As(err, &joinedErr) {
		errs := joinedErr.Unwrap()
		assert.Assert(t, len(errs) > 0, "should contain sub-errors")
	}
}