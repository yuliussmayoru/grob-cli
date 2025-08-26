package templates

// This file will hold all the boilerplate templates for code generation.

var GoModTmpl = `module {{.ProjectName}}

go 1.19

require (
	github.com/gin-gonic/gin v1.8.1
	github.com/yuliussmayoru/grob-framework v0.1.0
	go.uber.org/dig v1.15.0
)
`

var GitignoreTmpl = `
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

var InternalMainTmpl = `package main

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

var AppMainTmpl = `package {{.AppName}}

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

var ModuleTmpl = `package {{.ModuleName}}

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

var ServiceTmpl = `package {{.ModuleName}}

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

var ControllerTmpl = `package {{.ModuleName}}

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
