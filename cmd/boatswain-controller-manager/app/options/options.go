/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// The controller is responsible for running control loops that reconcile
// the state of boatswain API resources with service brokers, service
// classes, service instances, and service instance credentials.

package options

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/pflag"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/apimachinery/pkg/util/sets"
	utilfeature "k8s.io/apiserver/pkg/util/feature"

	k8scomponentconfig "github.com/staebler/boatswain/pkg/kubernetes/pkg/apis/componentconfig"
	"github.com/staebler/boatswain/pkg/kubernetes/pkg/client/leaderelectionconfig"

	"github.com/staebler/boatswain/pkg/apis/componentconfig"
)

// CMServer is the main context object for the controller manager.
type CMServer struct {
	componentconfig.ControllerManagerConfiguration
}

const (
	defaultContentType             = "application/json"
	defaultBindAddress             = "0.0.0.0"
	defaultPort                    = 10000
	defaultK8sKubeconfigPath       = "./kubeconfig"
	defaultBoatswainKubeconfigPath = "./boatswain-kubeconfig"
	defaultConcurrentSyncs         = 5
	defaultLeaderElectionNamespace = "kube-system"
)

// NewCMServer creates a new CMServer with a default config.
func NewCMServer() *CMServer {
	s := CMServer{
		ControllerManagerConfiguration: componentconfig.ControllerManagerConfiguration{
			Controllers:               []string{"*"},
			Address:                   defaultBindAddress,
			Port:                      defaultPort,
			ContentType:               defaultContentType,
			K8sKubeconfigPath:         defaultK8sKubeconfigPath,
			BoatswainKubeconfigPath:   defaultBoatswainKubeconfigPath,
			MinResyncPeriod:           metav1.Duration{Duration: 12 * time.Hour},
			ConcurrentClusterSyncs:    defaultConcurrentSyncs,
			ConcurrentNodeGroupSyncs:  defaultConcurrentSyncs,
			ConcurrentNodeSyncs:       defaultConcurrentSyncs,
			ConcurrentMasterNodeSyncs: defaultConcurrentSyncs,
			LeaderElection:            leaderelectionconfig.DefaultLeaderElectionConfiguration(),
			LeaderElectionNamespace:   defaultLeaderElectionNamespace,
			ControllerStartInterval:   metav1.Duration{Duration: 0 * time.Second},
			EnableProfiling:           true,
			EnableContentionProfiling: false,
		},
	}
	s.LeaderElection.LeaderElect = true
	return &s
}

// AddFlags adds flags for a ControllerManagerServer to the specified FlagSet.
func (s *CMServer) AddFlags(fs *pflag.FlagSet, allControllers []string, disabledByDefaultControllers []string) {
	fs.StringSliceVar(&s.Controllers, "controllers", s.Controllers, fmt.Sprintf(""+
		"A list of controllers to enable.  '*' enables all on-by-default controllers, 'foo' enables the controller "+
		"named 'foo', '-foo' disables the controller named 'foo'.\nAll controllers: %s\nDisabled-by-default controllers: %s",
		strings.Join(allControllers, ", "), strings.Join(disabledByDefaultControllers, ", ")))
	fs.Var(k8scomponentconfig.IPVar{Val: &s.Address}, "address", "The IP address to serve on (set to 0.0.0.0 for all interfaces)")
	fs.Int32Var(&s.Port, "port", s.Port, "The port that the controller-manager's http service runs on")
	fs.StringVar(&s.ContentType, "api-content-type", s.ContentType, "Content type of requests sent to API servers")
	fs.StringVar(&s.K8sAPIServerURL, "k8s-api-server-url", "", "The URL for the k8s API server")
	fs.StringVar(&s.K8sKubeconfigPath, "k8s-kubeconfig", "", "Path to k8s core kubeconfig")
	fs.StringVar(&s.BoatswainAPIServerURL, "boatswain-api-server-url", "", "The URL for the boatswain API server")
	fs.StringVar(&s.BoatswainKubeconfigPath, "boatswain-kubeconfig", "", "Path to boatswain kubeconfig")
	fs.BoolVar(&s.BoatswainInsecureSkipVerify, "boatswain-insecure-skip-verify", s.BoatswainInsecureSkipVerify, "Skip verification of the TLS certificate for the boatswain API server")
	fs.DurationVar(&s.MinResyncPeriod.Duration, "min-resync-period", s.MinResyncPeriod.Duration, "The resync period in reflectors will be random between MinResyncPeriod and 2*MinResyncPeriod")
	fs.Int32Var(&s.ConcurrentClusterSyncs, "concurrent-cluster-syncs", s.ConcurrentClusterSyncs, "The number of cluster objects that are allowed to sync concurrently. Larger number = more responsive clusters, but more CPU (and network) load")
	fs.Int32Var(&s.ConcurrentNodeGroupSyncs, "concurrent-node-group-syncs", s.ConcurrentNodeGroupSyncs, "The number of node group objects that are allowed to sync concurrently. Larger number = more responsive node groups, but more CPU (and network) load")
	fs.Int32Var(&s.ConcurrentNodeSyncs, "concurrent-node-syncs", s.ConcurrentNodeSyncs, "The number of node objects that are allowed to sync concurrently. Larger number = more responsive nodes, but more CPU (and network) load")
	fs.Int32Var(&s.ConcurrentMasterNodeSyncs, "concurrent-master-node-syncs", s.ConcurrentMasterNodeSyncs, "The number of master node objects that are allowed to sync concurrently. Larger number = more responsive master nodes, but more CPU (and network) load")
	fs.BoolVar(&s.EnableProfiling, "profiling", s.EnableProfiling, "Enable profiling via web interface host:port/debug/pprof/")
	fs.BoolVar(&s.EnableContentionProfiling, "contention-profiling", s.EnableContentionProfiling, "Enable lock contention profiling, if profiling is enabled")
	leaderelectionconfig.BindFlags(&s.LeaderElection, fs)
	fs.StringVar(&s.LeaderElectionNamespace, "leader-election-namespace", s.LeaderElectionNamespace, "Namespace to use for leader election lock")
	fs.DurationVar(&s.ControllerStartInterval.Duration, "controller-start-interval", s.ControllerStartInterval.Duration, "Interval between starting controller managers.")

	utilfeature.DefaultFeatureGate.AddFlag(fs)
}

// Validate is used to validate the options and config before launching the controller manager
func (s *CMServer) Validate(allControllers []string, disabledByDefaultControllers []string) error {
	var errs []error

	allControllersSet := sets.NewString(allControllers...)
	for _, controller := range s.Controllers {
		if controller == "*" {
			continue
		}
		if strings.HasPrefix(controller, "-") {
			controller = controller[1:]
		}

		if !allControllersSet.Has(controller) {
			errs = append(errs, fmt.Errorf("%q is not in the list of known controllers", controller))
		}
	}

	return utilerrors.NewAggregate(errs)
}
