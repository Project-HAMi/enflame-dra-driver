/*
 * Copyright 2023 The Kubernetes Authors.
 * Copyright The Project-HAMi Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/urfave/cli/v2"

	coreclientset "k8s.io/client-go/kubernetes"
	"k8s.io/dynamic-resource-allocation/kubeletplugin"
	"k8s.io/klog/v2"

	"github.com/Project-HAMi/enflame-dra-driver/internal/profiles"
	"github.com/Project-HAMi/enflame-dra-driver/internal/profiles/gcu"
	"github.com/Project-HAMi/enflame-dra-driver/pkg/flags"
	"github.com/Project-HAMi/enflame-dra-driver/pkg/metrics"
)

const (
	DriverPluginCheckpointFile = "checkpoint.json"
)

type Flags struct {
	kubeClientConfig flags.KubeClientConfig
	loggingConfig    *flags.LoggingConfig

	nodeName                      string
	cdiRoot                       string
	hostRoot                      string
	numDevices                    int
	kubeletRegistrarDirectoryPath string
	kubeletPluginsDirectoryPath   string
	healthcheckPort               int
	metricsPort                   int
	profile                       string
	driverName                    string
	podUID                        string
	deviceHealth                  bool
}

type Config struct {
	flags         *Flags
	coreclient    coreclientset.Interface
	cancelMainCtx func(error)

	profile profiles.Profile
}

func (c Config) DriverPluginPath() string {
	return filepath.Join(c.flags.kubeletPluginsDirectoryPath, c.flags.driverName)
}

func main() {
	if err := newApp().Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func newApp() *cli.App {
	flags := &Flags{
		loggingConfig: flags.NewLoggingConfig(),
	}
	cliFlags := []cli.Flag{
		&cli.StringFlag{
			Name:        "node-name",
			Usage:       "The name of the node to be worked on.",
			Required:    true,
			Destination: &flags.nodeName,
			EnvVars:     []string{"NODE_NAME"},
		},
		&cli.StringFlag{
			Name:        "cdi-root",
			Usage:       "Absolute path to the directory where CDI files will be generated.",
			Value:       "/var/run/cdi",
			Destination: &flags.cdiRoot,
			EnvVars:     []string{"CDI_ROOT"},
		},
		&cli.StringFlag{
			Name:        "host-root",
			Usage:       "Absolute path to the host filesystem root used when discovering GCU driver dependencies for CDI specs.",
			Value:       "/",
			Destination: &flags.hostRoot,
			EnvVars:     []string{"HOST_ROOT"},
		},
		&cli.IntFlag{
			Name:        "num-devices",
			Usage:       "Deprecated. Device discovery uses libefml only; this flag is ignored.",
			Value:       8,
			Destination: &flags.numDevices,
			EnvVars:     []string{"NUM_DEVICES"},
			Hidden:      true,
		},
		&cli.StringFlag{
			Name:        "kubelet-registrar-directory-path",
			Usage:       "Absolute path to the directory where kubelet stores plugin registrations.",
			Value:       kubeletplugin.KubeletRegistryDir,
			Destination: &flags.kubeletRegistrarDirectoryPath,
			EnvVars:     []string{"KUBELET_REGISTRAR_DIRECTORY_PATH"},
		},
		&cli.StringFlag{
			Name:        "kubelet-plugins-directory-path",
			Usage:       "Absolute path to the directory where kubelet stores plugin data.",
			Value:       kubeletplugin.KubeletPluginsDir,
			Destination: &flags.kubeletPluginsDirectoryPath,
			EnvVars:     []string{"KUBELET_PLUGINS_DIRECTORY_PATH"},
		},
		&cli.IntFlag{
			Name:        "healthcheck-port",
			Usage:       "Port to start a gRPC healthcheck service. When positive, a literal port number. When zero, a random port is allocated. When negative, the healthcheck service is disabled.",
			Value:       -1,
			Destination: &flags.healthcheckPort,
			EnvVars:     []string{"HEALTHCHECK_PORT"},
		},
		&cli.IntFlag{
			Name:        "metrics-port",
			Usage:       "Port to expose Prometheus metrics at /metrics. When positive, a literal port number. When zero, a random port is allocated. When negative, metrics are disabled.",
			Value:       -1,
			Destination: &flags.metricsPort,
			EnvVars:     []string{"METRICS_PORT"},
		},
		&cli.StringFlag{
			Name:        "device-profile",
			Usage:       "Name of the device profile. Only \"gcu\" is supported.",
			Value:       gcu.ProfileName,
			Destination: &flags.profile,
			EnvVars:     []string{"DEVICE_PROFILE"},
		},
		&cli.StringFlag{
			Name:        "driver-name",
			Usage:       "Name of the DRA driver.",
			Value:       gcu.DefaultDriverName,
			Destination: &flags.driverName,
			EnvVars:     []string{"DRIVER_NAME"},
		},
		&cli.StringFlag{
			Name:        "pod-uid",
			Usage:       "UID of the pod (used for seamless upgrades to create unique socket names).",
			Destination: &flags.podUID,
			EnvVars:     []string{"POD_UID"},
		},
		&cli.BoolFlag{
			Name:        "device-health",
			Usage:       "Enable device health reporting (KEP-4680). When false, the driver advertises no health service to the kubelet and WatchHealthStatus returns ErrHealthNotSupported.",
			Value:       true,
			Destination: &flags.deviceHealth,
			EnvVars:     []string{"DEVICE_HEALTH"},
		},
	}
	cliFlags = append(cliFlags, flags.kubeClientConfig.Flags()...)
	cliFlags = append(cliFlags, flags.loggingConfig.Flags()...)

	app := &cli.App{
		Name:            "enflame-dra-kubeletplugin",
		Usage:           "enflame-dra-kubeletplugin implements the Enflame GCU DRA driver plugin.",
		ArgsUsage:       " ",
		HideHelpCommand: true,
		Flags:           cliFlags,
		Before: func(c *cli.Context) error {
			if c.Args().Len() > 0 {
				return fmt.Errorf("arguments not supported: %v", c.Args().Slice())
			}
			return flags.loggingConfig.Apply()
		},
		Action: func(c *cli.Context) error {
			ctx := c.Context
			clientSets, err := flags.kubeClientConfig.NewClientSets()
			if err != nil {
				return fmt.Errorf("create client: %w", err)
			}

			if flags.profile != gcu.ProfileName {
				return fmt.Errorf("invalid device profile %q, only %q is supported", flags.profile, gcu.ProfileName)
			}
			if flags.driverName == "" {
				flags.driverName = gcu.DefaultDriverName
			}

			config := &Config{
				flags:      flags,
				coreclient: clientSets.Core,
				profile:    gcu.NewProfile(flags.nodeName, flags.hostRoot, flags.numDevices),
			}

			return RunPlugin(ctx, config)
		},
	}

	return app
}

func RunPlugin(ctx context.Context, config *Config) error {
	logger := klog.FromContext(ctx)

	err := os.MkdirAll(config.DriverPluginPath(), 0750)
	if err != nil {
		return err
	}

	info, err := os.Stat(config.flags.cdiRoot)
	switch {
	case err != nil && os.IsNotExist(err):
		err := os.MkdirAll(config.flags.cdiRoot, 0750)
		if err != nil {
			return err
		}
	case err != nil:
		return err
	case !info.IsDir():
		return fmt.Errorf("path for cdi file generation is not a directory: '%v'", err)
	}

	ctx, stop := signal.NotifyContext(ctx, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer stop()
	ctx, cancel := context.WithCancelCause(ctx)
	config.cancelMainCtx = cancel

	metricsServer, err := metrics.StartServer(ctx, config.flags.metricsPort)
	if err != nil {
		return fmt.Errorf("start metrics server: %w", err)
	}
	defer func(ctx context.Context) {
		shutdownCtx, shutdownCancel := context.WithTimeout(ctx, 5*time.Second)
		defer shutdownCancel()
		if err := metricsServer.Stop(shutdownCtx); err != nil {
			logger.Error(err, "failed to stop metrics server")
		}
	}(context.WithoutCancel(ctx))

	driver, err := NewDriver(ctx, config)
	if err != nil {
		return err
	}

	<-ctx.Done()
	// restore default signal behavior as soon as possible in case graceful
	// shutdown gets stuck.
	stop()
	if err := context.Cause(ctx); err != nil && !errors.Is(err, context.Canceled) {
		// A canceled context is the normal case here when the process receives
		// a signal. Only log the error for more interesting cases.
		logger.Error(err, "error from context")
	}

	err = driver.Shutdown(logger)
	if err != nil {
		logger.Error(err, "Unable to cleanly shutdown driver")
	}

	return nil
}
