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
		Long: `Delete a todo item from the list.

By default, todos with children cannot be deleted to prevent accidental
loss of nested tasks. Use the --force flag to delete a parent todo;
its children will be promoted to the root level.`,
		RunE: DeleteCmdFunc(),
	}
	cmd.Flags().Bool("force", false, "Force deletion even if todo has children (children will be promoted to root)")
	return cmd
}

// DeleteCmdFunc returns a RunE function for the delete command
func DeleteCmdFunc() func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("received %d args, need one todo ID to delete", len(args))
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

		if err := tl.Save(); err != nil {
			return err
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Deleted todo %s\n", todoID)
		return nil
	}
}
