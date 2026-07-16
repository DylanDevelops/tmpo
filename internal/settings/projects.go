package settings

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/DylanDevelops/tmpo/internal/fsperm"
	"go.yaml.in/yaml/v3"
)

// GlobalProject represents a global project configuration
type GlobalProject struct {
	Name        string   `yaml:"name"`
	HourlyRate  *float64 `yaml:"hourly_rate,omitempty"`
	Description string   `yaml:"description,omitempty"`
	ExportPath  string   `yaml:"export_path,omitempty"`
}

// ProjectsRegistry holds all global projects
type ProjectsRegistry struct {
	Projects []GlobalProject `yaml:"projects"`
}

// GetProjectsPath returns the path to the global projects registry file
func GetProjectsPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	tmpoDir := filepath.Join(home, ".tmpo")
	if devMode := os.Getenv("TMPO_DEV"); devMode == "1" || devMode == "true" {
		tmpoDir = filepath.Join(home, ".tmpo-dev")
	}

	return filepath.Join(tmpoDir, "projects.yaml"), nil
}

// LoadProjects loads the global projects registry
func LoadProjects() (*ProjectsRegistry, error) {
	projectsPath, err := GetProjectsPath()
	if err != nil {
		return &ProjectsRegistry{Projects: []GlobalProject{}}, nil
	}

	// if does not exist return empty registry list
	if _, err := os.Stat(projectsPath); os.IsNotExist(err) {
		return &ProjectsRegistry{Projects: []GlobalProject{}}, nil
	}

	data, err := os.ReadFile(projectsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read projects registry: %w", err)
	}

	var registry ProjectsRegistry
	if err := yaml.Unmarshal(data, &registry); err != nil {
		return nil, fmt.Errorf("failed to parse projects registry at %s: %w (check file syntax)", projectsPath, err)
	}

	if registry.Projects == nil {
		registry.Projects = []GlobalProject{}
	}

	return &registry, nil
}

// Save saves the projects registry to disk
func (pr *ProjectsRegistry) Save() error {
	projectsPath, err := GetProjectsPath()
	if err != nil {
		return err
	}

	tmpoDir := filepath.Dir(projectsPath)
	if err := fsperm.SecureDir(tmpoDir); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := yaml.Marshal(pr)
	if err != nil {
		return fmt.Errorf("failed to marshal projects registry: %w", err)
	}

	if err := os.WriteFile(projectsPath, data, fsperm.FilePerm); err != nil {
		return fmt.Errorf("failed to write projects registry: %w", err)
	}

	if err := fsperm.SecureFile(projectsPath); err != nil {
		return err
	}

	return nil
}

// GetProject retrieves a project by name
func (pr *ProjectsRegistry) GetProject(name string) (*GlobalProject, error) {
	normalizedName := strings.TrimSpace(name)
	if normalizedName == "" {
		return nil, fmt.Errorf("project name cannot be empty")
	}

	for i := range pr.Projects {
		if strings.EqualFold(pr.Projects[i].Name, normalizedName) {
			return &pr.Projects[i], nil
		}
	}

	return nil, fmt.Errorf("project '%s' not found in global registry", name)
}

// AddProject adds a new project to the registry
func (pr *ProjectsRegistry) AddProject(project GlobalProject) error {
	normalizedName := strings.TrimSpace(project.Name)
	if normalizedName == "" {
		return fmt.Errorf("project name cannot be empty")
	}

	// does project exist
	if _, err := pr.GetProject(normalizedName); err == nil {
		return fmt.Errorf("project '%s' already exists", normalizedName)
	}

	// Normalize the name in the project
	project.Name = normalizedName

	pr.Projects = append(pr.Projects, project)

	return nil
}

// UpdateProject updates an existing project in the registry
func (pr *ProjectsRegistry) UpdateProject(name string, updatedProject GlobalProject) error {
	normalizedName := strings.TrimSpace(name)
	if normalizedName == "" {
		return fmt.Errorf("project name cannot be empty")
	}

	for i := range pr.Projects {
		if strings.EqualFold(pr.Projects[i].Name, normalizedName) {
			// preserve original name
			if updatedProject.Name == "" {
				updatedProject.Name = pr.Projects[i].Name
			}
			pr.Projects[i] = updatedProject
			return nil
		}
	}

	return fmt.Errorf("project '%s' not found", name)
}

// DeleteProject removes a project from the registry
func (pr *ProjectsRegistry) DeleteProject(name string) error {
	normalizedName := strings.TrimSpace(name)
	if normalizedName == "" {
		return fmt.Errorf("project name cannot be empty")
	}

	for i := range pr.Projects {
		if strings.EqualFold(pr.Projects[i].Name, normalizedName) {
			pr.Projects = append(pr.Projects[:i], pr.Projects[i+1:]...)
			return nil
		}
	}

	return fmt.Errorf("project '%s' not found", name)
}

// ListProjects returns all projects in the registry
func (pr *ProjectsRegistry) ListProjects() []GlobalProject {
	return pr.Projects
}

// Exists checks if a project exists in the registry (case-insensitive)
func (pr *ProjectsRegistry) Exists(name string) bool {
	_, err := pr.GetProject(name)
	return err == nil
}
