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

// Package options contains flags and options for initializing a boatswain apiserver
package options

import (
	utilnet "k8s.io/apimachinery/pkg/util/net"
	genericoptions "k8s.io/apiserver/pkg/server/options"
	//	"k8s.io/kubernetes/pkg/api"
	//	"k8s.io/kubernetes/pkg/api/validation"
	//	kubeoptions "k8s.io/kubernetes/pkg/kubeapiserver/options"
	//	kubeletclient "k8s.io/kubernetes/pkg/kubelet/client"
	//	"k8s.io/kubernetes/pkg/master/ports"
	"k8s.io/apiserver/pkg/storage/storagebackend"

	// add the kubernetes feature gates
	//	_ "k8s.io/kubernetes/pkg/features"

	"github.com/spf13/pflag"

	"github.com/staebler/boatswain/pkg/api"
)

const (
	// Store generated SSL certificates in a place that won't collide with the
	// k8s core API server.
	certDirectory = "/var/run/openshift-boatswain"

	// DefaultEtcdPathPrefix is the default prefix that is prepended to all
	// resource paths in etcd.  It is intended to allow an operator to
	// differentiate the storage of different API servers from one another in
	// a single etcd.
	DefaultEtcdPathPrefix = "/boatswain"
)

// DefaultServiceNodePortRange is the default port range for NodePort services.
var DefaultServiceNodePortRange = utilnet.PortRange{Base: 30000, Size: 2768}

// BoatswainServerRunOptions runs a boatswain api server.
type BoatswainServerRunOptions struct {
	GenericServerRunOptions *genericoptions.ServerRunOptions
	Etcd                    *genericoptions.EtcdOptions
	SecureServing           *genericoptions.SecureServingOptions
	//InsecureServing         *kubeoptions.InsecureServingOptions
	Audit          *genericoptions.AuditOptions
	Features       *genericoptions.FeatureOptions
	Admission      *genericoptions.AdmissionOptions
	Authentication *genericoptions.DelegatingAuthenticationOptions
	Authorization  *genericoptions.DelegatingAuthorizationOptions

	EnableLogsHandler bool
	MasterCount       int
	// DisableAuth disables delegating authentication and authorization for testing scenarios
	DisableAuth bool
}

// NewServerRunOptions creates a new ServerRunOptions object with default parameters
func NewServerRunOptions() *BoatswainServerRunOptions {
	s := BoatswainServerRunOptions{
		GenericServerRunOptions: genericoptions.NewServerRunOptions(),
		Etcd:          NewEtcdOptions(),
		SecureServing: genericoptions.NewSecureServingOptions(),
		//InsecureServing:      kubeoptions.NewInsecureServingOptions(),
		Audit:          genericoptions.NewAuditOptions(),
		Features:       genericoptions.NewFeatureOptions(),
		Admission:      genericoptions.NewAdmissionOptions(),
		Authentication: genericoptions.NewDelegatingAuthenticationOptions(),
		Authorization:  genericoptions.NewDelegatingAuthorizationOptions(),

		EnableLogsHandler: true,
		MasterCount:       1,
	}
	// Set generated SSL cert path correctly
	s.SecureServing.ServerCert.CertDirectory = certDirectory

	// register all admission plugins
	RegisterAllAdmissionPlugins(s.Admission.Plugins)
	return &s
}

// AddFlags adds flags for a specific APIServer to the specified FlagSet
func (s *BoatswainServerRunOptions) AddFlags(fs *pflag.FlagSet) {
	// Add the generic flags.
	s.GenericServerRunOptions.AddUniversalFlags(fs)
	s.Etcd.AddFlags(fs)
	s.SecureServing.AddFlags(fs)
	s.SecureServing.AddDeprecatedFlags(fs)
	//s.InsecureServing.AddFlags(fs)
	//s.InsecureServing.AddDeprecatedFlags(fs)
	s.Audit.AddFlags(fs)
	s.Features.AddFlags(fs)
	s.Authentication.AddFlags(fs)
	s.Authorization.AddFlags(fs)
	s.Admission.AddFlags(fs)

	// Note: the weird ""+ in below lines seems to be the only way to get gofmt to
	// arrange these text blocks sensibly. Grrr.

	fs.BoolVar(&s.EnableLogsHandler, "enable-logs-handler", s.EnableLogsHandler,
		"If true, install a /logs handler for the apiserver logs.")

	fs.IntVar(&s.MasterCount, "apiserver-count", s.MasterCount,
		"The number of apiservers running in the cluster, must be a positive number.")

	fs.BoolVar(&s.DisableAuth, "disable-auth", false,
		"Disable authentication and authorization for testing purposes")
}

// NewEtcdOptions creates a new, empty, EtcdOptions instance
func NewEtcdOptions() *genericoptions.EtcdOptions {
	return genericoptions.NewEtcdOptions(storagebackend.NewDefaultConfig(DefaultEtcdPathPrefix, api.Scheme, nil))
}
