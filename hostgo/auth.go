package main

import (
	"fmt"
	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
	"github.com/saas/hostgo/pkg/auth"
	"github.com/saas/hostgo/pkg/http"
	"github.com/saas/hostgo/pkg/ops"
	"github.com/spf13/cobra"
	"time"
)

func authCmd() {
	authCommand := &cobra.Command{
		Use: "login",
		Run: func(cmd *cobra.Command, args []string) {
			performAuthentication()
		},
	}
	rootCmd.AddCommand(authCommand)
}

func performAuthentication() {
	promptEmail := promptui.Prompt{
		Label:    "Email",
		Validate: nil,
	}
	email, _ := promptEmail.Run()
	promptPassword := promptui.Prompt{
		Label:    "Password",
		Validate: nil,
	}
	password, _ := promptPassword.Run()

	s := spinner.New(spinner.CharSets[4], 200*time.Millisecond)
	s.Prefix = "Authenticating..."
	s.Start()
	httpClient := http.NewHttpClient(nil)
	provider := auth.NewAuthProvider()
	op := ops.NewAuthenticateAccountOp(httpClient, provider)
	account, err := op.AuthenticateAccount(email, fmt.Sprintf("%s", password))
	if err != nil {
		color.Red(fmt.Sprintf("\n%s", err.Error()))
		return
	}
	s.Stop()
	fmt.Printf("Success. Authenticated as: %s\n\n", color.GreenString(account.Email))
}
