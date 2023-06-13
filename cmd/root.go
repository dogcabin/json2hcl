package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "jsonvars2hcl [flags] jsonvars",
	Short: "Converts Terraform variables in JSON format to HCL format",
	Long:  `Converts Terraform variables in JSON format to HCL format..`,
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := convert(args[0]); err != nil {
			return err
		}
		return nil
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
