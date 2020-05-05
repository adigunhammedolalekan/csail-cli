package main

import (
	"fmt"
	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	"github.com/saas/hostgo/pkg/auth"
	"github.com/saas/hostgo/pkg/http"
	"github.com/saas/hostgo/pkg/types"
	"github.com/spf13/cobra"
	"os"
	"strings"
	"time"
)

func envCmd() {
	eCmd := &cobra.Command{
		Use: "env",
		Run: func(cmd *cobra.Command, args []string) {
			provider := auth.NewAuthProvider()
			account, err := provider.CurrentAuth()
			if err != nil ||  account.Token == "" {
				color.Red("\n\nYou have to be authenticated before you can access an app. Run `hostgo login` to authenticate your account")
				os.Exit(1)
			}
			cfg, err := readAppConfig()
			if err != nil {
				color.Red("failed to read app config: ", err.Error())
				os.Exit(1)
			}
			type serverResponse struct {
				Error bool `json:"error"`
				Message string `json:"message"`
				Data []struct{
					EnvKey string `json:"env_key"`
					EnvValue string `json:"env_value"`
				}
			}
			s := spinner.New(spinner.CharSets[4], 200 * time.Millisecond)
			s.Prefix = "working..."
			s.Start()
			httpClient := http.NewHttpClient(account)
			srvResponse := &serverResponse{}
			err = httpClient.Do("/apps/configs/" + cfg.AppName, "GET", nil, srvResponse)
			if err != nil {
				fmt.Println()
				color.Red(err.Error())
				fmt.Println()
				os.Exit(1)
			}
			s.Stop()
			fmt.Printf("Total envs: %d\n", len(srvResponse.Data))
			for _, value := range srvResponse.Data {
				fmt.Println(color.WhiteString("%s=%s", value.EnvKey, value.EnvValue))
			}
		},
		Short: "Print/List application environment variables",
	}
	setEnvCmd := &cobra.Command{
		Use: "set",
		Run: func(cmd *cobra.Command, args []string) {
			provider := auth.NewAuthProvider()
			account, err := provider.CurrentAuth()
			if err != nil ||  account.Token == "" {
				color.Red("\n\nYou have to be authenticated before you can access an app. Run `hostgo login` to authenticate your account")
				os.Exit(1)
			}
			cfg, err := readAppConfig()
			if err != nil {
				color.Red("failed to read app config: ", err.Error())
				os.Exit(1)
			}
			params := make([]types.Env, 0)
			if len(args) > 0 {
				for _, arg := range args {
					envs := strings.Split(arg, "=")
					if len(envs) != 2 {
						color.Red(fmt.Sprintf("invalid env format: %s, expected format is KEY=VALUE. run 'hostgo set env KEY1=VALUE1 KEY2=VALUE2 ...'", arg))
						os.Exit(1)
					}else {
						k, v := envs[0], envs[1]
						params = append(params, types.Env{Key: k, Value: v})
					}
				}

				s := spinner.New(spinner.CharSets[4], 200 * time.Millisecond)
				s.Prefix = "setting env..."
				s.Start()

				httpClient := http.NewHttpClient(account)
				type serverResponse struct {
					Error bool `json:"error"`
					Message string `json:"message"`
				}
				srvResponse := &serverResponse{}
				err = httpClient.Do("/apps/configs/" + cfg.AppName, "POST", params, srvResponse)
				if err != nil {
					fmt.Println()
					color.Red(err.Error())
					fmt.Println()
					os.Exit(1)
				}
				s.Stop()
				fmt.Println(color.GreenString("%s\n", srvResponse.Message))
			}
		},
		Short: "Set a new environment variable or update an existing one",
		Long: "`hostgo env set KEY1=VALUE1 KEY2=VALUE2 ...`",
	}
	unsetEnvCmd := &cobra.Command{
		Use: "unset",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) > 0 {
				provider := auth.NewAuthProvider()
				account, err := provider.CurrentAuth()
				if err != nil ||  account.Token == "" {
					color.Red("\n\nYou have to be authenticated before you can access an app. Run `hostgo login` to authenticate your account")
					os.Exit(1)
				}
				cfg, err := readAppConfig()
				if err != nil {
					color.Red("failed to read app config: ", err.Error())
					os.Exit(1)
				}
				type serverResponse struct {
					Error bool `json:"error"`
					Message string `json:"message"`
				}
				s := spinner.New(spinner.CharSets[4], 200 * time.Millisecond)
				s.Prefix = "working..."
				s.Start()
				httpClient := http.NewHttpClient(account)
				srvResponse := &serverResponse{}
				err = httpClient.Do("/apps/configs/unset/" + cfg.AppName, "DELETE", args, srvResponse)
				if err != nil {
					fmt.Println()
					color.Red(err.Error())
					fmt.Println()
					os.Exit(1)
				}
				s.Stop()
				color.Green("\noperation successful\n")
			}
		},
		Short: "Unset or delete an environment variable",
		Long: "`hostgo env unset KEY1 KEY2 KEY3...",
	}
	eCmd.AddCommand(setEnvCmd, unsetEnvCmd)
	rootCmd.AddCommand(eCmd)
}
