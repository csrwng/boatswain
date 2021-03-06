/*
Copyright 2016 The Kubernetes Authors.

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

package validation

import (
	apivalidation "k8s.io/apimachinery/pkg/api/validation"
	"k8s.io/apimachinery/pkg/util/validation/field"

	boatswain "github.com/staebler/boatswain/pkg/apis/boatswain"
)

// ValidateNodeGroupName is the validation function for NodeGroup names.
var ValidateNodeGroupName = apivalidation.NameIsDNSSubdomain

// ValidateNodeGroup implements the validation rules for a NodeGroup.
func ValidateNodeGroup(nodeGroup *boatswain.NodeGroup) field.ErrorList {
	allErrs := field.ErrorList{}

	allErrs = append(allErrs,
		apivalidation.ValidateObjectMeta(&nodeGroup.ObjectMeta,
			true, /* namespace required */
			ValidateNodeGroupName,
			field.NewPath("metadata"))...)

	allErrs = append(allErrs, validateNodeGroupSpec(&nodeGroup.Spec, field.NewPath("spec"))...)
	return allErrs
}

func validateNodeGroupSpec(spec *boatswain.NodeGroupSpec, fldPath *field.Path) field.ErrorList {
	return field.ErrorList{}
}

// ValidateNodeGroupUpdate checks that when changing from an older nodegroup to a newer nodegroup is okay ?
func ValidateNodeGroupUpdate(new *boatswain.NodeGroup, old *boatswain.NodeGroup) field.ErrorList {
	return field.ErrorList{}
}

// ValidateNodeGroupStatusUpdate checks that when changing from an older nodegroup to a newer nodegroup is okay.
func ValidateNodeGroupStatusUpdate(new *boatswain.NodeGroup, old *boatswain.NodeGroup) field.ErrorList {
	allErrs := field.ErrorList{}
	allErrs = append(allErrs, ValidateNodeGroupUpdate(new, old)...)
	return allErrs
}
