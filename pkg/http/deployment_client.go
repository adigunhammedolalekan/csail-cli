package http

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/saas/hostgo/pkg/types"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"time"
)

// DeploymentClient
type DeploymentClient struct {
	cmd        *CmdClient
	httpClient *http.Client
	appName    string
	account    *types.Account
}

func NewDeploymentClient(appName string, account *types.Account) *DeploymentClient {
	cmd := NewCmdClient(appName)
	httpClient := &http.Client{Timeout: 60 * time.Second}
	return &DeploymentClient{
		cmd:        cmd,
		httpClient: httpClient,
		appName:    appName,
		account:    account,
	}
}

func (s *DeploymentClient) DeployApp(binPath string, result *types.DeploymentResult) error {
	buf := &bytes.Buffer{}
	writer := multipart.NewWriter(buf)
	if err := writer.WriteField("app_name", s.appName); err != nil {
		return err
	}
	in, err := os.Open(binPath)
	if err != nil {
		return err
	}
	out, err := writer.CreateFormFile("bin", binPath)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	if err := writer.Close(); err != nil {
		return err
	}

	serverUrl := fmt.Sprintf("%s/apps/deploy", serverUrl)
	req, err := http.NewRequest("POST", serverUrl, buf)
	if err != nil {
		return err
	}
	if s.account != nil {
		req.Header.Set("X-Auth-Token", s.account.Token)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	r, err := s.httpClient.Do(req)
	if err != nil {
		return err
	}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(body, result); err != nil {
		return err
	}
	if r.StatusCode != http.StatusOK {
		return errors.New(result.Message)
	}
	return nil
}

func (s *DeploymentClient) BuildBinary() error {
	return s.cmd.ExecBuildCommand()
}

// env GOOS=linux go build -ldflags="-s -w" -o stormTest main.go
type CmdClient struct {
	appName string
}

func NewCmdClient(appName string) *CmdClient {
	return &CmdClient{appName: appName}
}

func (c *CmdClient) ExecBuildCommand() error {
	// change build env to linux, we need linux container
	if err := os.Setenv("GOOS", "linux"); err != nil {
		return err
	}
	cmd := exec.Command("go", "build", "-o", c.appName)
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	cmd.Dir = wd
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
