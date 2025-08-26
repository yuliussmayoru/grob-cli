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
	rootCmd.AddCommand(createModuleCmd)
}

var createModuleCmd = &cobra.Command{
	Use:   "create-module [app-name] [module-name]",
	Short: "Create a new module within a web application",
	Args:  cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		appName := args[0]
		moduleName := args[1]
		log.Printf("Creating new module '%s' in app '%s'", moduleName, appName)

		projectRoot, err := utils.FindProjectRoot()
		if err != nil {
			log.Fatalf("Error: %v. Make sure you are inside a Grob project.", err)
		}
		projectName := utils.GetProjectName(projectRoot)

		moduleDir := filepath.Join(projectRoot, "internal", appName, moduleName)
		if err := os.Mkdir(moduleDir, 0755); err != nil {
			log.Fatalf("Failed to create module directory: %v", err)
		}

		data := map[string]string{
			"ProjectName": projectName,
			"AppName":     appName,
			"ModuleName":  moduleName,
		}
		utils.CreateFileFromTmpl(filepath.Join(moduleDir, fmt.Sprintf("%s.module.go", moduleName)), templates.ModuleTmpl, data)
		utils.CreateFileFromTmpl(filepath.Join(moduleDir, fmt.Sprintf("%s.service.go", moduleName)), templates.ServiceTmpl, data)
		utils.CreateFileFromTmpl(filepath.Join(moduleDir, fmt.Sprintf("%s.controller.go", moduleName)), templates.ControllerTmpl, data)

		appMainPath := filepath.Join(projectRoot, "internal", appName, fmt.Sprintf("%s_main.go", appName))
		if err := utils.AddModuleToAppMain(appMainPath, projectName, appName, moduleName); err != nil {
			log.Fatalf("Failed to auto-register module: %v", err)
		}

		log.Printf("Module '%s' created and registered successfully in app '%s'.", moduleName, appName)
	},
}
