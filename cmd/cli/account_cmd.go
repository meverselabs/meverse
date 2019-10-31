package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func accountCommand(pHostURL *string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "account",
		Short: "manages accounts and balances",
	}
	cmd.AddCommand(&cobra.Command{
		Use:   "list [name]",
		Short: "returns account addresses of the name",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			res, err := DoRequest((*pHostURL), "bank.accounts", []interface{}{args[0]})
			if err != nil {
				fmt.Println("error :", err)
			} else {
				fmt.Println(res)
			}
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:   "send [from] [to] [amount]",
		Short: "send the amount of FLETA",
		Args:  cobra.MinimumNArgs(3),
		Run: func(cmd *cobra.Command, args []string) {
			res, err := DoRequest((*pHostURL), "bank.send", []interface{}{args[0], args[1], args[2]})
			if err != nil {
				fmt.Println("error :", err)
			} else {
				fmt.Println(res)
			}
		},
	})
	return cmd
}
