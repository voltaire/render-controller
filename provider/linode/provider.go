package linode

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"path/filepath"

	"github.com/docker/docker/api"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/tlsconfig"
	"github.com/docker/machine/libmachine"
	"github.com/linode/docker-machine-driver-linode/pkg/drivers/linode"
	"github.com/voltaire/render-controller/provider"
)

type Provider struct {
	docker client.APIClient
	Config *Config

	libMachineCertsDir  string
	libMachineStorePath string
}

type Config struct {
	LinodeLabel string `required:"true" split_words:"true"`
	LinodeToken string `required:"true" split_words:"true"`

	LinodeRegion string `default:"us-east" split_words:"true"`

	VolumeSize int `default:"40" split_words:"true"`

	LogAWSRegion          string `default:"us-west-2" envconfig:"LOG_AWS_REGION"`
	LogAWSAccessKeyId     string `required:"true" envconfig:"LOG_AWS_ACCESS_KEY_ID"`
	LogAWSSecretAccessKey string `required:"true" envconfig:"LOG_AWS_SECRET_ACCESS_KEY"`
}

func (svc *Provider) getOrInstallLinodeVolumePlugin(ctx context.Context) error {
	currentPlugins, err := svc.docker.PluginList(ctx, filters.NewArgs(filters.KeyValuePair{
		Key:   "enabled",
		Value: "true",
	}, filters.KeyValuePair{
		Key:   "capability",
		Value: "volumedriver",
	}))
	if err != nil {
		return err
	}
	for _, v := range currentPlugins {
		if v.Name == "linode" {
			return nil
		}
	}

	progress, err := svc.docker.PluginInstall(ctx, "linode", types.PluginInstallOptions{AcceptAllPermissions: true, Args: []string{
		"linode-token=" + svc.Config.LinodeToken,
		"linode-label=" + svc.Config.LinodeLabel,
	},
		RemoteRef: "linode/docker-volume-linode:latest",
	})
	if err != nil {
		return err
	}

	defer progress.Close()
	_, err = io.Copy(ioutil.Discard, progress)
	if err != nil {
		return err
	}
	return nil
}

func New(cfg *Config) (*Provider, error) {
	svc := &Provider{
		Config: cfg,
	}

	var err error
	svc.docker, err = client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	err = svc.getOrInstallLinodeVolumePlugin(ctx)
	if err != nil {
		return nil, err
	}

	tmpDir, err := ioutil.TempDir("", "")
	if err != nil {
		return nil, err
	}

	svc.libMachineStorePath = tmpDir
	svc.libMachineCertsDir = filepath.Join(tmpDir, "certs")

	return svc, nil
}

func (svc *Provider) GetRendererInstance(ctx context.Context) (provider.RendererInstance, error) {
	machine := libmachine.NewClient(svc.libMachineStorePath, svc.libMachineCertsDir)
	defer machine.Close()

	machineName, err := newLinodeMachineName()
	if err != nil {
		return nil, err
	}

	driver := linode.NewDriver(machineName, svc.libMachineStorePath)
	driver.APIToken = svc.Config.LinodeToken
	driver.CreatePrivateIP = true
	driver.Region = svc.Config.LinodeRegion
	driver.InstanceLabel = machineName
	driver.Tags = "Service=renderer,Function=renderer"

	data, err := json.Marshal(driver)
	if err != nil {
		return nil, err
	}

	h, err := machine.NewHost("linode", data)
	if err != nil {
		return nil, err
	}
	h.HostOptions.EngineOptions.Env = []string{
		fmt.Sprintf("AWS_ACCESS_KEY_ID=%s", svc.Config.LogAWSAccessKeyId),
		fmt.Sprintf("AWS_SECRET_ACCESS_KEY=%s", svc.Config.LogAWSSecretAccessKey),
	}
	h.HostOptions.EngineOptions.ArbitraryFlags = []string{
		"--log-driver=awslogs",
		"--log-opt=awslogs-create-group=true",
		fmt.Sprintf("--log-opt=awslogs-region=%s", svc.Config.LogAWSRegion),
		"--log-opt=awslogs-group=renderer/linode",
	}

	err = machine.Create(h)
	if err != nil {
		return nil, err
	}

	dockerUrl, err := driver.GetURL()
	if err != nil {
		return nil, err
	}

	dockerClient, err := newDockerClient(dockerUrl, filepath.Join(svc.libMachineCertsDir, machineName))
	if err != nil {
		return nil, err
	}

	return &RendererInstance{
		host:         h,
		DockerClient: dockerClient,
	}, nil
}

func newDockerClient(host string, dockerCertPath string) (*client.Client, error) {
	options := tlsconfig.Options{
		CAFile:             filepath.Join(dockerCertPath, "ca.pem"),
		CertFile:           filepath.Join(dockerCertPath, "cert.pem"),
		KeyFile:            filepath.Join(dockerCertPath, "key.pem"),
		InsecureSkipVerify: false,
	}
	tlsc, err := tlsconfig.Client(options)
	if err != nil {
		return nil, err
	}

	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsc,
		},
		CheckRedirect: client.CheckRedirect,
	}

	return client.NewClient(host, api.DefaultVersion, httpClient, nil)
}

func newLinodeMachineName() (string, error) {
	bs := make([]byte, 6)
	_, err := rand.Read(bs)
	if err != nil {
		return "", err
	}
	return "renderer-" + hex.EncodeToString(bs), nil
}
