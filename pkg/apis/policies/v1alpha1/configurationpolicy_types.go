// Copyright 2019 The Kubernetes Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// RemediationAction : enforce or inform
type RemediationAction string

// Severity : low, medium or high
type Severity string

const (
	// Enforce is an remediationAction to make changes
	Enforce RemediationAction = "Enforce"

	// Inform is an remediationAction to only inform
	Inform RemediationAction = "Inform"
)

// ComplianceState shows the state of enforcement
type ComplianceState string

const (
	// Compliant is an ComplianceState
	Compliant ComplianceState = "Compliant"

	// NonCompliant is an ComplianceState
	NonCompliant ComplianceState = "NonCompliant"

	// UnknownCompliancy is an ComplianceState
	UnknownCompliancy ComplianceState = "UnknownCompliancy"
)

// Condition is the base struct for representing resource conditions
type Condition struct {
	// Type of condition, e.g Complete or Failed.
	Type string `json:"type"`
	// Status of the condition, one of True, False, Unknown.
	Status corev1.ConditionStatus `json:"status,omitempty" protobuf:"bytes,12,rep,name=status"`
	// The last time the condition transitioned from one status to another.
	// +optional
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty" protobuf:"bytes,3,opt,name=lastTransitionTime"`
	// The reason for the condition's last transition.
	// +optional
	Reason string `json:"reason,omitempty" protobuf:"bytes,4,opt,name=reason"`
	// A human readable message indicating details about the transition.
	// +optional
	Message string `json:"message,omitempty" protobuf:"bytes,5,opt,name=message"`
}

// Target defines the list of namespaces to include/exclude
type Target struct {
	Include []string `json:"include,omitempty"`
	Exclude []string `json:"exclude,omitempty"`
}

// ConfigurationPolicySpec defines the desired state of ConfigurationPolicy
// +k8s:openapi-gen=true
type ConfigurationPolicySpec struct {
	Severity                         Severity          `json:"severity,omitempty"`          //low, medium, high
	RemediationAction                RemediationAction `json:"remediationAction,omitempty"` //enforce, inform
	NamespaceSelector                Target            `json:"namespaceSelector,omitempty"` // selecting a list of namespaces where the policy applies
	LabelSelector                    map[string]string `json:"labelSelector,omitempty"`
	MaxRoleBindingUsersPerNamespace  int               `json:"maxRoleBindingUsersPerNamespace,omitempty"`
	MaxRoleBindingGroupsPerNamespace int               `json:"maxRoleBindingGroupsPerNamespace,omitempty"`
	MaxClusterRoleBindingUsers       int               `json:"maxClusterRoleBindingUsers,omitempty"`
	MaxClusterRoleBindingGroups      int               `json:"maxClusterRoleBindingGroups,omitempty"`
	RoleTemplates                    []*RoleTemplate   `json:"role-templates,omitempty"`
	ObjectTemplates                  []*ObjectTemplate `json:"object-templates,omitempty"`
}

//ObjectTemplate describes how an object should look
type ObjectTemplate struct {
	// ComplianceType specifies wether it is a : //musthave, mustnothave, mustonlyhave
	ComplianceType ComplianceType `json:"complianceType"`

	//Selector *metav1.LabelSelector `json:"selector,omitempty" protobuf:"bytes,2,opt,name=selector"`

	// RoleBinding
	ObjectDefinition runtime.RawExtension `json:"objectDefinition,omitempty"`
	//Status shows the individual status of each template within a policy
	Status TemplateStatus `json:"status,omitempty" protobuf:"bytes,12,rep,name=status"`
}

//RoleTemplate describes how a role should look
type RoleTemplate struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object's metadata.
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	ComplianceType ComplianceType `json:"complianceType,omitempty"`

	Selector *metav1.LabelSelector `json:"selector,omitempty" protobuf:"bytes,2,opt,name=selector"`

	// Rules holds all the PolicyRules for this Role
	Rules []PolicyRuleTemplate `json:"rules,omitempty" protobuf:"bytes,2,rep,name=rules"`

	Status TemplateStatus `json:"status,omitempty" protobuf:"bytes,12,rep,name=status"`
}

// ConfigurationPolicyStatus is the status for a Policy resource
type ConfigurationPolicyStatus struct {
	ComplianceState   ComplianceState                `json:"compliant,omitempty"`         // Compliant, NonCompliant, UnkownCompliancy
	CompliancyDetails map[string]map[string][]string `json:"compliancyDetails,omitempty"` // reason for non-compliancy
}

//CompliancePerClusterStatus contains aggregate status of other policies in cluster
type CompliancePerClusterStatus struct {
	AggregatePolicyStatus map[string]*ConfigurationPolicyStatus `json:"aggregatePoliciesStatus,omitempty"`
	ComplianceState       ComplianceState                       `json:"compliant,omitempty"`
	ClusterName           string                                `json:"clustername,omitempty"`
}

//ComplianceMap map to hold CompliancePerClusterStatus objects
type ComplianceMap map[string]*CompliancePerClusterStatus

//ResourceState genric description of a state
type ResourceState string

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ConfigurationPolicy is the Schema for the configurationpolicies API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=configurationpolicies,scope=Namespaced
type ConfigurationPolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ConfigurationPolicySpec   `json:"spec,omitempty"`
	Status ConfigurationPolicyStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ConfigurationPolicyList contains a list of ConfigurationPolicy
type ConfigurationPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ConfigurationPolicy `json:"items"`
}

// Policy is a specification for a Policy resource
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +genclient
type Policy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
}

// PolicyList is a list of Policy resources
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:lister-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type PolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []Policy `json:"items"`
}

//PolicyTemplate template for custom security policy
type PolicyTemplate struct {
	ObjectDefinition runtime.RawExtension `json:"objectDefinition,omitempty"`
	//Status shows the individual status of each template within a policy
	Status TemplateStatus `json:"status,omitempty" protobuf:"bytes,12,rep,name=status"`
}

//TemplateStatus hold the status result
type TemplateStatus struct {
	ComplianceState ComplianceState `json:"Compliant,omitempty"` // Compliant, NonCompliant, UnkownCompliancy
	// +optional
	// +patchMergeKey=type
	// +patchStrategy=merge
	Conditions []Condition `json:"conditions,omitempty"`

	Validity Validity `json:"Validity,omitempty"` // a template can be invalid if it has conflicting roles
}

//Validity describes if it is valid or not
type Validity struct {
	Valid  *bool  `json:"valid,omitempty"`
	Reason string `json:"reason,omitempty"`
}

//ComplianceType describe whether we must or must not have a given resource
type ComplianceType string

const (
	// MustNotHave is an enforcement state to exclude a resource
	MustNotHave ComplianceType = "Mustnothave"

	// MustHave is an enforcement state to include a resource
	MustHave ComplianceType = "Musthave"

	// MustOnlyHave is an enforcement state to exclusively include a resource
	MustOnlyHave ComplianceType = "Mustonlyhave"
)

// PolicyRuleTemplate holds information that describes a policy rule, but does not contain information
// about who the rule applies to or which namespace the rule applies to. We added the compliance type to it for HCM
type PolicyRuleTemplate struct {
	// ComplianceType specifies wether it is a : //musthave, mustnothave, mustonlyhave
	ComplianceType ComplianceType `json:"complianceType"`
	// PolicyRule
	PolicyRule rbacv1.PolicyRule `json:"policyRule"`
}

func init() {
	SchemeBuilder.Register(&ConfigurationPolicy{}, &ConfigurationPolicyList{})
}
