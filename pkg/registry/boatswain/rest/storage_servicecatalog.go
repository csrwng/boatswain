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

package rest

import (
	"github.com/staebler/boatswain/pkg/api"
	"github.com/staebler/boatswain/pkg/apis/boatswain"
	boatswainv1alpha1 "github.com/staebler/boatswain/pkg/apis/boatswain/v1alpha1"
	"github.com/staebler/boatswain/pkg/registry/boatswain/host"
	"github.com/staebler/boatswain/pkg/storage/etcd"
	"k8s.io/apiserver/pkg/registry/generic"
	"k8s.io/apiserver/pkg/registry/rest"
	genericapiserver "k8s.io/apiserver/pkg/server"
	serverstorage "k8s.io/apiserver/pkg/server/storage"
	"k8s.io/apiserver/pkg/storage"
	restclient "k8s.io/client-go/rest"
)

// StorageProvider provides a factory method to create a new APIGroupInfo for
// the boatswain API group. It implements (./pkg/apiserver).RESTStorageProvider
type StorageProvider struct {
	DefaultNamespace string
	StorageType      server.StorageType
	RESTClient       restclient.Interface
}

// NewRESTStorage is a factory method to make a new APIGroupInfo for the
// boatswain API group.
func (p StorageProvider) NewRESTStorage(
	apiResourceConfigSource serverstorage.APIResourceConfigSource,
	restOptionsGetter generic.RESTOptionsGetter,
) (*genericapiserver.APIGroupInfo, error) {

	storage, err := p.v1alpha1Storage(apiResourceConfigSource, restOptionsGetter)
	if err != nil {
		return nil, err
	}

	apiGroupInfo := genericapiserver.NewDefaultAPIGroupInfo(boatswain.GroupName, api.Registry, api.Scheme, api.ParameterCodec, api.Codecs)
	apiGroupInfo.GroupMeta.GroupVersion = boatswainv1alpha1.SchemeGroupVersion

	apiGroupInfo.VersionedResourcesStorageMap = map[string]map[string]rest.Storage{
		boatswainv1alpha1.SchemeGroupVersion.Version: storage,
	}

	return &apiGroupInfo, nil
}

func (p StorageProvider) v1alpha1Storage(
	apiResourceConfigSource serverstorage.APIResourceConfigSource,
	restOptionsGetter generic.RESTOptionsGetter,
) (map[string]rest.Storage, error) {
	hostRESTOptions, err := restOptionsGetter.GetRESTOptions(boatswain.Resource("hosts"))
	if err != nil {
		return nil, err
	}
	hostOpts := server.NewOptions(
		etcd.Options{
			RESTOptions:   hostRESTOptions,
			Capacity:      1000,
			ObjectType:    host.EmptyObject(),
			ScopeStrategy: host.NewScopeStrategy(),
			NewListFunc:   host.NewList,
			GetAttrsFunc:  host.GetAttrs,
			Trigger:       storage.NoTriggerPublisher,
		},
		p.StorageType,
	)

	hostStorage, hostStatusStorage := host.NewStorage(*hostOpts)

	return map[string]rest.Storage{
		"hosts":        hostStorage,
		"hosts/status": hostStatusStorage,
	}, nil
}

// GroupName returns the API group name.
func (p StorageProvider) GroupName() string {
	return boatswain.GroupName
}