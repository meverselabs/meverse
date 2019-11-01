package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func keyCommand(pHostURL *string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "key",
		Short: "manages keys of the connected wallet",
	}
	cmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "returns names of keys",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			res, err := DoRequest((*pHostURL), "bank.keyNames", []interface{}{})
			if err != nil {
				fmt.Println("error :", err)
			} else {
				fmt.Println(res)
			}
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:   "create [name] (password)",
		Short: "creates new key with name",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if strings.Contains(args[0], " ") {
				fmt.Println("error : name cannot include a white space")
				return
			}
			var Password string
			if len(args) > 1 {
				Password = args[1]
			}
			_, err := DoRequest((*pHostURL), "bank.createKey", []interface{}{args[0], Password})
			if err != nil {
				fmt.Println("error :", err)
			} else {
				fmt.Println("the key is created")
			}
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:   "import [name] [keyhex] (password)",
		Short: "imports the key with name",
		Args:  cobra.MinimumNArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			if strings.Contains(args[0], " ") {
				fmt.Println("error : name cannot include a white space")
				return
			}
			var Password string
			if len(args) > 2 {
				Password = args[2]
			}
			_, err := DoRequest((*pHostURL), "bank.importKey", []interface{}{args[0], args[1], Password})
			if err != nil {
				fmt.Println("error :", err)
			} else {
				fmt.Println("the key is imported")
			}
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:   "password [name] [oldpassword] [newpassword]",
		Short: "changes the password of a key with the name",
		Args:  cobra.MinimumNArgs(3),
		Run: func(cmd *cobra.Command, args []string) {
			if strings.Contains(args[0], " ") {
				fmt.Println("error : name cannot include a white space")
				return
			}
			_, err := DoRequest((*pHostURL), "bank.changePassword", []interface{}{args[0], args[1], args[2]})
			if err != nil {
				fmt.Println("error :", err)
			} else {
				fmt.Println("the pasword is changed")
			}
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:   "delete [name] (password)",
		Short: "deletes key that has the name",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if strings.Contains(args[0], " ") {
				fmt.Println("error : name cannot include a white space")
				return
			}
			var Password string
			if len(args) > 1 {
				Password = args[1]
			}
			_, err := DoRequest((*pHostURL), "bank.deleteKey", []interface{}{args[0], Password})
			if err != nil {
				fmt.Println("error :", err)
			} else {
				fmt.Println("the key is deleted")
			}
		},
	})
	return cmd
}
