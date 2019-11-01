package main

import (
	"encoding/json"
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
		Short: "returns addresses of the name",
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
		Use:   "get [address]",
		Short: "returns account data of the address",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			res, err := DoRequest((*pHostURL), "bank.accountDetail", []interface{}{args[0]})
			if err != nil {
				fmt.Println("error :", err)
			} else {
				bs, err := json.MarshalIndent(res, "", "\t")
				if err != nil {
					fmt.Println("error :", err)
				} else {
					fmt.Println(string(bs))
				}
			}
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:   "balance [address]",
		Short: "returns account balance of the address",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			res, err := DoRequest((*pHostURL), "vault.balance", []interface{}{args[0]})
			if err != nil {
				fmt.Println("error :", err)
			} else {
				fmt.Println(res)
			}
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:   "send [from] [to] [amount] (password)",
		Short: "sends the amount of FLETA",
		Args:  cobra.MinimumNArgs(3),
		Run: func(cmd *cobra.Command, args []string) {
			var Password string
			if len(args) > 3 {
				Password = args[3]
			}
			res, err := DoRequest((*pHostURL), "bank.send", []interface{}{args[0], args[1], args[2], Password})
			if err != nil {
				fmt.Println("error :", err)
			} else {
				fmt.Println(res)
			}
		},
	})
	return cmd
}
