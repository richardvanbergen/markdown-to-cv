package init

import (
	"context"
	"errors"
	"os"
	"path/filepath"

	"github.com/richq/m2cv/internal/config"
	"github.com/richq/m2cv/internal/executor"
)

// ErrAlreadyInitialized is returned when attempting to initialize
// a project that already has an m2cv.yml file.
var ErrAlreadyInitialized = errors.New("project already initialized")

// Service orchestrates m2cv project initialization.
// It coordinates config file creation, npm package installation,
// and ensures proper project setup.
type Service struct {
	configRepo  config.Repository
	npmExecutor executor.NPMExecutor
}

// InitOptions contains options for initializing a new m2cv project.
type InitOptions struct {
	// ProjectDir is the directory where the project will be initialized.
	ProjectDir string

	// BaseCVPath is the path to the base CV markdown file.
	BaseCVPath string

	// Theme is the JSON Resume theme to use.
	Theme string

	// DefaultModel is the default Claude model for optimization.
	DefaultModel string
}

// NewService creates a new init service with the given dependencies.
func NewService(configRepo config.Repository, npm executor.NPMExecutor) *Service {
	return &Service{
		configRepo:  configRepo,
		npmExecutor: npm,
	}
}

// Init initializes a new m2cv project in the specified directory.
// It performs the following steps:
// 1. Check if m2cv.yml already exists (fail if so)
// 2. Run npm init if no package.json exists
// 3. Install resumed and the selected theme package
// 4. Create and save the m2cv.yml config file
func (s *Service) Init(ctx context.Context, opts InitOptions) error {
	// 1. Check if already initialized
	configPath := filepath.Join(opts.ProjectDir, "m2cv.yml")
	if _, err := os.Stat(configPath); err == nil {
		return ErrAlreadyInitialized
	}

	// 2. Run npm init if no package.json exists
	pkgPath := filepath.Join(opts.ProjectDir, "package.json")
	if _, err := os.Stat(pkgPath); os.IsNotExist(err) {
		if err := s.npmExecutor.Init(ctx, opts.ProjectDir); err != nil {
			return err
		}
	}

	// 3. Install resumed and theme package
	themePackage := ThemePackageName(opts.Theme)
	if err := s.npmExecutor.Install(ctx, opts.ProjectDir, "resumed", themePackage); err != nil {
		return err
	}

	// 4. Create and save config
	cfg := &config.Config{
		BaseCVPath:   opts.BaseCVPath,
		DefaultTheme: opts.Theme,
		Themes:       []string{opts.Theme},
		DefaultModel: opts.DefaultModel,
	}

	if err := s.configRepo.Save(configPath, cfg); err != nil {
		return err
	}

	return nil
}
