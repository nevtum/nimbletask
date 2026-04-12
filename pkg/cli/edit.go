package cli

import (
	"github.com/nevtum/nimbletask/pkg/todo"
	"github.com/spf13/cobra"
)

func EditCmd() *cobra.Command {
	var title string

	cmd := &cobra.Command{
		Use:   "edit [ID]",
		Short: "Edit an existing todo",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]
			path := todoPath
			if path == "" {
				path = "todo.md"
			}

			tl, _, _ := loadTodoList(path)

			updates := todo.TodoUpdate{
				Title: &title,
			}

			_, err := tl.Update(id, updates)
			if err != nil {
				return err
			}

			return tl.Save()
		},
	}

	cmd.Flags().StringVar(&title, "title", "", "New title for the todo")

	return cmd
}
