package ops

import (
	"context"
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
	"github.com/docker/docker/pkg/jsonmessage"
	"io"
	"time"
)

const registryUrl = "registry.csail.app/"

type DockerService interface {
	BuildImage(buildDir, appName string) (string, error)
	PushImage(ref string) error
}

type DockerOps struct {
	client *client.Client
}

func NewDockerOps() (DockerService, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}
	return &DockerOps{client: cli}, nil
}

func (op *DockerOps) BuildImage(dir, appName string) (string, error) {
	buildCtx, err := op.createBuildContext(dir)
	if err != nil {
		return "", err
	}
	tag := op.randomMd5()[:6]
	pushUrl := fmt.Sprintf("%s%s:%s", registryUrl, appName, tag)
	r, err := op.client.ImageBuild(context.Background(),
		buildCtx, types.ImageBuildOptions{
			NoCache: false,
			Tags: []string{pushUrl},
		})
	if err != nil {
		return "", err
	}
	jm := jsonmessage.JSONMessage{}
	dec := json.NewDecoder(r.Body)
	for {
		if err := dec.Decode(&jm); err != nil {
			break
		}
		if jm.Error != nil && jm.Error.Message != "" {
			return "", errors.New(jm.Error.Message)
		}
	}
	return pushUrl, nil
}

func (op *DockerOps) PushImage(ref string) error {
	r, err := op.client.ImagePush(context.Background(),
		ref, types.ImagePushOptions{RegistryAuth: op.registryAuthAsBase64()})
	if err != nil {
		return err
	}
	dec := json.NewDecoder(r)
	jm := jsonmessage.JSONMessage{}
	for {
		if err := dec.Decode(&jm); err != nil {
			break
		}
		if jm.Error != nil && jm.Error.Message != "" {
			return errors.New(jm.Error.Message)
		}
	}
	return nil
}

func (op *DockerOps) createBuildContext(dir string) (io.Reader, error) {
	return archive.Tar(dir, archive.Uncompressed)
}

func (op *DockerOps) randomMd5() string {
	m := md5.New()
	m.Write([]byte(time.Now().String()))
	return fmt.Sprintf("%x", m.Sum(nil))
}

func (op *DockerOps) registryAuthAsBase64() string {
	authConfig := types.AuthConfig{
		Username: "lekan",
		Password: "manman",
	}
	encoded, err := json.Marshal(authConfig)
	if err != nil {
		return ""
	}
	return base64.StdEncoding.EncodeToString(encoded)
}