package cli

import (
	"github.com/spf13/cobra"
)

func ListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all todos",
		RunE:  ListCmdFunc(),
	}
}

func ListCmdFunc() func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		// Check for config file first
		_, err := loadConfig(cmd)
		if err != nil {
			return err
		}

		return nil
	}
}
