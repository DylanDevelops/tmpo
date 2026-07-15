package settings

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/DylanDevelops/tmpo/internal/fsperm"
	"github.com/stretchr/testify/assert"
)

func TestLoadProjects(t *testing.T) {
	tmpDir := t.TempDir()

	t.Setenv("TMPO_DEV", "0")
	t.Setenv("HOME", tmpDir)        // Unix/macOS
	t.Setenv("USERPROFILE", tmpDir) // Windows

	t.Run("loads empty registry when file doesn't exist", func(t *testing.T) {
		registry, err := LoadProjects()
		assert.NoError(t, err)
		assert.NotNil(t, registry)
		assert.Empty(t, registry.Projects)
	})

	t.Run("loads valid projects registry", func(t *testing.T) {
		// Create .tmpo directory
		tmpoDir := filepath.Join(tmpDir, ".tmpo")
		err := os.MkdirAll(tmpoDir, 0755)
		assert.NoError(t, err)

		projectsPath := filepath.Join(tmpoDir, "projects.yaml")
		rate1 := 100.0
		rate2 := 150.5
		content := `projects:
  - name: "Project Alpha"
    hourly_rate: 100.0
    description: "First project"
    export_path: "/tmp/alpha"
  - name: "Project Beta"
    hourly_rate: 150.5
    description: "Second project"
`
		err = os.WriteFile(projectsPath, []byte(content), 0644)
		assert.NoError(t, err)

		registry, err := LoadProjects()
		assert.NoError(t, err)
		assert.NotNil(t, registry)
		assert.Len(t, registry.Projects, 2)
		assert.Equal(t, "Project Alpha", registry.Projects[0].Name)
		assert.Equal(t, &rate1, registry.Projects[0].HourlyRate)
		assert.Equal(t, "First project", registry.Projects[0].Description)
		assert.Equal(t, "/tmp/alpha", registry.Projects[0].ExportPath)
		assert.Equal(t, "Project Beta", registry.Projects[1].Name)
		assert.Equal(t, &rate2, registry.Projects[1].HourlyRate)
	})

	t.Run("handles projects without optional fields", func(t *testing.T) {
		// Create .tmpo directory
		tmpoDir := filepath.Join(tmpDir, ".tmpo")
		err := os.MkdirAll(tmpoDir, 0755)
		assert.NoError(t, err)

		projectsPath := filepath.Join(tmpoDir, "projects.yaml")
		content := `projects:
  - name: "Minimal Project"
`
		err = os.WriteFile(projectsPath, []byte(content), 0644)
		assert.NoError(t, err)

		registry, err := LoadProjects()
		assert.NoError(t, err)
		assert.Len(t, registry.Projects, 1)
		assert.Equal(t, "Minimal Project", registry.Projects[0].Name)
		assert.Nil(t, registry.Projects[0].HourlyRate)
		assert.Empty(t, registry.Projects[0].Description)
		assert.Empty(t, registry.Projects[0].ExportPath)
	})

	t.Run("returns error for invalid YAML", func(t *testing.T) {
		tmpoDir := filepath.Join(tmpDir, ".tmpo")
		err := os.MkdirAll(tmpoDir, 0755)
		assert.NoError(t, err)

		projectsPath := filepath.Join(tmpoDir, "projects.yaml")
		content := `projects: [ invalid yaml
`
		err = os.WriteFile(projectsPath, []byte(content), 0644)
		assert.NoError(t, err)

		_, err = LoadProjects()
		assert.Error(t, err)
	})
}

func TestProjectsRegistrySave(t *testing.T) {
	tmpDir := t.TempDir()

	t.Setenv("TMPO_DEV", "0")
	t.Setenv("HOME", tmpDir)        // Unix/macOS
	t.Setenv("USERPROFILE", tmpDir) // Windows

	t.Run("saves registry successfully", func(t *testing.T) {
		rate := 125.0
		registry := &ProjectsRegistry{
			Projects: []GlobalProject{
				{
					Name:        "Test Project",
					HourlyRate:  &rate,
					Description: "Test description",
					ExportPath:  "/tmp/test",
				},
			},
		}

		err := registry.Save()
		assert.NoError(t, err)

		// Verify file was created
		projectsPath, _ := GetProjectsPath()
		_, err = os.Stat(projectsPath)
		assert.NoError(t, err)

		// Verify content can be loaded
		loaded, err := LoadProjects()
		assert.NoError(t, err)
		assert.Len(t, loaded.Projects, 1)
		assert.Equal(t, "Test Project", loaded.Projects[0].Name)
		assert.Equal(t, &rate, loaded.Projects[0].HourlyRate)
		assert.Equal(t, "Test description", loaded.Projects[0].Description)
		assert.Equal(t, "/tmp/test", loaded.Projects[0].ExportPath)
	})

	t.Run("creates directory if it doesn't exist", func(t *testing.T) {
		// Remove .tmpo directory if it exists
		tmpoDir := filepath.Join(tmpDir, ".tmpo")
		os.RemoveAll(tmpoDir)

		registry := &ProjectsRegistry{
			Projects: []GlobalProject{
				{Name: "New Project"},
			},
		}

		err := registry.Save()
		assert.NoError(t, err)

		// Verify directory was created
		_, err = os.Stat(tmpoDir)
		assert.NoError(t, err)
	})
}

func TestGetProject(t *testing.T) {
	rate1 := 100.0
	rate2 := 150.0
	registry := &ProjectsRegistry{
		Projects: []GlobalProject{
			{Name: "Project Alpha", HourlyRate: &rate1},
			{Name: "Project Beta", HourlyRate: &rate2},
			{Name: "lowercase", HourlyRate: nil},
		},
	}

	t.Run("finds project by exact name", func(t *testing.T) {
		project, err := registry.GetProject("Project Alpha")
		assert.NoError(t, err)
		assert.NotNil(t, project)
		assert.Equal(t, "Project Alpha", project.Name)
		assert.Equal(t, &rate1, project.HourlyRate)
	})

	t.Run("finds project case-insensitively", func(t *testing.T) {
		project, err := registry.GetProject("project beta")
		assert.NoError(t, err)
		assert.NotNil(t, project)
		assert.Equal(t, "Project Beta", project.Name)
	})

	t.Run("finds project with different casing", func(t *testing.T) {
		project, err := registry.GetProject("LOWERCASE")
		assert.NoError(t, err)
		assert.NotNil(t, project)
		assert.Equal(t, "lowercase", project.Name)
	})

	t.Run("handles whitespace in name", func(t *testing.T) {
		project, err := registry.GetProject("  Project Alpha  ")
		assert.NoError(t, err)
		assert.NotNil(t, project)
		assert.Equal(t, "Project Alpha", project.Name)
	})

	t.Run("returns error for non-existent project", func(t *testing.T) {
		_, err := registry.GetProject("Non-existent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("returns error for empty name", func(t *testing.T) {
		_, err := registry.GetProject("")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot be empty")
	})
}

func TestAddProject(t *testing.T) {
	t.Run("adds new project successfully", func(t *testing.T) {
		registry := &ProjectsRegistry{Projects: []GlobalProject{}}
		rate := 125.0
		newProject := GlobalProject{
			Name:        "New Project",
			HourlyRate:  &rate,
			Description: "Description",
		}

		err := registry.AddProject(newProject)
		assert.NoError(t, err)
		assert.Len(t, registry.Projects, 1)
		assert.Equal(t, "New Project", registry.Projects[0].Name)
		assert.Equal(t, &rate, registry.Projects[0].HourlyRate)
	})

	t.Run("normalizes project name", func(t *testing.T) {
		registry := &ProjectsRegistry{Projects: []GlobalProject{}}
		newProject := GlobalProject{Name: "  Spaced Name  "}

		err := registry.AddProject(newProject)
		assert.NoError(t, err)
		assert.Equal(t, "Spaced Name", registry.Projects[0].Name)
	})

	t.Run("returns error for duplicate project", func(t *testing.T) {
		registry := &ProjectsRegistry{
			Projects: []GlobalProject{
				{Name: "Existing Project"},
			},
		}

		newProject := GlobalProject{Name: "Existing Project"}
		err := registry.AddProject(newProject)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already exists")
	})

	t.Run("detects duplicates case-insensitively", func(t *testing.T) {
		registry := &ProjectsRegistry{
			Projects: []GlobalProject{
				{Name: "Existing Project"},
			},
		}

		newProject := GlobalProject{Name: "existing project"}
		err := registry.AddProject(newProject)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already exists")
	})

	t.Run("returns error for empty name", func(t *testing.T) {
		registry := &ProjectsRegistry{Projects: []GlobalProject{}}
		newProject := GlobalProject{Name: ""}

		err := registry.AddProject(newProject)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot be empty")
	})

	t.Run("returns error for whitespace-only name", func(t *testing.T) {
		registry := &ProjectsRegistry{Projects: []GlobalProject{}}
		newProject := GlobalProject{Name: "   "}

		err := registry.AddProject(newProject)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot be empty")
	})
}

func TestUpdateProject(t *testing.T) {
	rate1 := 100.0
	rate2 := 200.0

	t.Run("updates existing project", func(t *testing.T) {
		registry := &ProjectsRegistry{
			Projects: []GlobalProject{
				{Name: "Original", HourlyRate: &rate1},
			},
		}

		updatedProject := GlobalProject{
			Name:        "Original",
			HourlyRate:  &rate2,
			Description: "Updated description",
		}

		err := registry.UpdateProject("Original", updatedProject)
		assert.NoError(t, err)
		assert.Len(t, registry.Projects, 1)
		assert.Equal(t, "Original", registry.Projects[0].Name)
		assert.Equal(t, &rate2, registry.Projects[0].HourlyRate)
		assert.Equal(t, "Updated description", registry.Projects[0].Description)
	})

	t.Run("finds project case-insensitively", func(t *testing.T) {
		registry := &ProjectsRegistry{
			Projects: []GlobalProject{
				{Name: "CamelCase", HourlyRate: &rate1},
			},
		}

		updatedProject := GlobalProject{Name: "CamelCase", HourlyRate: &rate2}
		err := registry.UpdateProject("camelcase", updatedProject)
		assert.NoError(t, err)
		assert.Equal(t, &rate2, registry.Projects[0].HourlyRate)
	})

	t.Run("returns error for non-existent project", func(t *testing.T) {
		registry := &ProjectsRegistry{Projects: []GlobalProject{}}
		updatedProject := GlobalProject{Name: "NonExistent"}

		err := registry.UpdateProject("NonExistent", updatedProject)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("returns error for empty name", func(t *testing.T) {
		registry := &ProjectsRegistry{Projects: []GlobalProject{}}
		updatedProject := GlobalProject{Name: "Test"}

		err := registry.UpdateProject("", updatedProject)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot be empty")
	})
}

func TestDeleteProject(t *testing.T) {
	rate1 := 100.0
	rate2 := 150.0

	t.Run("deletes existing project", func(t *testing.T) {
		registry := &ProjectsRegistry{
			Projects: []GlobalProject{
				{Name: "Project 1", HourlyRate: &rate1},
				{Name: "Project 2", HourlyRate: &rate2},
			},
		}

		err := registry.DeleteProject("Project 1")
		assert.NoError(t, err)
		assert.Len(t, registry.Projects, 1)
		assert.Equal(t, "Project 2", registry.Projects[0].Name)
	})

	t.Run("finds project case-insensitively", func(t *testing.T) {
		registry := &ProjectsRegistry{
			Projects: []GlobalProject{
				{Name: "DeleteMe", HourlyRate: &rate1},
			},
		}

		err := registry.DeleteProject("deleteme")
		assert.NoError(t, err)
		assert.Empty(t, registry.Projects)
	})

	t.Run("returns error for non-existent project", func(t *testing.T) {
		registry := &ProjectsRegistry{Projects: []GlobalProject{}}

		err := registry.DeleteProject("NonExistent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("returns error for empty name", func(t *testing.T) {
		registry := &ProjectsRegistry{Projects: []GlobalProject{}}

		err := registry.DeleteProject("")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot be empty")
	})
}

func TestListProjects(t *testing.T) {
	rate1 := 100.0
	rate2 := 150.0

	t.Run("lists all projects", func(t *testing.T) {
		registry := &ProjectsRegistry{
			Projects: []GlobalProject{
				{Name: "Project 1", HourlyRate: &rate1},
				{Name: "Project 2", HourlyRate: &rate2},
			},
		}

		projects := registry.ListProjects()
		assert.Len(t, projects, 2)
		assert.Equal(t, "Project 1", projects[0].Name)
		assert.Equal(t, "Project 2", projects[1].Name)
	})

	t.Run("returns empty list for empty registry", func(t *testing.T) {
		registry := &ProjectsRegistry{Projects: []GlobalProject{}}

		projects := registry.ListProjects()
		assert.Empty(t, projects)
	})
}

func TestExists(t *testing.T) {
	registry := &ProjectsRegistry{
		Projects: []GlobalProject{
			{Name: "Existing Project"},
		},
	}

	t.Run("returns true for existing project", func(t *testing.T) {
		exists := registry.Exists("Existing Project")
		assert.True(t, exists)
	})

	t.Run("returns true case-insensitively", func(t *testing.T) {
		exists := registry.Exists("existing project")
		assert.True(t, exists)
	})

	t.Run("returns false for non-existent project", func(t *testing.T) {
		exists := registry.Exists("Non-existent")
		assert.False(t, exists)
	})
}

func TestGetProjectsPath(t *testing.T) {
	tmpDir := t.TempDir()

	t.Setenv("TMPO_DEV", "0")
	t.Setenv("HOME", tmpDir)        // Unix/macOS
	t.Setenv("USERPROFILE", tmpDir) // Windows

	t.Run("returns correct path", func(t *testing.T) {
		path, err := GetProjectsPath()
		assert.NoError(t, err)
		assert.Contains(t, path, tmpDir)
	})

	t.Run("path ends with projects.yaml", func(t *testing.T) {
		path, err := GetProjectsPath()
		assert.NoError(t, err)
		assert.Equal(t, "projects.yaml", filepath.Base(path))
	})
}

func TestProjectsRegistrySave_SetsPrivatePermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX permissions are not enforced on Windows")
	}

	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("USERPROFILE", tmpHome)
	t.Setenv("TMPO_DEV", "")

	reg := &ProjectsRegistry{Projects: []GlobalProject{}}
	assert.NoError(t, reg.Save())

	path, err := GetProjectsPath()
	assert.NoError(t, err)

	fileInfo, err := os.Stat(path)
	assert.NoError(t, err)
	assert.Equal(t, fsperm.FilePerm, fileInfo.Mode().Perm())
}
