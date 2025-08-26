package utils

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// CreateFileFromTmpl executes a template and writes it to a file.
func CreateFileFromTmpl(path, tmplStr string, data map[string]string) {
	tmpl, err := template.New("").Funcs(template.FuncMap{"Title": strings.Title}).Parse(tmplStr)
	if err != nil {
		log.Fatalf("Failed to parse template for %s: %v", path, err)
	}

	f, err := os.Create(path)
	if err != nil {
		log.Fatalf("Failed to create file %s: %v", path, err)
	}
	defer f.Close()

	if err := tmpl.Execute(f, data); err != nil {
		log.Fatalf("Failed to execute template for %s: %v", path, err)
	}
}

// FindProjectRoot finds the root of the Grob project by looking for a go.mod file.
func FindProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		if dir == filepath.Dir(dir) {
			return "", fmt.Errorf("go.mod not found in any parent directory")
		}
		dir = filepath.Dir(dir)
	}
}

// GetProjectName reads the module name from the go.mod file.
func GetProjectName(projectRoot string) string {
	goModBytes, err := os.ReadFile(filepath.Join(projectRoot, "go.mod"))
	if err != nil {
		log.Fatalf("Could not read go.mod: %v", err)
	}
	return strings.Split(strings.Split(string(goModBytes), "\n")[0], " ")[1]
}
