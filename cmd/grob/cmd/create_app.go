package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/yuliussmayoru/grob-cli/internal/templates"
	"github.com/yuliussmayoru/grob-cli/internal/utils"
)

func init() {
	rootCmd.AddCommand(createAppCmd)
}

var createAppCmd = &cobra.Command{
	Use:   "create-app [app-name]",
	Short: "Create a new web application inside a Grob project",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		appName := args[0]
		log.Printf("Creating new application: %s", appName)

		projectRoot, err := utils.FindProjectRoot()
		if err != nil {
			log.Fatalf("Error: %v. Make sure you are inside a Grob project.", err)
		}
		projectName := utils.GetProjectName(projectRoot)

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
		utils.CreateFileFromTmpl(appMainPath, templates.AppMainTmpl, map[string]string{
			"ProjectName": projectName,
			"AppName":     appName,
		})

		internalMainPath := filepath.Join(projectRoot, "internal", "main.go")
		if err := utils.AddAppToInternalMain(internalMainPath, projectName, appName); err != nil {
			log.Fatalf("Failed to auto-register app: %v", err)
		}

		log.Printf("Application '%s' created and registered successfully.", appName)
	},
}
