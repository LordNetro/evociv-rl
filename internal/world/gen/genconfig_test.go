package gen

import (
	"testing"
	"testing/fstest"
)

func TestGenConfigDefaults(t *testing.T) {
	fsys := fstest.MapFS{
		"gen-config.yaml": &fstest.MapFile{
			Data: []byte(`kind: gen-config
data:
  width: 64
  height: 64
`),
		},
	}
	cfg, err := LoadGenConfig("gen-config.yaml", fsys)
	if err != nil {
		t.Fatalf("LoadGenConfig error: %v", err)
	}
	if cfg.Width != 64 {
		t.Errorf("width = %d, want 64", cfg.Width)
	}
	if cfg.Height != 64 {
		t.Errorf("height = %d, want 64", cfg.Height)
	}
	if cfg.Octaves != 6 {
		t.Errorf("octaves default = %d, want 6", cfg.Octaves)
	}
	if cfg.Lacunarity != 2.0 {
		t.Errorf("lacunarity default = %v, want 2.0", cfg.Lacunarity)
	}
	if cfg.Gain != 0.5 {
		t.Errorf("gain default = %v, want 0.5", cfg.Gain)
	}
	if cfg.Scale != 100.0 {
		t.Errorf("scale default = %v, want 100.0", cfg.Scale)
	}
	if cfg.Seed != 0 {
		t.Errorf("seed default = %d, want 0", cfg.Seed)
	}
}

func TestGenConfigValidate(t *testing.T) {
	cases := []struct {
		name    string
		cfg     GenConfig
		wantErr bool
	}{
		{
			name:    "zero width",
			cfg:     GenConfig{Width: 0, Height: 64},
			wantErr: true,
		},
		{
			name:    "zero height",
			cfg:     GenConfig{Width: 64, Height: 0},
			wantErr: true,
		},
		{
			name:    "negative width",
			cfg:     GenConfig{Width: -1, Height: 64},
			wantErr: true,
		},
		{
			name:    "valid",
			cfg:     GenConfig{Width: 64, Height: 64},
			wantErr: false,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := c.cfg.Validate()
			if c.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !c.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestLoadGenConfig(t *testing.T) {
	fsys := fstest.MapFS{
		"gen-config.yaml": &fstest.MapFile{
			Data: []byte(`kind: gen-config
data:
  seed: 42
  width: 256
  height: 256
  octaves: 6
  lacunarity: 2.0
  gain: 0.5
  scale: 100.0
`),
		},
	}
	cfg, err := LoadGenConfig("gen-config.yaml", fsys)
	if err != nil {
		t.Fatalf("LoadGenConfig error: %v", err)
	}
	if cfg.Seed != 42 {
		t.Errorf("seed = %d, want 42", cfg.Seed)
	}
	if cfg.Width != 256 {
		t.Errorf("width = %d, want 256", cfg.Width)
	}
	if cfg.Height != 256 {
		t.Errorf("height = %d, want 256", cfg.Height)
	}
	if cfg.Octaves != 6 {
		t.Errorf("octaves = %d, want 6", cfg.Octaves)
	}
	if cfg.Lacunarity != 2.0 {
		t.Errorf("lacunarity = %v, want 2.0", cfg.Lacunarity)
	}
	if cfg.Gain != 0.5 {
		t.Errorf("gain = %v, want 0.5", cfg.Gain)
	}
	if cfg.Scale != 100.0 {
		t.Errorf("scale = %v, want 100.0", cfg.Scale)
	}
}

func TestLoadGenConfigInvalidYAML(t *testing.T) {
	fsys := fstest.MapFS{
		"bad.yaml": &fstest.MapFile{
			Data: []byte(`this is not: valid yaml: [`),
		},
	}
	_, err := LoadGenConfig("bad.yaml", fsys)
	if err == nil {
		t.Error("expected error for invalid YAML")
	}
}

func TestLoadGenConfigMissingFile(t *testing.T) {
	fsys := fstest.MapFS{}
	_, err := LoadGenConfig("missing.yaml", fsys)
	if err == nil {
		t.Error("expected error for missing file")
	}
}
