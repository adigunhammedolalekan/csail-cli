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

func appsCmd() {
	createCmd := &cobra.Command{
		Use: "create",
		Run: func(cmd *cobra.Command, args []string) {
			name, _ := cmd.Flags().GetString("name")
			createNewApp(name)
		},
		Short: "Create a new app on hostgolang.com",
		Long: "Create a new app on hostgolang.com - you can specify a -n flag to assign a name to your app, a random name will be chosen if no name is specified. `hostgo create -n sample`",
	}
	logsCmd := &cobra.Command{
		Use: "log",
		Run: func(cmd *cobra.Command, args []string) {
			getAppLogs()
		},
		Short: "Retrieve application logs",
		Long: "Run `hostgo log` to retrieve application logs",
	}
	deploymentCmd := &cobra.Command{
		Use: "deploy",
		Run: func(cmd *cobra.Command, args []string) {
			deployApp()
		},
		Short: "Deploy or update application deployment.",
		Long: "`hostgo deploy` will pack and deploy your application to hostgolang.com",
	}
	scaleCmd := &cobra.Command{
		Use: "scale",
		Run: func(cmd *cobra.Command, args []string) {
			i, _ := cmd.Flags().GetInt32("instances")
			if i < 2 {
				color.Red("invalid instance count. Instance should be at least 2")
				os.Exit(1)
			}
			scaleApp(i)
		},
		Short: "Scale application instances",
		Long: "You can run `hostgo scale -i {instances}` to scale your app horizontally.",
	}
	psCmd := &cobra.Command{
		Use: "ps",
		Run: func(cmd *cobra.Command, args []string) {
			listInstances()
		},
		Short: "Print running application instances",
	}
	rollbackCmd := &cobra.Command{
		Use: "rollback",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) > 0 {
				v := args[0]
				provider := auth.NewAuthProvider()
				account, err := provider.CurrentAuth()
				if err != nil {
					color.Red("\n\nYou have to be authenticated before you can create an app. Run `hostgo login` to authenticate your account")
					os.Exit(1)
				}
				app, err := readAppConfig()
				if err != nil {
					color.Red("failed to read app config: ", err.Error())
					os.Exit(1)
				}
				httpClient := http.NewHttpClient(account)
				s := spinner.New(spinner.CharSets[4], 200*time.Millisecond)
				s.Prefix = "working..."
				s.Start()
				op := ops.NewAppsOp(httpClient)
				r, err := op.RollbackDeployment(app.AppName, v)
				if err != nil {
					s.Stop()
					color.Red(err.Error())
					fmt.Println()
					os.Exit(1)
				}
				s.Stop()
				fmt.Println(color.WhiteString("working...done"))
				color.Green(r)
			}
		},
	}
	resourceCmd := &cobra.Command{
		Use: "resource",
	}
	addResourceCmd := &cobra.Command{
		Use: "add",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) > 0 {
				name := args[0]
				provisionResource(name)
			}
		},
	}
	removeResouceCmd := &cobra.Command{
		Use: "remove",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) > 0 {
				removeResource(args[0])
			}
		},
	}
	resourceCmd.AddCommand(addResourceCmd, removeResouceCmd)
	scaleCmd.Flags().Int32P("instances", "i", 0, "number of instances to scale to")
	createCmd.Flags().StringP("name", "n", "", "Preferred app name")
	rootCmd.AddCommand(createCmd, logsCmd, deploymentCmd, scaleCmd, psCmd, rollbackCmd, resourceCmd)
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
	s.Stop()
	fmt.Println(color.WhiteString("creating app...done"))
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
	s := spinner.New(spinner.CharSets[4], 200 * time.Millisecond)
	s.Prefix = "packing app..."
	s.Start()
	deploymentClient := http.NewDeploymentClient(cfg.AppName, account)
	if err := deploymentClient.BuildBinary(); err != nil {
		color.Red(err.Error())
		os.Exit(1)
	}
	s.Stop()
	fmt.Println(color.WhiteString("packing app...done"))
	wd, _ := os.Getwd()
	binPath := filepath.Join(wd, cfg.AppName)
	ss := spinner.New(spinner.CharSets[4], 200 * time.Millisecond)
	ss.Prefix = "creating deployment..."
	ss.Start()
	startTime := time.Now()
	r := &types.DeploymentResult{}
	err = deploymentClient.DeployApp(binPath, r)
	if err != nil {
		fmt.Println()
		color.Red(err.Error())
		fmt.Println()
		os.Exit(1)
	}
	ss.Stop()
	fmt.Println(color.WhiteString("creating deployment...done"))
	fmt.Println("====")
	message := fmt.Sprintf("%s", color.GreenString("Deployment updated! | https://%s.hostgoapp.com | %s:%s", cfg.AppName, cfg.AppName, r.Data.Version))
	fmt.Println(message)
	fmt.Println()
	endTime := time.Since(startTime).Seconds()
	color.Green("Request took: %f", endTime)
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

func scaleApp(instance int32) {
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
	s := spinner.New(spinner.CharSets[4], 200 * time.Millisecond)
	s.Prefix = "working..."
	s.Start()

	type serverResponse struct {
		Error bool
		Message string
	}
	srvResponse := &serverResponse{}
	httpClient := http.NewHttpClient(account)
	err = httpClient.Do(fmt.Sprintf("/apps/scale/%s?replicas=%d", cfg.AppName, instance), "GET", nil, srvResponse)
	s.Stop()
	fmt.Println(color.WhiteString("working...done"))
	if err != nil {
		fmt.Println()
		color.Red(err.Error())
		fmt.Println()
		os.Exit(1)
	}
	color.Green(srvResponse.Message)
}

func listInstances() {
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
	s := spinner.New(spinner.CharSets[4], 200 * time.Millisecond)
	s.Prefix = "working..."
	s.Start()

	httpClient := http.NewHttpClient(account)
	type serverResponse struct {
		Error bool
		Message string
		Data []struct{
			Id string `json:"id"`
			Name string `json:"name"`
			Status string `json:"status"`
			Started string `json:"started"`
		} `json:"data"`
	}
	srvResponse := &serverResponse{}
	err = httpClient.Do(fmt.Sprintf("/apps/ps/%s", cfg.AppName), "GET", nil, srvResponse)
	s.Stop()
	fmt.Println(color.WhiteString("working...done"))
	if err != nil {
		fmt.Println()
		color.Red(err.Error())
		fmt.Println()
		os.Exit(1)
	}
	fmt.Println("ID\t\tNAME\t\tSTATUS\t\tSTARTED")
	for _, p := range srvResponse.Data {
		fmt.Println(fmt.Sprintf("%s\t\t%s\t\t%s\t\t%s", p.Id, p.Name, p.Status, p.Started))
	}
	fmt.Println()
}

func provisionResource(name string) {
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
	loading := fmt.Sprintf("adding %s...", name)
	s := spinner.New(spinner.CharSets[4], 200 * time.Millisecond)
	s.Prefix = loading
	s.Start()

	httpClient := http.NewHttpClient(account)
	op := ops.NewAppsOp(httpClient)
	r, err := op.ProvisionResource(cfg.AppName, name)
	if err != nil {
		fmt.Println()
		color.Red(err.Error())
		fmt.Println()
		os.Exit(1)
	}
	s.Stop()
	fmt.Println(color.WhiteString(loading, "done"))
	fmt.Println(color.GreenString(r))
}

func removeResource(name string) {
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
	loading := fmt.Sprintf("removing %s...", name)
	s := spinner.New(spinner.CharSets[4], 200 * time.Millisecond)
	s.Prefix = loading
	s.Start()

	httpClient := http.NewHttpClient(account)
	op := ops.NewAppsOp(httpClient)
	r, err := op.DeleteResource(cfg.AppName, name)
	if err != nil {
		fmt.Println()
		color.Red(err.Error())
		fmt.Println()
		os.Exit(1)
	}
	s.Stop()
	fmt.Println(color.WhiteString(loading, "done"))
	fmt.Println(color.GreenString(r))
}