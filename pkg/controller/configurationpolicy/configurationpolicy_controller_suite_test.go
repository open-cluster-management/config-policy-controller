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
	stdlog "log"
	"os"
	"path/filepath"
	"testing"

	"github.com/open-cluster-management/config-policy-controller/pkg/apis"
	sub "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
)

var cfg *rest.Config

var samplePolicy = policiesv1.ConfigurationPolicy{
	ObjectMeta: metav1.ObjectMeta{
		Name:      "foo",
		Namespace: "default",
	},
	Spec: policiesv1.ConfigurationPolicySpec{
		Severity: "low",
		NamespaceSelector: policiesv1.Target{
			Include: []string{"default", "kube-*"},
			Exclude: []string{"kube-system"},
		},
		RemediationAction: "inform",
		RoleTemplates: []*policiesv1.RoleTemplate{
			&policiesv1.RoleTemplate{
				TypeMeta: metav1.TypeMeta{
					Kind: "roletemplate",
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "",
					Name:      "operator-role-policy",
				},
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"dev": "true",
					},
				},
				ComplianceType: "musthave",
				Rules: []policiesv1.PolicyRuleTemplate{
					policiesv1.PolicyRuleTemplate{
						ComplianceType: "musthave",
						PolicyRule: sub.PolicyRule{
							APIGroups: []string{"extensions", "apps"},
							Resources: []string{"deployments"},
							Verbs:     []string{"get", "list", "watch", "create", "delete", "patch"},
						},
					},
				},
			},
		},
		ObjectTemplates: []*policiesv1.ObjectTemplate{
			&policiesv1.ObjectTemplate{
				ComplianceType:   "musthave",
				ObjectDefinition: runtime.RawExtension{},
			},
		},
	},
}

func TestMain(m *testing.M) {
	t := &envtest.Environment{
		CRDDirectoryPaths: []string{filepath.Join("..", "..", "..", "deploy", "crds")},
	}
	apis.AddToScheme(scheme.Scheme)

	var err error
	if cfg, err = t.Start(); err != nil {
		stdlog.Fatal(err)
	}

	code := m.Run()
	t.Stop()
	os.Exit(code)
}
