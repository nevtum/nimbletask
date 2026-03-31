package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// MoveCmd returns a *cobra.Command instance for the move command
func MoveCmd() *cobra.Command {
	var parentID string
	var position int

	cmd := &cobra.Command{
		Use:   "move [id]",
		Short: "Move a todo to a new parent",
		Long:  `Move a todo item under a different parent or to root level.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			todoID := args[0]

			// Check for config file first
			config, err := loadConfig(cmd)
			if err != nil {
				return err
			}

			// Determine todoPath if not set via flag
			if todoPath == "" {
				cwd, _ := os.Getwd()
				todoPath = filepath.Join(cwd, config.Filename)
			}

			// Load todo list from file
			tl, _, err := loadTodoList(todoPath)
			if err != nil {
				return err
			}

			// Move the todo
			if err := tl.Move(todoID, parentID, position); err != nil {
				return err
			}

			// Save the updated list
			if err := tl.Save(); err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Todo moved successfully\n")
			return nil
		},
	}

	cmd.Flags().StringVar(&parentID, "parent", "", "Parent todo ID (empty for root level)")
	cmd.Flags().IntVar(&position, "position", -1, "Position under parent (-1 for append)")

	return cmd
}
