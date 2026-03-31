package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// DeleteCmd returns a *cobra.Command instance for the delete command
// Uses global variables configRoot and todoPath
func DeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete [id]",
		Short: "Delete a todo item from the list",
		RunE:  DeleteCmdFunc(),
	}
	cmd.Flags().Bool("force", false, "Force deletion even if todo has children")
	return cmd
}

// DeleteCmdFunc returns a RunE function for the delete command
func DeleteCmdFunc() func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("accepts 1 arg(s), received %d", len(args))
		}

		config, err := loadConfig(cmd)
		if err != nil {
			return err
		}

		if todoPath == "" {
			cwd, _ := os.Getwd()
			todoPath = filepath.Join(cwd, config.Filename)
		}

		tl, _, _ := loadTodoList(todoPath)

		todoID := args[0]
		force, _ := cmd.Flags().GetBool("force")

		if err := tl.Delete(todoID, force); err != nil {
			return err
		}

		return tl.Save()
	}
}
