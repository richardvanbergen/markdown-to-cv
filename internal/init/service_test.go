package init

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/richq/m2cv/internal/config"
)

// mockConfigRepository tracks Save and Load calls for testing.
type mockConfigRepository struct {
	savedConfig *config.Config
	savedPath   string
	saveErr     error
	loadConfig  *config.Config
	loadErr     error
	findPath    string
	findErr     error
}

func (m *mockConfigRepository) Save(configPath string, cfg *config.Config) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.savedPath = configPath
	m.savedConfig = cfg
	return nil
}

func (m *mockConfigRepository) Load(configPath string) (*config.Config, error) {
	if m.loadErr != nil {
		return nil, m.loadErr
	}
	return m.loadConfig, nil
}

func (m *mockConfigRepository) Find(startDir string) (string, error) {
	if m.findErr != nil {
		return "", m.findErr
	}
	return m.findPath, nil
}

// mockNPMExecutor tracks npm command calls for testing.
type mockNPMExecutor struct {
	initCalled     bool
	initDir        string
	initErr        error
	installCalled  bool
	installDir     string
	installPkgs    []string
	installErr     error
	checkInstalled bool
	checkErr       error
}

func (m *mockNPMExecutor) Init(ctx context.Context, dir string) error {
	m.initCalled = true
	m.initDir = dir
	return m.initErr
}

func (m *mockNPMExecutor) Install(ctx context.Context, dir string, packages ...string) error {
	m.installCalled = true
	m.installDir = dir
	m.installPkgs = packages
	return m.installErr
}

func (m *mockNPMExecutor) CheckInstalled(ctx context.Context, dir string, pkg string) (bool, error) {
	return m.checkInstalled, m.checkErr
}

func TestService_Init_Success(t *testing.T) {
	// Setup temp directory (no package.json, no m2cv.yml)
	tmpDir, err := os.MkdirTemp("", "m2cv-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	configRepo := &mockConfigRepository{}
	npmExec := &mockNPMExecutor{}

	svc := NewService(configRepo, npmExec)

	opts := InitOptions{
		ProjectDir:   tmpDir,
		BaseCVPath:   "./cv.md",
		Theme:        "even",
		DefaultModel: "sonnet",
	}

	err = svc.Init(context.Background(), opts)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// Verify npm init was called (no package.json)
	if !npmExec.initCalled {
		t.Error("Expected npm init to be called")
	}
	if npmExec.initDir != tmpDir {
		t.Errorf("npm init called with dir %q, want %q", npmExec.initDir, tmpDir)
	}

	// Verify npm install was called with correct packages
	if !npmExec.installCalled {
		t.Error("Expected npm install to be called")
	}
	if len(npmExec.installPkgs) != 2 {
		t.Errorf("Expected 2 packages, got %d", len(npmExec.installPkgs))
	}
	if npmExec.installPkgs[0] != "resumed" {
		t.Errorf("Expected first package 'resumed', got %q", npmExec.installPkgs[0])
	}
	if npmExec.installPkgs[1] != "jsonresume-theme-even" {
		t.Errorf("Expected second package 'jsonresume-theme-even', got %q", npmExec.installPkgs[1])
	}

	// Verify config was saved
	if configRepo.savedConfig == nil {
		t.Fatal("Expected config to be saved")
	}
	if configRepo.savedConfig.BaseCVPath != "./cv.md" {
		t.Errorf("BaseCVPath = %q, want %q", configRepo.savedConfig.BaseCVPath, "./cv.md")
	}
	if configRepo.savedConfig.DefaultTheme != "even" {
		t.Errorf("DefaultTheme = %q, want %q", configRepo.savedConfig.DefaultTheme, "even")
	}
	if len(configRepo.savedConfig.Themes) != 1 || configRepo.savedConfig.Themes[0] != "even" {
		t.Errorf("Themes = %v, want [even]", configRepo.savedConfig.Themes)
	}
	if configRepo.savedConfig.DefaultModel != "sonnet" {
		t.Errorf("DefaultModel = %q, want %q", configRepo.savedConfig.DefaultModel, "sonnet")
	}

	expectedPath := filepath.Join(tmpDir, "m2cv.yml")
	if configRepo.savedPath != expectedPath {
		t.Errorf("Config saved to %q, want %q", configRepo.savedPath, expectedPath)
	}
}

func TestService_Init_ExistingConfig(t *testing.T) {
	// Setup temp directory with existing m2cv.yml
	tmpDir, err := os.MkdirTemp("", "m2cv-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create existing m2cv.yml
	configPath := filepath.Join(tmpDir, "m2cv.yml")
	if err := os.WriteFile(configPath, []byte("existing: config"), 0644); err != nil {
		t.Fatalf("Failed to create existing config: %v", err)
	}

	configRepo := &mockConfigRepository{}
	npmExec := &mockNPMExecutor{}

	svc := NewService(configRepo, npmExec)

	opts := InitOptions{
		ProjectDir:   tmpDir,
		BaseCVPath:   "./cv.md",
		Theme:        "even",
		DefaultModel: "sonnet",
	}

	err = svc.Init(context.Background(), opts)
	if err != ErrAlreadyInitialized {
		t.Errorf("Expected ErrAlreadyInitialized, got %v", err)
	}

	// Verify no npm commands were called
	if npmExec.initCalled {
		t.Error("npm init should not be called when config exists")
	}
	if npmExec.installCalled {
		t.Error("npm install should not be called when config exists")
	}
}

func TestService_Init_ExistingPackageJson(t *testing.T) {
	// Setup temp directory with existing package.json (but no m2cv.yml)
	tmpDir, err := os.MkdirTemp("", "m2cv-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create existing package.json
	pkgPath := filepath.Join(tmpDir, "package.json")
	if err := os.WriteFile(pkgPath, []byte(`{"name":"test"}`), 0644); err != nil {
		t.Fatalf("Failed to create package.json: %v", err)
	}

	configRepo := &mockConfigRepository{}
	npmExec := &mockNPMExecutor{}

	svc := NewService(configRepo, npmExec)

	opts := InitOptions{
		ProjectDir:   tmpDir,
		BaseCVPath:   "./cv.md",
		Theme:        "elegant",
		DefaultModel: "sonnet",
	}

	err = svc.Init(context.Background(), opts)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// Verify npm init was NOT called (package.json exists)
	if npmExec.initCalled {
		t.Error("npm init should not be called when package.json exists")
	}

	// Verify npm install was still called
	if !npmExec.installCalled {
		t.Error("Expected npm install to be called")
	}
	if npmExec.installPkgs[1] != "jsonresume-theme-elegant" {
		t.Errorf("Expected theme package 'jsonresume-theme-elegant', got %q", npmExec.installPkgs[1])
	}
}

func TestService_Init_NPMInstallFailed(t *testing.T) {
	// Setup temp directory
	tmpDir, err := os.MkdirTemp("", "m2cv-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	configRepo := &mockConfigRepository{}
	npmExec := &mockNPMExecutor{
		installErr: os.ErrPermission, // Simulate install failure
	}

	svc := NewService(configRepo, npmExec)

	opts := InitOptions{
		ProjectDir:   tmpDir,
		BaseCVPath:   "./cv.md",
		Theme:        "even",
		DefaultModel: "sonnet",
	}

	err = svc.Init(context.Background(), opts)
	if err == nil {
		t.Error("Expected error when npm install fails")
	}

	// Verify config was NOT saved (install failed first)
	if configRepo.savedConfig != nil {
		t.Error("Config should not be saved when npm install fails")
	}
}

func TestService_Init_NPMInitFailed(t *testing.T) {
	// Setup temp directory (no package.json)
	tmpDir, err := os.MkdirTemp("", "m2cv-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	configRepo := &mockConfigRepository{}
	npmExec := &mockNPMExecutor{
		initErr: os.ErrPermission, // Simulate init failure
	}

	svc := NewService(configRepo, npmExec)

	opts := InitOptions{
		ProjectDir:   tmpDir,
		BaseCVPath:   "./cv.md",
		Theme:        "even",
		DefaultModel: "sonnet",
	}

	err = svc.Init(context.Background(), opts)
	if err == nil {
		t.Error("Expected error when npm init fails")
	}

	// Verify install was NOT called (init failed first)
	if npmExec.installCalled {
		t.Error("npm install should not be called when npm init fails")
	}

	// Verify config was NOT saved
	if configRepo.savedConfig != nil {
		t.Error("Config should not be saved when npm init fails")
	}
}

func TestNewService(t *testing.T) {
	configRepo := &mockConfigRepository{}
	npmExec := &mockNPMExecutor{}

	svc := NewService(configRepo, npmExec)

	if svc == nil {
		t.Fatal("NewService returned nil")
	}

	// Verify dependencies are set (through behavior testing)
	// The service should work with the provided mocks
	tmpDir, err := os.MkdirTemp("", "m2cv-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	opts := InitOptions{
		ProjectDir:   tmpDir,
		Theme:        "flat",
		DefaultModel: "sonnet",
	}

	// Should not panic and should call the mocks
	_ = svc.Init(context.Background(), opts)

	if !npmExec.initCalled && !npmExec.installCalled {
		t.Error("Service should use provided executors")
	}
}
