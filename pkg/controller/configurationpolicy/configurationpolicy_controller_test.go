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
package configurationpolicy

import (
	"encoding/json"
	"strings"
	"testing"

	policiesv1alpha1 "github.com/open-cluster-management/config-policy-controller/pkg/apis/policies/v1alpha1"
	"github.com/open-cluster-management/config-policy-controller/pkg/common"
	"github.com/stretchr/testify/assert"
	coretypes "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	sub "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	testclient "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var mgr manager.Manager
var err error

func TestReconcile(t *testing.T) {
	var (
		name      = "foo"
		namespace = "default"
	)
	instance := &policiesv1alpha1.ConfigurationPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo",
			Namespace: "default",
		},
		Spec: policiesv1alpha1.ConfigurationPolicySpec{
			Severity: "low",
			NamespaceSelector: policiesv1alpha1.Target{
				Include: []string{"default", "kube-*"},
				Exclude: []string{"kube-system"},
			},
			RemediationAction: "inform",
			ObjectTemplates: []*policiesv1alpha1.ObjectTemplate{
				&policiesv1alpha1.ObjectTemplate{
					ComplianceType:   "musthave",
					ObjectDefinition: runtime.RawExtension{},
				},
			},
		},
	}

	// Objects to track in the fake client.
	objs := []runtime.Object{instance}
	// Register operator types with the runtime scheme.
	s := scheme.Scheme
	s.AddKnownTypes(policiesv1alpha1.SchemeGroupVersion, instance)

	// Create a fake client to mock API calls.
	cl := fake.NewFakeClient(objs...)
	// Create a ReconcileConfigurationPolicy object with the scheme and fake client
	r := &ReconcileConfigurationPolicy{client: cl, scheme: s, recorder: nil}

	// Mock request to simulate Reconcile() being called on an event for a
	// watched resource .
	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      name,
			Namespace: namespace,
		},
	}
	var simpleClient kubernetes.Interface = testclient.NewSimpleClientset()
	common.Initialize(&simpleClient, nil)
	InitializeClient(&simpleClient)
	res, err := r.Reconcile(req)
	if err != nil {
		t.Fatalf("reconcile: (%v)", err)
	}
	t.Log(res)
}

func TestPeriodicallyExecSamplePolicies(t *testing.T) {
	var (
		name      = "foo"
		namespace = "default"
	)
	var typeMeta = metav1.TypeMeta{
		Kind: "namespace",
	}
	var objMeta = metav1.ObjectMeta{
		Name: "default",
	}
	var ns = coretypes.Namespace{
		TypeMeta:   typeMeta,
		ObjectMeta: objMeta,
	}
	var def = map[string]string{
		"apiDefinition": "v1",
		"kind":          "Pod",
	}
	defJSON, err := json.Marshal(def)
	if err != nil {
		t.Log(err)
	}

	// Mock request to simulate Reconcile() being called on an event for a
	// watched resource .
	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      name,
			Namespace: namespace,
		},
	}
	instance := &policiesv1alpha1.ConfigurationPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo",
			Namespace: "default",
		},
		Spec: policiesv1alpha1.ConfigurationPolicySpec{
			Severity: "low",
			NamespaceSelector: policiesv1alpha1.Target{
				Include: []string{"default", "kube-*"},
				Exclude: []string{"kube-system"},
			},
			RemediationAction: "inform",
			ObjectTemplates: []*policiesv1alpha1.ObjectTemplate{
				&policiesv1alpha1.ObjectTemplate{
					ComplianceType: "musthave",
					ObjectDefinition: runtime.RawExtension{
						Raw: defJSON,
					},
				},
			},
		},
	}

	// Objects to track in the fake client.
	objs := []runtime.Object{instance}
	// Register operator types with the runtime scheme.
	s := scheme.Scheme
	s.AddKnownTypes(policiesv1alpha1.SchemeGroupVersion, instance)

	// Create a fake client to mock API calls.
	cl := fake.NewFakeClient(objs...)

	// Create a ReconcileConfigurationPolicy object with the scheme and fake client.
	r := &ReconcileConfigurationPolicy{client: cl, scheme: s, recorder: nil}
	var simpleClient kubernetes.Interface = testclient.NewSimpleClientset()
	simpleClient.CoreV1().Namespaces().Create(&ns)
	common.Initialize(&simpleClient, nil)
	res, err := r.Reconcile(req)
	if err != nil {
		t.Fatalf("reconcile: (%v)", err)
	}
	t.Log(res)
	var target = []string{"default"}
	samplePolicy.Spec.NamespaceSelector.Include = target
	err = handleAddingPolicy(&samplePolicy)
	assert.Nil(t, err)
	PeriodicallyExecSamplePolicies(1, true)
}

func TestCheckUnNamespacedPolicies(t *testing.T) {
	var simpleClient kubernetes.Interface = testclient.NewSimpleClientset()
	common.Initialize(&simpleClient, nil)
	var samplePolicy = policiesv1alpha1.ConfigurationPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo",
			Namespace: "default",
		}}

	var policies = map[string]*policiesv1alpha1.ConfigurationPolicy{}
	policies["policy1"] = &samplePolicy

	err := checkUnNamespacedPolicies(policies)
	assert.Nil(t, err)
}

func TestEnsureDefaultLabel(t *testing.T) {
	updateNeeded := ensureDefaultLabel(&samplePolicy)
	assert.True(t, updateNeeded)

	var labels1 = map[string]string{}
	labels1["category"] = grcCategory
	samplePolicy.Labels = labels1
	updateNeeded = ensureDefaultLabel(&samplePolicy)
	assert.False(t, updateNeeded)

	var labels2 = map[string]string{}
	labels2["category"] = "foo"
	samplePolicy.Labels = labels2
	updateNeeded = ensureDefaultLabel(&samplePolicy)
	assert.True(t, updateNeeded)

	var labels3 = map[string]string{}
	labels3["foo"] = grcCategory
	samplePolicy.Labels = labels3
	updateNeeded = ensureDefaultLabel(&samplePolicy)
	assert.True(t, updateNeeded)
}

func TestCheckAllClusterLevel(t *testing.T) {
	var subject = sub.Subject{
		APIGroup:  "",
		Kind:      "User",
		Name:      "user1",
		Namespace: "default",
	}
	var subjects = []sub.Subject{}
	subjects = append(subjects, subject)
	var clusterRoleBinding = sub.ClusterRoleBinding{
		Subjects: subjects,
	}
	var items = []sub.ClusterRoleBinding{}
	items = append(items, clusterRoleBinding)
	var clusterRoleBindingList = sub.ClusterRoleBindingList{
		Items: items,
	}
	var users, groups = checkAllClusterLevel(&clusterRoleBindingList)
	assert.Equal(t, 1, users)
	assert.Equal(t, 0, groups)
}

func TestCheckViolationsPerNamespace(t *testing.T) {
	var subject = sub.Subject{
		APIGroup:  "",
		Kind:      "User",
		Name:      "user1",
		Namespace: "default",
	}
	var subjects = []sub.Subject{}
	subjects = append(subjects, subject)
	var roleBinding = sub.RoleBinding{
		Subjects: subjects,
	}
	var items = []sub.RoleBinding{}
	items = append(items, roleBinding)
	var roleBindingList = sub.RoleBindingList{
		Items: items,
	}
	var samplePolicySpec = policiesv1alpha1.ConfigurationPolicySpec{
		MaxRoleBindingUsersPerNamespace:  1,
		MaxRoleBindingGroupsPerNamespace: 1,
		MaxClusterRoleBindingUsers:       1,
		MaxClusterRoleBindingGroups:      1,
	}
	samplePolicy.Spec = samplePolicySpec
	checkViolationsPerNamespace(&roleBindingList, &samplePolicy)
}

func TestCreateParentPolicy(t *testing.T) {
	var ownerReference = metav1.OwnerReference{
		Name: "foo",
	}
	var ownerReferences = []metav1.OwnerReference{}
	ownerReferences = append(ownerReferences, ownerReference)
	samplePolicy.OwnerReferences = ownerReferences

	policy := createParentPolicy(&samplePolicy)
	assert.NotNil(t, policy)
	createParentPolicyEvent(&samplePolicy)
}

func TestConvertPolicyStatusToString(t *testing.T) {
	var compliantDetail = policiesv1alpha1.TemplateStatus{
		ComplianceState: policiesv1alpha1.NonCompliant,
		Conditions:      []policiesv1alpha1.Condition{},
	}
	var compliantDetails = []policiesv1alpha1.TemplateStatus{}

	for i := 0; i < 3; i++ {
		compliantDetails = append(compliantDetails, compliantDetail)
	}

	samplePolicyStatus := policiesv1alpha1.ConfigurationPolicyStatus{
		ComplianceState:   "Compliant",
		CompliancyDetails: compliantDetails,
	}
	samplePolicy.Status = samplePolicyStatus
	var policyInString = convertPolicyStatusToString(&samplePolicy)
	assert.NotNil(t, policyInString)
}

func TestDeleteExternalDependency(t *testing.T) {
	mgr, err = manager.New(cfg, manager.Options{})
	reconcileConfigurationPolicy := ReconcileConfigurationPolicy{client: mgr.GetClient(), scheme: mgr.GetScheme(), recorder: mgr.GetEventRecorderFor("samplepolicy-controller")}
	reconcileConfigurationPolicy.deleteExternalDependency(&samplePolicy)
}

func TestHandleAddingPolicy(t *testing.T) {
	var simpleClient kubernetes.Interface = testclient.NewSimpleClientset()
	var typeMeta = metav1.TypeMeta{
		Kind: "namespace",
	}
	var objMeta = metav1.ObjectMeta{
		Name: "default",
	}
	var ns = coretypes.Namespace{
		TypeMeta:   typeMeta,
		ObjectMeta: objMeta,
	}
	simpleClient.CoreV1().Namespaces().Create(&ns)
	common.Initialize(&simpleClient, nil)
	err := handleAddingPolicy(&samplePolicy)
	assert.Nil(t, err)
	handleRemovingPolicy(&samplePolicy)
}

func TestGetContainerID(t *testing.T) {
	var containerStateWaiting = coretypes.ContainerStateWaiting{
		Reason: "unknown",
	}
	var containerState = coretypes.ContainerState{
		Waiting: &containerStateWaiting,
	}
	var containerStatus = coretypes.ContainerStatus{
		State:       containerState,
		ContainerID: "id",
	}
	var containerStatuses []coretypes.ContainerStatus
	containerStatuses = append(containerStatuses, containerStatus)
	var podStatus = coretypes.PodStatus{
		ContainerStatuses: containerStatuses,
	}
	var pod = coretypes.Pod{
		Status: podStatus,
	}
	getContainerID(pod, "foo")
}

func newRule(verbs, apiGroups, resources, nonResourceURLs string) rbacv1.PolicyRule {
	return rbacv1.PolicyRule{
		Verbs:           strings.Split(verbs, ","),
		APIGroups:       strings.Split(apiGroups, ","),
		Resources:       strings.Split(resources, ","),
		NonResourceURLs: strings.Split(nonResourceURLs, ","),
	}
}

func newRole(name, namespace string, rules ...rbacv1.PolicyRule) *rbacv1.Role {
	return &rbacv1.Role{ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: name}, Rules: rules}
}

func newRuleTemplate(verbs, apiGroups, resources, nonResourceURLs string, complianceT policiesv1alpha1.ComplianceType) policiesv1alpha1.PolicyRuleTemplate {
	return policiesv1alpha1.PolicyRuleTemplate{
		ComplianceType: complianceT,
		PolicyRule: rbacv1.PolicyRule{
			Verbs:           strings.Split(verbs, ","),
			APIGroups:       strings.Split(apiGroups, ","),
			Resources:       strings.Split(resources, ","),
			NonResourceURLs: strings.Split(nonResourceURLs, ","),
		},
	}
}
