// Copyright (c) 2020 Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project

package e2e

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/open-cluster-management/config-policy-controller/test/utils"
)

const case13Secret string = "e2esecret"
const case13SecretCopy string = "e2esecret2"
const case13SecretYaml string = "../resources/case13_templatization/case13_secret.yaml"
const case13CfgPolCreateSecret string = "tmplt-policy-secret-duplicate"
const case13CfgPolCheckSecret string = "tmplt-policy-secret-duplicate-check"
const case13CfgPolCreateSecretYaml string = "../resources/case13_templatization/case13_copysecret.yaml"
const case13CfgPolCheckSecretYaml string = "../resources/case13_templatization/case13_verifysecret.yaml"

const case13ClusterClaim string = "testclaim.open-cluster-management.io"
const case13ClusterClaimYaml string = "../resources/case13_templatization/case13_clusterclaim.yaml"
const case13CfgPolVerifyPod string = "policy-pod-templatized-name-verify"
const case13CfgPolCreatePod string = "policy-pod-templatized-name"
const case13CfgPolCreatePodYaml string = "../resources/case13_templatization/case13_pod_nameFromClusterClaim.yaml"
const case13CfgPolVerifyPodYaml string = "../resources/case13_templatization/case13_pod_name_verify.yaml"

const case13LookupSecret string = "tmplt-policy-secret-lookup-check"
const case13LookupSecretYaml string = "../resources/case13_templatization/case13_lookup_secret.yaml"
const case13LookupClusterClaim string = "policy-pod-lookup-verify"
const case13LookupClusterClaimYaml string = "../resources/case13_templatization/case13_lookup_cc.yaml"

var _ = Describe("Test templatization", func() {
	Describe("Create a secret and pull data from it into a configurationPolicy", func() {
		It("should be created properly on the managed cluster", func() {
			By("Creating " + case13CfgPolCreateSecret + " and " + case13CfgPolCheckSecret + " on managed")
			//create secret
			utils.Kubectl("apply", "-f", case13SecretYaml, "-n", "default")
			secret := utils.GetWithTimeout(clientManagedDynamic, gvrSecret, case13Secret, "default", true, defaultTimeoutSeconds)
			Expect(secret).NotTo(BeNil())
			//create copy with password from original secret using a templatized policy
			utils.Kubectl("apply", "-f", case13CfgPolCreateSecretYaml, "-n", testNamespace)
			plc := utils.GetWithTimeout(clientManagedDynamic, gvrConfigPolicy, case13CfgPolCreateSecret, testNamespace, true, defaultTimeoutSeconds)
			Expect(plc).NotTo(BeNil())
			Eventually(func() interface{} {
				managedPlc := utils.GetWithTimeout(clientManagedDynamic, gvrConfigPolicy, case13CfgPolCreateSecret, testNamespace, true, defaultTimeoutSeconds)
				return utils.GetComplianceState(managedPlc)
			}, defaultTimeoutSeconds, 1).Should(Equal("Compliant"))
			Eventually(func() interface{} {
				copiedSecret := utils.GetWithTimeout(clientManagedDynamic, gvrSecret, case13Secret, "default", true, defaultTimeoutSeconds)
				return utils.GetFieldFromSecret(copiedSecret, "PASSWORD")
			}, defaultTimeoutSeconds, 1).Should(Equal("MWYyZDFlMmU2N2Rm"))
			//check copied secret with a templatized inform policy
			utils.Kubectl("apply", "-f", case13CfgPolCheckSecretYaml, "-n", testNamespace)
			plc = utils.GetWithTimeout(clientManagedDynamic, gvrConfigPolicy, case13CfgPolCheckSecret, testNamespace, true, defaultTimeoutSeconds)
			Expect(plc).NotTo(BeNil())
			Eventually(func() interface{} {
				managedPlc := utils.GetWithTimeout(clientManagedDynamic, gvrConfigPolicy, case13CfgPolCheckSecret, testNamespace, true, defaultTimeoutSeconds)
				return utils.GetComplianceState(managedPlc)
			}, defaultTimeoutSeconds, 1).Should(Equal("Compliant"))
		})
	})
	Describe("Create a clusterclaim and pull data from it into a configurationPolicy", func() {
		It("should be created properly on the managed cluster", func() {
			By("Creating " + case13CfgPolCreatePod + " and " + case13CfgPolVerifyPod + " on managed")
			//create clusterclaim
			utils.Kubectl("apply", "-f", case13ClusterClaimYaml)
			cc := utils.GetClusterLevelWithTimeout(clientManagedDynamic, gvrClusterClaim, case13ClusterClaim, true, defaultTimeoutSeconds)
			Expect(cc).NotTo(BeNil())
			//create pod named after value from clusterclaim using a templatized policy
			utils.Kubectl("apply", "-f", case13CfgPolCreatePodYaml, "-n", testNamespace)
			plc := utils.GetWithTimeout(clientManagedDynamic, gvrConfigPolicy, case13CfgPolCreatePod, testNamespace, true, defaultTimeoutSeconds)
			Expect(plc).NotTo(BeNil())
			Eventually(func() interface{} {
				managedPlc := utils.GetWithTimeout(clientManagedDynamic, gvrConfigPolicy, case13CfgPolCreatePod, testNamespace, true, defaultTimeoutSeconds)
				return utils.GetComplianceState(managedPlc)
			}, defaultTimeoutSeconds, 1).Should(Equal("Compliant"))
			//check copied value with an inform policy
			utils.Kubectl("apply", "-f", case13CfgPolVerifyPodYaml, "-n", testNamespace)
			plc = utils.GetWithTimeout(clientManagedDynamic, gvrConfigPolicy, case13CfgPolVerifyPod, testNamespace, true, defaultTimeoutSeconds)
			Expect(plc).NotTo(BeNil())
			Eventually(func() interface{} {
				managedPlc := utils.GetWithTimeout(clientManagedDynamic, gvrConfigPolicy, case13CfgPolVerifyPod, testNamespace, true, defaultTimeoutSeconds)
				return utils.GetComplianceState(managedPlc)
			}, defaultTimeoutSeconds, 1).Should(Equal("Compliant"))
		})
	})
	Describe("Use the generic lookup template to get the same resources from the previous tests", func() {
		It("should match the values pulled by resource-specific functions", func() {
			By("Creating inform policies on managed")
			//create inform policy to check secret using generic lookup
			utils.Kubectl("apply", "-f", case13LookupSecretYaml, "-n", testNamespace)
			plc := utils.GetWithTimeout(clientManagedDynamic, gvrConfigPolicy, case13LookupSecret, testNamespace, true, defaultTimeoutSeconds)
			Expect(plc).NotTo(BeNil())
			Eventually(func() interface{} {
				managedPlc := utils.GetWithTimeout(clientManagedDynamic, gvrConfigPolicy, case13LookupSecret, testNamespace, true, defaultTimeoutSeconds)
				return utils.GetComplianceState(managedPlc)
			}, defaultTimeoutSeconds, 1).Should(Equal("Compliant"))
			//create inform policy to check clusterclaim using generic lookup
			utils.Kubectl("apply", "-f", case13LookupClusterClaimYaml, "-n", testNamespace)
			plc = utils.GetWithTimeout(clientManagedDynamic, gvrConfigPolicy, case13LookupClusterClaim, testNamespace, true, defaultTimeoutSeconds)
			Expect(plc).NotTo(BeNil())
			Eventually(func() interface{} {
				managedPlc := utils.GetWithTimeout(clientManagedDynamic, gvrConfigPolicy, case13LookupClusterClaim, testNamespace, true, defaultTimeoutSeconds)
				return utils.GetComplianceState(managedPlc)
			}, defaultTimeoutSeconds, 1).Should(Equal("Compliant"))
		})
	})
})
