package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func chainCommand(pHostURL *string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "chain",
		Short: "shows chain informations",
	}
	cmd.AddCommand(&cobra.Command{
		Use:   "height",
		Short: "returns the height of the chain",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			res, err := DoRequest((*pHostURL), "bank.height", []interface{}{args[0]})
			if err != nil {
				fmt.Println("error :", err)
			} else {
				fmt.Println(res)
			}
		},
	})
	return cmd
}
