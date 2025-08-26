package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/spf13/cobra"
)

// --- Template Definitions (Corrected) ---

var goModTmpl = `module {{.ProjectName}}

go 1.19

require (
	github.com/gin-gonic/gin v1.8.1
	github.com/yuliussmayoru/grob-framework v0.1.0
	go.uber.org/dig v1.15.0
)
`

var gitignoreTmpl = `
# Binaries for programs and plugins
*.exe
*.exe~
*.dll
*.so
*.dylib
*.test
*.out
.idea/
`

var internalMainTmpl = `package main

import (
	"log"
	"sync"
)

// AppRunner defines the interface for a runnable application.
type AppRunner interface {
    Run()
}

func main() {
    apps := map[string]AppRunner{}

    var wg sync.WaitGroup

    if len(apps) == 0 {
        log.Println("No applications to run. Use 'grob create-app <app-name>' to create one.")
        return
    }

    for name, app := range apps {
        wg.Add(1)
        
        go func(appName string, runner AppRunner) {
            defer wg.Done()
            log.Printf("Starting application: %s", appName)
            runner.Run()
        }(name, app)
    }

    log.Println("All applications are starting...")
    wg.Wait()
    log.Println("All applications have been shut down.")
}
`

var appMainTmpl = `package {{.AppName}}

import (
	"{{.ProjectName}}/internal/{{.AppName}}/core"
)

// App struct holds the application instance.
type App struct{}

// Run initializes and starts the web application.
func (a App) Run() {
	// TODO: Make port configurable
	port := ":8081" 
	
	app := core.New()

	// Example of creating a route group for this app
	// api := app.Router().Group("/api/{{.AppName}}")
	// You would then invoke controllers to register their routes with this group.

	app.Start(port)
}
`

var moduleTmpl = `package {{.ModuleName}}

import (
	"{{.ProjectName}}/internal/{{.AppName}}/core"
	"go.uber.org/dig"
)

// {{.ModuleName | Title}}Module implements the framework.Module interface.
type {{.ModuleName | Title}}Module struct{}

// Register provides the components of this module to the dependency injection container.
func (m {{.ModuleName | Title}}Module) Register(container *dig.Container) error {
	// Provide the Service
	if err := container.Provide(New{{.ModuleName | Title}}Service); err != nil {
		return err
	}

	// Provide the Controller
	if err := container.Provide(New{{.ModuleName | Title}}Controller); err != nil {
		return err
	}

	return nil
}
`

var serviceTmpl = `package {{.ModuleName}}

import "log"

// {{.ModuleName | Title}}Service defines the business logic for the {{.ModuleName}} module.
type {{.ModuleName | Title}}Service struct {
	// Add dependencies here, e.g., a database connection
}

// New{{.ModuleName | Title}}Service creates a new service instance.
func New{{.ModuleName | Title}}Service() *{{.ModuleName | Title}}Service {
	return &{{.ModuleName | Title}}Service{}
}

// ExampleMethod is an example of a service method.
func (s *{{.ModuleName | Title}}Service) ExampleMethod() string {
	log.Println("{{.ModuleName | Title}}Service: ExampleMethod called")
	return "Hello from {{.ModuleName | Title}}Service!"
}
`

var controllerTmpl = `package {{.ModuleName}}

import (
	"net/http"
	"github.com/gin-gonic/gin"
)

// {{.ModuleName | Title}}Controller handles the HTTP requests for the {{.ModuleName}} module.
type {{.ModuleName | Title}}Controller struct {
	service *{{.ModuleName | Title}}Service
}

// New{{.ModuleName | Title}}Controller creates a new controller with its dependencies.
func New{{.ModuleName | Title}}Controller(service *{{.ModuleName | Title}}Service) *{{.ModuleName | Title}}Controller {
	return &{{.ModuleName | Title}}Controller{service: service}
}

// RegisterRoutes sets up the routes for this controller.
// Note: In a real app, you'd invoke this method to connect routes to the main app router.
func (c *{{.ModuleName | Title}}Controller) RegisterRoutes(router *gin.RouterGroup) {
	router.GET("/", c.GetExample)
}

// GetExample is an example handler function.
func (c *{{.ModuleName | Title}}Controller) GetExample(ctx *gin.Context) {
	message := c.service.ExampleMethod()
	ctx.JSON(http.StatusOK, gin.H{"message": message})
}
`

// --- Main CLI Logic ---

func main() {
	var rootCmd = &cobra.Command{Use: "grob"}

	var newCmd = &cobra.Command{
		Use:   "new [project-name]",
		Short: "Create a new Grob project",
		Args:  cobra.MinimumNArgs(1),
		Run:   newProject,
	}

	var createAppCmd = &cobra.Command{
		Use:   "create-app [app-name]",
		Short: "Create a new web application inside a Grob project",
		Args:  cobra.MinimumNArgs(1),
		Run:   createApp,
	}

	var createModuleCmd = &cobra.Command{
		Use:   "create-module [app-name] [module-name]",
		Short: "Create a new module within a web application",
		Args:  cobra.MinimumNArgs(2),
		Run:   createModule,
	}

	rootCmd.AddCommand(newCmd, createAppCmd, createModuleCmd)
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

// --- Command Implementations ---

func newProject(cmd *cobra.Command, args []string) {
	projectName := args[0]
	log.Printf("Creating new project: %s", projectName)

	if err := os.Mkdir(projectName, 0755); err != nil {
		log.Fatalf("Failed to create project directory: %v", err)
	}

	dirs := []string{
		filepath.Join(projectName, "internal"),
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Fatalf("Failed to create directory %s: %v", dir, err)
		}
	}

	createFileFromTmpl(filepath.Join(projectName, "go.mod"), goModTmpl, map[string]string{"ProjectName": projectName})
	createFileFromTmpl(filepath.Join(projectName, ".gitignore"), gitignoreTmpl, nil)
	createFileFromTmpl(filepath.Join(projectName, "internal", "main.go"), internalMainTmpl, nil)

	log.Printf("Project '%s' created successfully.", projectName)
	log.Println("Next steps:")
	log.Printf("  cd %s", projectName)
	log.Println("  go mod tidy  # To download the framework")
	log.Println("  grob create-app myapp")
}

func createApp(cmd *cobra.Command, args []string) {
	appName := args[0]
	log.Printf("Creating new application: %s", appName)

	projectRoot, err := findProjectRoot()
	if err != nil {
		log.Fatalf("Error: %v. Make sure you are inside a Grob project.", err)
	}
	projectName := getProjectName(projectRoot)

	appDir := filepath.Join(projectRoot, "internal", appName)
	if err := os.Mkdir(appDir, 0755); err != nil {
		log.Fatalf("Failed to create app directory: %v", err)
	}

	coreDir := filepath.Join(appDir, "core")
	if err := os.Mkdir(coreDir, 0755); err != nil {
		log.Fatalf("Failed to create app core directory: %v", err)
	}

	coreFileContent := `package core
import "github.com/yuliussmayoru/grob-framework/pkg/framework"

// Re-export the framework types to make them local to the app
type App = framework.App
type Module = framework.Module
var New = framework.New
`
	if err := os.WriteFile(filepath.Join(coreDir, "core.go"), []byte(coreFileContent), 0644); err != nil {
		log.Fatalf("Failed to create core.go: %v", err)
	}

	appMainPath := filepath.Join(appDir, fmt.Sprintf("%s_main.go", appName))
	createFileFromTmpl(appMainPath, appMainTmpl, map[string]string{
		"ProjectName": projectName,
		"AppName":     appName,
	})

	internalMainPath := filepath.Join(projectRoot, "internal", "main.go")
	if err := addAppToInternalMain(internalMainPath, projectName, appName); err != nil {
		log.Fatalf("Failed to auto-register app: %v", err)
	}

	log.Printf("Application '%s' created and registered successfully.", appName)
}

func createModule(cmd *cobra.Command, args []string) {
	appName := args[0]
	moduleName := args[1]
	log.Printf("Creating new module '%s' in app '%s'", moduleName, appName)

	projectRoot, err := findProjectRoot()
	if err != nil {
		log.Fatalf("Error: %v. Make sure you are inside a Grob project.", err)
	}
	projectName := getProjectName(projectRoot)

	moduleDir := filepath.Join(projectRoot, "internal", appName, moduleName)
	if err := os.Mkdir(moduleDir, 0755); err != nil {
		log.Fatalf("Failed to create module directory: %v", err)
	}

	data := map[string]string{
		"ProjectName": projectName,
		"AppName":     appName,
		"ModuleName":  moduleName,
	}
	createFileFromTmpl(filepath.Join(moduleDir, fmt.Sprintf("%s.module.go", moduleName)), moduleTmpl, data)
	createFileFromTmpl(filepath.Join(moduleDir, fmt.Sprintf("%s.service.go", moduleName)), serviceTmpl, data)
	createFileFromTmpl(filepath.Join(moduleDir, fmt.Sprintf("%s.controller.go", moduleName)), controllerTmpl, data)

	appMainPath := filepath.Join(projectRoot, "internal", appName, fmt.Sprintf("%s_main.go", appName))
	if err := addModuleToAppMain(appMainPath, projectName, appName, moduleName); err != nil {
		log.Fatalf("Failed to auto-register module: %v", err)
	}

	log.Printf("Module '%s' created and registered successfully in app '%s'.", moduleName, appName)
}

// --- Helper Functions ---

func createFileFromTmpl(path, tmplStr string, data map[string]string) {
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

func findProjectRoot() (string, error) {
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

func getProjectName(projectRoot string) string {
	goModBytes, err := os.ReadFile(filepath.Join(projectRoot, "go.mod"))
	if err != nil {
		log.Fatalf("Could not read go.mod: %v", err)
	}
	return strings.Split(strings.Split(string(goModBytes), "\n")[0], " ")[1]
}

func addAppToInternalMain(path, projectName, appName string) error {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		return err
	}

	ast.Inspect(node, func(n ast.Node) bool {
		if gd, ok := n.(*ast.GenDecl); ok && gd.Tok == token.IMPORT {
			newImport := &ast.ImportSpec{
				Path: &ast.BasicLit{
					Kind:  token.STRING,
					Value: fmt.Sprintf(`"%s/internal/%s"`, projectName, appName),
				},
			}
			gd.Specs = append(gd.Specs, newImport)
			return false
		}
		return true
	})

	ast.Inspect(node, func(n ast.Node) bool {
		if cl, ok := n.(*ast.CompositeLit); ok {
			if kv, ok := cl.Type.(*ast.MapType); ok {
				if ident, ok := kv.Key.(*ast.Ident); ok && ident.Name == "string" {
					newAppEntry := &ast.KeyValueExpr{
						Key: &ast.BasicLit{
							Kind:  token.STRING,
							Value: fmt.Sprintf(`"%s"`, appName),
						},
						Value: &ast.CompositeLit{
							Type: &ast.SelectorExpr{
								X:   ast.NewIdent(appName),
								Sel: ast.NewIdent("App"),
							},
						},
					}
					cl.Elts = append(cl.Elts, newAppEntry)
					return false
				}
			}
		}
		return true
	})

	var buf bytes.Buffer
	if err := format.Node(&buf, fset, node); err != nil {
		return err
	}
	return os.WriteFile(path, buf.Bytes(), 0644)
}

func addModuleToAppMain(path, projectName, appName, moduleName string) error {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		return err
	}

	ast.Inspect(node, func(n ast.Node) bool {
		if gd, ok := n.(*ast.GenDecl); ok && gd.Tok == token.IMPORT {
			newImport := &ast.ImportSpec{
				Name: ast.NewIdent(moduleName),
				Path: &ast.BasicLit{
					Kind:  token.STRING,
					Value: fmt.Sprintf(`"%s/internal/%s/%s"`, projectName, appName, moduleName),
				},
			}
			gd.Specs = append(gd.Specs, newImport)
			return false
		}
		return true
	})

	ast.Inspect(node, func(n ast.Node) bool {
		if ce, ok := n.(*ast.CallExpr); ok {
			if se, ok := ce.Fun.(*ast.SelectorExpr); ok {
				if x, ok := se.X.(*ast.Ident); ok && x.Name == "core" && se.Sel.Name == "New" {
					newModuleEntry := &ast.CompositeLit{
						Type: &ast.SelectorExpr{
							X:   ast.NewIdent(moduleName),
							Sel: ast.NewIdent(strings.Title(moduleName) + "Module"),
						},
					}
					ce.Args = append(ce.Args, newModuleEntry)
					return false
				}
			}
		}
		return true
	})

	var buf bytes.Buffer
	if err := format.Node(&buf, fset, node); err != nil {
		return err
	}
	return os.WriteFile(path, buf.Bytes(), 0644)
}
