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

package cluster

import (
	"errors"
	"fmt"

	"github.com/staebler/boatswain/pkg/api"
	boatswainmeta "github.com/staebler/boatswain/pkg/api/meta"
	"github.com/staebler/boatswain/pkg/apis/boatswain"
	"github.com/staebler/boatswain/pkg/storage/etcd"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	genericapirequest "k8s.io/apiserver/pkg/endpoints/request"
	"k8s.io/apiserver/pkg/registry/generic"
	"k8s.io/apiserver/pkg/registry/generic/registry"
	"k8s.io/apiserver/pkg/registry/rest"
	"k8s.io/apiserver/pkg/storage"
)

var (
	errNotACluster = errors.New("not a cluster")
)

// NewSingular returns a new shell of a service cluster, according to the given namespace and
// name
func NewSingular(ns, name string) runtime.Object {
	return &boatswain.Cluster{
		TypeMeta: metav1.TypeMeta{
			Kind: "Cluster",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: ns,
			Name:      name,
		},
	}
}

// EmptyObject returns an empty cluster
func EmptyObject() runtime.Object {
	return &boatswain.Cluster{}
}

// NewList returns a new shell of a cluster list
func NewList() runtime.Object {
	return &boatswain.ClusterList{
		TypeMeta: metav1.TypeMeta{
			Kind: "ClusterList",
		},
		Items: []boatswain.Cluster{},
	}
}

// CheckObject returns a non-nil error if obj is not a cluster object
func CheckObject(obj runtime.Object) error {
	_, ok := obj.(*boatswain.Cluster)
	if !ok {
		return errNotACluster
	}
	return nil
}

// Match determines whether an Cluster matches a field and label selector.
func Match(label labels.Selector, field fields.Selector) storage.SelectionPredicate {
	return storage.SelectionPredicate{
		Label:    label,
		Field:    field,
		GetAttrs: GetAttrs,
	}
}

// toSelectableFields returns a field set that represents the object for matching purposes.
func toSelectableFields(cluster *boatswain.Cluster) fields.Set {
	objectMetaFieldsSet := generic.ObjectMetaFieldsSet(&cluster.ObjectMeta, true)
	return generic.MergeFieldsSets(objectMetaFieldsSet, nil)
}

// GetAttrs returns labels and fields of a given object for filtering purposes.
func GetAttrs(obj runtime.Object) (labels.Set, fields.Set, bool, error) {
	cluster, ok := obj.(*boatswain.Cluster)
	if !ok {
		return nil, nil, false, fmt.Errorf("given object is not a Cluster")
	}
	return labels.Set(cluster.ObjectMeta.Labels), toSelectableFields(cluster), cluster.Initializers != nil, nil
}

// NewStorage creates a new rest.Storage responsible for accessing
// Cluster resources
func NewStorage(opts etcd.Options) (clusters, clustersStatus rest.Storage) {
	restOpts := opts.RESTOptions

	prefix := "/" + restOpts.ResourcePrefix

	storageInterface, dFunc := restOpts.Decorator(
		api.Scheme,
		restOpts.StorageConfig,
		&boatswain.Cluster{},
		prefix,
		nil, /* keyFunc for decorator -- looks to be unused everywhere */
		NewList,
		nil,
		storage.NoTriggerPublisher,
	)

	store := registry.Store{
		NewFunc:     EmptyObject,
		NewListFunc: NewList,
		KeyRootFunc: func(ctx genericapirequest.Context) string {
			return registry.NamespaceKeyRootFunc(ctx, prefix)
		},
		KeyFunc: func(ctx genericapirequest.Context, name string) (string, error) {
			return registry.NamespaceKeyFunc(ctx, prefix, name)
		},
		// Retrieve the name field of the resource.
		ObjectNameFunc: func(obj runtime.Object) (string, error) {
			return boatswainmeta.GetAccessor().Name(obj)
		},
		// Used to match objects based on labels/fields for list.
		PredicateFunc: Match,
		// DefaultQualifiedResource should always be plural
		DefaultQualifiedResource: boatswain.Resource("clusters"),

		CreateStrategy:          clusterRESTStrategies,
		UpdateStrategy:          clusterRESTStrategies,
		DeleteStrategy:          clusterRESTStrategies,
		EnableGarbageCollection: true,

		Storage:     storageInterface,
		DestroyFunc: dFunc,
	}

	options := &generic.StoreOptions{RESTOptions: restOpts, AttrFunc: GetAttrs}
	if err := store.CompleteWithOptions(options); err != nil {
		panic(err) // TODO: Propagate error up
	}

	statusStore := store
	statusStore.UpdateStrategy = clusterStatusUpdateStrategy

	return &store, &StatusREST{&statusStore}
}

// StatusREST defines the REST operations for the status subresource via
// implementation of various rest interfaces.  It supports the http verbs GET,
// PATCH, and PUT.
type StatusREST struct {
	store *registry.Store
}

// New returns a new Cluster.
func (r *StatusREST) New() runtime.Object {
	return &boatswain.Cluster{}
}

// Get retrieves the object from the storage. It is required to support Patch
// and to implement the rest.Getter interface.
func (r *StatusREST) Get(ctx genericapirequest.Context, name string, options *metav1.GetOptions) (runtime.Object, error) {
	return r.store.Get(ctx, name, options)
}

// Update alters the status subset of an object and implements the
// rest.Updater interface.
func (r *StatusREST) Update(ctx genericapirequest.Context, name string, objInfo rest.UpdatedObjectInfo) (runtime.Object, bool, error) {
	return r.store.Update(ctx, name, objInfo)
}
