package dalec

import (
	"io/fs"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateArtifactDirectories_GetConfig(t *testing.T) {
	tests := []struct {
		name     string
		dirs     *CreateArtifactDirectories
		expected map[string]ArtifactDirConfig
	}{
		{
			name:     "nil directories",
			dirs:     nil,
			expected: nil,
		},
		{
			name: "empty config",
			dirs: &CreateArtifactDirectories{},
			expected: nil,
		},
		{
			name: "with config directories",
			dirs: &CreateArtifactDirectories{
				Config: map[string]ArtifactDirConfig{
					"nginx": {Mode: 0755},
					"ssl":   {Mode: 0700},
				},
			},
			expected: map[string]ArtifactDirConfig{
				"nginx": {Mode: 0755},
				"ssl":   {Mode: 0700},
			},
		},
		{
			name: "with state but no config",
			dirs: &CreateArtifactDirectories{
				State: map[string]ArtifactDirConfig{
					"data": {Mode: 0755},
				},
			},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.dirs.GetConfig()
			assert.Equal(t, tt.expected, result)

			// Test that returned map is a clone (modification doesn't affect original)
			if result != nil && tt.dirs != nil {
				result["test"] = ArtifactDirConfig{Mode: 0777}
				assert.NotContains(t, tt.dirs.Config, "test")
			}
		})
	}
}

func TestCreateArtifactDirectories_GetState(t *testing.T) {
	tests := []struct {
		name     string
		dirs     *CreateArtifactDirectories
		expected map[string]ArtifactDirConfig
	}{
		{
			name:     "nil directories",
			dirs:     nil,
			expected: nil,
		},
		{
			name: "empty state",
			dirs: &CreateArtifactDirectories{},
			expected: nil,
		},
		{
			name: "with state directories",
			dirs: &CreateArtifactDirectories{
				State: map[string]ArtifactDirConfig{
					"data":  {Mode: 0755},
					"cache": {Mode: 0750},
				},
			},
			expected: map[string]ArtifactDirConfig{
				"data":  {Mode: 0755},
				"cache": {Mode: 0750},
			},
		},
		{
			name: "with config but no state",
			dirs: &CreateArtifactDirectories{
				Config: map[string]ArtifactDirConfig{
					"nginx": {Mode: 0755},
				},
			},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.dirs.GetState()
			assert.Equal(t, tt.expected, result)

			// Test that returned map is a clone (modification doesn't affect original)
			if result != nil && tt.dirs != nil {
				result["test"] = ArtifactDirConfig{Mode: 0777}
				assert.NotContains(t, tt.dirs.State, "test")
			}
		})
	}
}

func TestArtifacts_IsEmpty(t *testing.T) {
	tests := []struct {
		name      string
		artifacts *Artifacts
		expected  bool
	}{
		{
			name:      "completely empty artifacts",
			artifacts: &Artifacts{},
			expected:  true,
		},
		{
			name: "with binaries",
			artifacts: &Artifacts{
				Binaries: map[string]ArtifactConfig{
					"app": {},
				},
			},
			expected: false,
		},
		{
			name: "with manpages",
			artifacts: &Artifacts{
				Manpages: map[string]ArtifactConfig{
					"app.1": {},
				},
			},
			expected: false,
		},
		{
			name: "with config directories",
			artifacts: &Artifacts{
				Directories: &CreateArtifactDirectories{
					Config: map[string]ArtifactDirConfig{
						"nginx": {Mode: 0755},
					},
				},
			},
			expected: false,
		},
		{
			name: "with state directories",
			artifacts: &Artifacts{
				Directories: &CreateArtifactDirectories{
					State: map[string]ArtifactDirConfig{
						"data": {Mode: 0755},
					},
				},
			},
			expected: false,
		},
		{
			name: "with empty directories object",
			artifacts: &Artifacts{
				Directories: &CreateArtifactDirectories{},
			},
			expected: true,
		},
		{
			name: "with data dirs",
			artifacts: &Artifacts{
				DataDirs: map[string]ArtifactConfig{
					"share": {},
				},
			},
			expected: false,
		},
		{
			name: "with config files",
			artifacts: &Artifacts{
				ConfigFiles: map[string]ArtifactConfig{
					"app.conf": {},
				},
			},
			expected: false,
		},
		{
			name: "with systemd units",
			artifacts: &Artifacts{
				Systemd: &SystemdConfiguration{
					Units: map[string]SystemdUnitConfig{
						"app.service": {},
					},
				},
			},
			expected: false,
		},
		{
			name: "with systemd dropins",
			artifacts: &Artifacts{
				Systemd: &SystemdConfiguration{
					Dropins: map[string]SystemdDropinConfig{
						"override.conf": {},
					},
				},
			},
			expected: false,
		},
		{
			name: "with empty systemd config",
			artifacts: &Artifacts{
				Systemd: &SystemdConfiguration{},
			},
			expected: true,
		},
		{
			name: "with docs",
			artifacts: &Artifacts{
				Docs: map[string]ArtifactConfig{
					"README": {},
				},
			},
			expected: false,
		},
		{
			name: "with licenses",
			artifacts: &Artifacts{
				Licenses: map[string]ArtifactConfig{
					"LICENSE": {},
				},
			},
			expected: false,
		},
		{
			name: "with libs",
			artifacts: &Artifacts{
				Libs: map[string]ArtifactConfig{
					"libapp.so": {},
				},
			},
			expected: false,
		},
		{
			name: "with links",
			artifacts: &Artifacts{
				Links: []ArtifactSymlinkConfig{
					{Source: "/usr/bin/app", Dest: "/usr/local/bin/app"},
				},
			},
			expected: false,
		},
		{
			name: "with headers",
			artifacts: &Artifacts{
				Headers: map[string]ArtifactConfig{
					"app.h": {},
				},
			},
			expected: false,
		},
		{
			name: "multiple empty collections",
			artifacts: &Artifacts{
				Binaries:    map[string]ArtifactConfig{},
				Manpages:    map[string]ArtifactConfig{},
				DataDirs:    map[string]ArtifactConfig{},
				ConfigFiles: map[string]ArtifactConfig{},
				Docs:        map[string]ArtifactConfig{},
				Licenses:    map[string]ArtifactConfig{},
				Libs:        map[string]ArtifactConfig{},
				Links:       []ArtifactSymlinkConfig{},
				Headers:     map[string]ArtifactConfig{},
				Directories: &CreateArtifactDirectories{},
				Systemd:     &SystemdConfiguration{},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.artifacts.IsEmpty()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestArtifacts_HasDocs(t *testing.T) {
	tests := []struct {
		name      string
		artifacts Artifacts
		expected  bool
	}{
		{
			name:      "no docs or manpages",
			artifacts: Artifacts{},
			expected:  false,
		},
		{
			name: "with docs",
			artifacts: Artifacts{
				Docs: map[string]ArtifactConfig{
					"README": {},
				},
			},
			expected: true,
		},
		{
			name: "with manpages",
			artifacts: Artifacts{
				Manpages: map[string]ArtifactConfig{
					"app.1": {},
				},
			},
			expected: true,
		},
		{
			name: "with both docs and manpages",
			artifacts: Artifacts{
				Docs: map[string]ArtifactConfig{
					"README": {},
				},
				Manpages: map[string]ArtifactConfig{
					"app.1": {},
				},
			},
			expected: true,
		},
		{
			name: "with empty docs map",
			artifacts: Artifacts{
				Docs: map[string]ArtifactConfig{},
			},
			expected: false,
		},
		{
			name: "with empty manpages map",
			artifacts: Artifacts{
				Manpages: map[string]ArtifactConfig{},
			},
			expected: false,
		},
		{
			name: "with other artifacts but no docs",
			artifacts: Artifacts{
				Binaries: map[string]ArtifactConfig{
					"app": {},
				},
				ConfigFiles: map[string]ArtifactConfig{
					"config": {},
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.artifacts.HasDocs()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestArtifactConfig_ResolveName(t *testing.T) {
	tests := []struct {
		name     string
		config   *ArtifactConfig
		path     string
		expected string
	}{
		{
			name:     "with custom name",
			config:   &ArtifactConfig{Name: "custom-name"},
			path:     "/usr/bin/original-name",
			expected: "custom-name",
		},
		{
			name:     "without custom name - simple path",
			config:   &ArtifactConfig{},
			path:     "/usr/bin/app",
			expected: "app",
		},
		{
			name:     "without custom name - complex path",
			config:   &ArtifactConfig{},
			path:     "/usr/local/bin/my-application",
			expected: "my-application",
		},
		{
			name:     "without custom name - root path",
			config:   &ArtifactConfig{},
			path:     "app",
			expected: "app",
		},
		{
			name:     "empty custom name falls back to path",
			config:   &ArtifactConfig{Name: ""},
			path:     "/usr/bin/fallback",
			expected: "fallback",
		},
		{
			name:     "with subpath but no name",
			config:   &ArtifactConfig{SubPath: "bin"},
			path:     "/build/output/myapp",
			expected: "myapp",
		},
		{
			name:     "with permissions but no name",
			config:   &ArtifactConfig{Permissions: fs.FileMode(0755)},
			path:     "/build/output/executable",
			expected: "executable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.ResolveName(tt.path)
			assert.Equal(t, tt.expected, result)
		})
	}
}