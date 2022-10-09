/*
   Copyright 2022 The minio-operator Authors.
   Licensed under the Apache License, PROJECT_VERSION 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at
       http://www.apache.org/licenses/LICENSE-2.0
   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/
package main

import (
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/3Xpl0it3r/minio-operator/pkg/crd"

	"github.com/3Xpl0it3r/minio-operator/cmd/miniooperator/options"
	"github.com/3Xpl0it3r/minio-operator/pkg/apis/install"
	crclientset "github.com/3Xpl0it3r/minio-operator/pkg/client/clientset/versioned"
	"github.com/3Xpl0it3r/minio-operator/pkg/client/clientset/versioned/scheme"
	crinformers "github.com/3Xpl0it3r/minio-operator/pkg/client/informers/externalversions"
	"github.com/3Xpl0it3r/minio-operator/pkg/controller"
	"github.com/3Xpl0it3r/minio-operator/pkg/controller/minio"
	"github.com/spf13/cobra"
	apicorev1 "k8s.io/api/core/v1"
	extensionsclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/component-base/term"
	"k8s.io/klog/v2"
)

func NewStartCommand(stopCh <-chan struct{}) *cobra.Command {
	opts := options.NewOptions()
	cmd := &cobra.Command{
		Short: "Launch minio-operator",
		Long:  "Launch minio-operator",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := opts.Validate(); err != nil {
				return fmt.Errorf("Options validate failed, %v. ", err)
			}
			if err := opts.Complete(); err != nil {
				return fmt.Errorf("Options Complete failed %v. ", err)
			}
			if err := runCommand(opts, stopCh); err != nil {
				return fmt.Errorf("Run %s failed. Err: %v", os.Args[0], err)
			}
			return nil
		},
	}
	fs := cmd.Flags()
	nfs := opts.NamedFlagSets()
	for _, f := range nfs.FlagSets {
		fs.AddFlagSet(f)
	}
	local := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	klog.InitFlags(local)
	nfs.FlagSet("logging").AddGoFlagSet(local)

	usageFmt := "Usage:\n  %s\n"
	cols, _, _ := term.TerminalSize(cmd.OutOrStdout())
	cmd.SetUsageFunc(func(cmd *cobra.Command) error {
		fmt.Fprintf(cmd.OutOrStderr(), usageFmt, cmd.UseLine())
		cliflag.PrintSections(cmd.OutOrStderr(), nfs, cols)
		return nil
	})
	cmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		fmt.Fprintf(cmd.OutOrStdout(), "%s\n\n"+usageFmt, cmd.Long, cmd.UseLine())
		cliflag.PrintSections(cmd.OutOrStdout(), nfs, cols)
	})
	return cmd
}

func runCommand(o *options.Options, signalCh <-chan struct{}) error {
	install.Install(scheme.Scheme)

	var err error
	var stopCh = make(chan struct{})
	restConfig, err := buildKubeConfig("", "")
	if err != nil {
		return err
	}
	extClientSet, err := extensionsclientset.NewForConfig(restConfig)
	if err != nil {
		return err
	}
	if err := crd.InstallCustomResourceDefineToApiServer(extClientSet); err != nil {
		return err
	}
	defer crd.UnInstallCustomResourceDefineToApiServer(extClientSet)

	kubeClientSet, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return err
	}
	crClientSet, err := crclientset.NewForConfig(restConfig)
	if err != nil {
		return err
	}
	crInformers := buildCustomResourceInformerFactory(crClientSet)
	kubeInformers := buildKubeStandardResourceInformerFactory(kubeClientSet)

	minioController := minio.NewController(kubeClientSet, kubeInformers, crClientSet, crInformers, nil)

	crInformers.Start(stopCh)
	kubeInformers.Start(stopCh)

	if err := runController1(stopCh, minioController); err != nil {
		return err
	}

	select {
	case <-signalCh:
		klog.Infof("exited")
		close(stopCh)
	case <-stopCh:
		minioController.Stop()
	}
	return nil
}

func runController1(stopCh <- chan struct{}, controller controller.Controller) error {
	if err := controller.Start(1, stopCh); err != nil {
		return err
	}
	return nil
}


func serve(srv *http.Server, listener net.Listener) func() error {
	return func() error {
		if err := srv.Serve(listener); err != http.ErrServerClosed {
			return err
		}
		return nil
	}
}

func serveTLS(srv *http.Server, listener net.Listener) func() error {
	return func() error {
		if err := srv.ServeTLS(listener, "", ""); err != http.ErrServerClosed {
			return err
		}
		return nil
	}
}

// buildKubeConfig build rest.Config from the following ways
// 1: path of kube_config 2: KUBECONFIG environment 3. ~/.kube/config, as kubeconfig may not in /.kube/
func buildKubeConfig(masterUrl, kubeConfig string) (*rest.Config, error) {
	cfgLoadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	cfgLoadingRules.DefaultClientConfig = &clientcmd.DefaultClientConfig
	cfgLoadingRules.ExplicitPath = kubeConfig
	clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(cfgLoadingRules, &clientcmd.ConfigOverrides{})
	config, err := clientConfig.ClientConfig()
	if err != nil {
		return nil, err
	}
	if err = rest.SetKubernetesDefaults(config); err != nil {
		return nil, err
	}
	return config, nil
}

// buildCustomResourceInformerFactory build crd informer factory according some options
func buildCustomResourceInformerFactory(crClient crclientset.Interface) crinformers.SharedInformerFactory {
	var factoryOpts []crinformers.SharedInformerOption
	factoryOpts = append(factoryOpts, crinformers.WithNamespace(apicorev1.NamespaceAll))
	factoryOpts = append(factoryOpts, crinformers.WithTweakListOptions(func(listOptions *v1.ListOptions) { }))
	return crinformers.NewSharedInformerFactoryWithOptions(crClient, 5*time.Second, factoryOpts...)
}

// buildKubeStandardResourceInformerFactory build a kube informer factory according some options
func buildKubeStandardResourceInformerFactory(kubeClient kubernetes.Interface) informers.SharedInformerFactory {
	var factoryOpts []informers.SharedInformerOption
	factoryOpts = append(factoryOpts, informers.WithNamespace(apicorev1.NamespaceAll))
	factoryOpts = append(factoryOpts, informers.WithTweakListOptions(func(listOptions *v1.ListOptions) { }))
	return informers.NewSharedInformerFactoryWithOptions(kubeClient, 5*time.Second, factoryOpts...)
}
