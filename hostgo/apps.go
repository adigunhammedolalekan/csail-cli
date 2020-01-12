package main

import (
	"fmt"
	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	"github.com/saas/hostgo/pkg/auth"
	"github.com/saas/hostgo/pkg/http"
	"github.com/saas/hostgo/pkg/ops"
	"github.com/saas/hostgo/pkg/types"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

func appsCmd() *cobra.Command {
	appCommand := &cobra.Command{
		Use: "apps",
	}
	createCmd := &cobra.Command{
		Use: "create",
		Run: func(cmd *cobra.Command, args []string) {
			name, _ := cmd.Flags().GetString("name")
			createNewApp(name)
		},
	}
	logsCmd := &cobra.Command{
		Use: "logs",
		Run: func(cmd *cobra.Command, args []string) {
			getAppLogs()
		},
	}
	deploymentCmd := &cobra.Command{
		Use: "deploy",
		Run: func(cmd *cobra.Command, args []string) {
			deployApp()
		},
	}
	createCmd.Flags().StringP("name", "n", "", "Preferred app name")
	appCommand.AddCommand(createCmd)
	appCommand.AddCommand(deploymentCmd)
	appCommand.AddCommand(logsCmd)
	return appCommand
}

func createNewApp(name string) {
	provider := auth.NewAuthProvider()
	account, err := provider.CurrentAuth()
	if err != nil {
		color.Red("\n\nYou have to be authenticated before you can create an app. Run `hostgo login` to authenticate your account")
		os.Exit(1)
	}
	httpClient := http.NewHttpClient(account)
	s := spinner.New(spinner.CharSets[4], 200*time.Millisecond)
	s.Prefix = "creating app..."
	s.Start()

	op := ops.NewAppsOp(httpClient)
	app, err := op.CreateNewApp(name)
	if err != nil {
		color.Red("\n\n%s", err.Error())
		os.Exit(1)
	}
	if err := createAppConfigFile(app.AppName); err != nil {
	}

	fmt.Println("\n===")
	fmt.Printf("created app %s\n", color.GreenString(app.AppName))
	fmt.Printf("access url: %s\n\n", color.GreenString(app.AccessUrl))
}

func getAppLogs() {
	cfg, err := readAppConfig()
	if err != nil {
		color.Red("failed to read app config: ", err.Error())
		os.Exit(1)
	}
	provider := auth.NewAuthProvider()
	account, err := provider.CurrentAuth()
	if err != nil {
		color.Red("\n\nYou have to be authenticated before you can access an app. Run `hostgo login` to authenticate your account")
		os.Exit(1)
	}
	httpClient := http.NewHttpClient(account)
	op := ops.NewAppsOp(httpClient)
	r, err := op.ReadLogs(cfg.AppName)
	if err != nil {
		color.Red(err.Error())
		fmt.Println()
		os.Exit(1)
	}
	fmt.Print("====\n\n")
	fmt.Print(r)
	fmt.Println()
}

func deployApp() {
	cfg, err := readAppConfig()
	if err != nil {
		color.Red("failed to read app config: ", err.Error())
		os.Exit(1)
	}
	provider := auth.NewAuthProvider()
	account, err := provider.CurrentAuth()
	if err != nil {
		color.Red("\n\nYou have to be authenticated before you can access an app. Run `hostgo login` to authenticate your account")
		os.Exit(1)
	}
	deploymentClient := http.NewDeploymentClient(cfg.AppName, account)
	fmt.Println("packing app...")
	if err := deploymentClient.BuildBinary(); err != nil {
		color.Red(err.Error())
		os.Exit(1)
	}
	wd, _ := os.Getwd()
	binPath := filepath.Join(wd, cfg.AppName)
	ss := spinner.New(spinner.CharSets[4], 200 * time.Microsecond)
	ss.Prefix = "updating deployment..."
	ss.Start()
	r := &types.DeploymentResult{}
	err = deploymentClient.DeployApp(binPath, r)
	if err != nil {
		fmt.Println()
		color.Red(err.Error())
		fmt.Println()
		os.Exit(1)
	}
	ss.Stop()
	message := fmt.Sprintf("Deployment updated! %s", color.GreenString("https://%s.hostgoapp.com", cfg.AppName))
	fmt.Println("===")
	fmt.Print(message, "\n\n")
}

func createAppConfigFile(appName string) error {
	c := &types.Config{
		AppName: appName,
	}
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	filename := filepath.Join(wd, "hostgo.yml")
	d, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filename, d, os.ModePerm)
}

func readAppConfig() (*types.Config, error) {
	c := &types.Config{}
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	filename := filepath.Join(wd, "hostgo.yml")
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	if err := yaml.Unmarshal(data, c); err != nil {
		return nil, err
	}
	return c, nil
}
