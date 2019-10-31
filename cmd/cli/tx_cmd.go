package main

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

func txCommand(pHostURL *string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tx",
		Short: "manages accounts and balances",
	}
	cmd.AddCommand(&cobra.Command{
		Use:   "list [address] (offset=0) (count=10)",
		Short: "returns transactions of the address from recents (default: offset=0, count=10)",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			offset := 0
			if len(args) > 1 && len(args[1]) > 0 {
				v, err := strconv.Atoi(args[1])
				if err != nil {
					fmt.Println("error :", err)
					return
				}
				offset = v
			}
			count := 10
			if len(args) > 2 && len(args[2]) > 0 {
				v, err := strconv.Atoi(args[2])
				if err != nil {
					fmt.Println("error :", err)
					return
				}
				offset = v
			}
			res, err := DoRequest((*pHostURL), "bank.transactions", []interface{}{args[0], offset, count})
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
		Use:   "deposits [address] (offset=0) (count=10)",
		Short: "returns deposits of the address from recents",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			offset := 0
			if len(args) > 1 && len(args[1]) > 0 {
				v, err := strconv.Atoi(args[1])
				if err != nil {
					fmt.Println("error :", err)
					return
				}
				offset = v
			}
			count := 10
			if len(args) > 2 && len(args[2]) > 0 {
				v, err := strconv.Atoi(args[2])
				if err != nil {
					fmt.Println("error :", err)
					return
				}
				offset = v
			}
			res, err := DoRequest((*pHostURL), "bank.transferRecvs", []interface{}{args[0], offset, count})
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
		Use:   "withdrawals [address] (offset=0) (count=10)",
		Short: "returns withdrawals of the address from recents (default: offset=0, count=10)",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			offset := 0
			if len(args) > 1 && len(args[1]) > 0 {
				v, err := strconv.Atoi(args[1])
				if err != nil {
					fmt.Println("error :", err)
					return
				}
				offset = v
			}
			count := 10
			if len(args) > 2 && len(args[2]) > 0 {
				v, err := strconv.Atoi(args[2])
				if err != nil {
					fmt.Println("error :", err)
					return
				}
				offset = v
			}
			res, err := DoRequest((*pHostURL), "bank.transferSends", []interface{}{args[0], offset, count})
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
	return cmd
}
