package cmd

import (
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/yuliussmayoru/grob-cli/internal/templates"
	"github.com/yuliussmayoru/grob-cli/internal/utils"
)

func init() {
	rootCmd.AddCommand(newCmd)
}

var newCmd = &cobra.Command{
	Use:   "new [project-name]",
	Short: "Create a new Grob project",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
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

		utils.CreateFileFromTmpl(filepath.Join(projectName, "go.mod"), templates.GoModTmpl, map[string]string{"ProjectName": projectName})
		utils.CreateFileFromTmpl(filepath.Join(projectName, ".gitignore"), templates.GitignoreTmpl, nil)
		utils.CreateFileFromTmpl(filepath.Join(projectName, "internal", "main.go"), templates.InternalMainTmpl, nil)

		log.Printf("Project '%s' created successfully.", projectName)
		log.Println("Next steps:")
		log.Printf("  cd %s", projectName)
		log.Println("  grob create-app myapp")
		log.Println("  go mod tidy  # To download dependencies")
	},
}
