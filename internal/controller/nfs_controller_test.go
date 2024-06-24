/*
Copyright 2024.

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

package controller

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	corev1 "k8s.io/api/core/v1"

	storagev1 "github.com/Cloud-for-You/storage-operator/api/v1"
)

var _ = Describe("Nfs Controller", func() {
	Context("When reconciling a resource", func() {
		const resourceName = "test-nfs-resource"
		const resourceNamespace = "default"
		const timeout = time.Second * 10
		const duration = time.Second * 10
		const interval = time.Millisecond * 250

		ctx := context.Background()

		typeNamespacedName := types.NamespacedName{
			Name:      resourceName,
			Namespace: resourceNamespace,
		}
		nfs := &storagev1.Nfs{}

		BeforeEach(func() {
			By("creating the custom resource for the Kind Nfs")
			err := k8sClient.Get(ctx, typeNamespacedName, nfs)
			if err != nil && errors.IsNotFound(err) {
				resource := &storagev1.Nfs{
					ObjectMeta: metav1.ObjectMeta{
						Name:      resourceName,
						Namespace: resourceNamespace,
					},
					// TODO(user): Specify other spec details if needed.
					Spec: storagev1.NfsSpec{
						Server: "localhost",
						Path:   "/volume1/directory",
					},
				}
				Expect(k8sClient.Create(ctx, resource)).To(Succeed())

				nfsLookupKey := types.NamespacedName{Name: resourceName, Namespace: resourceNamespace}
				createdNfs := &storagev1.Nfs{}
				Eventually(func() bool {
					err := k8sClient.Get(ctx, nfsLookupKey, createdNfs)
					return err == nil
				}, timeout, interval).Should(BeTrue())
				Expect(createdNfs.Spec.Capacity).Should(Equal("1Gi"))

				// Zde provedeme samostatne testy
				// Overime, ze po vytvoreni je stav objektu Nfs status.phase nastaveno na Pending
				By("By checking the Nfs has status Pending")
				Consistently(func() (string, error) {
					err := k8sClient.Get(ctx, nfsLookupKey, createdNfs)
					if err != nil {
						return "", err
					}
					return (createdNfs.Status.Phase), nil
				}, duration, interval).Should(Equal(storagev1.PhasePending))

				// Overime, zda byl controllerem vytvoren objekt PVC a PV podle predem definovaneho jmena
				By("By checking the PVC has created")
				pvcLookupKey := types.NamespacedName{Name: resourceName, Namespace: resourceNamespace}
				createdPvc := &corev1.PersistentVolumeClaim{}
				Eventually(func() bool {
					err := k8sClient.Get(ctx, pvcLookupKey, createdPvc)
					return err == nil
				}, timeout, interval).Should(BeTrue())
				By("By checking the PV has created")
				pvList := &corev1.PersistentVolumeList{}
				Eventually(func() bool {
					err := k8sClient.List(ctx, pvList, client.MatchingFields{"spec.claimRef.name": resourceName})
					return err == nil && len(pvList.Items) > 0
				}, timeout, interval).Should(BeTrue())

				createdPv := pvList.Items[0]
				Expect(createdPv.Spec.ClaimRef.Name).Should(Equal(resourceName))
				Expect(createdPv.Spec.ClaimRef.Namespace).Should(Equal(resourceNamespace))

				// Overime, ze stav objektu Nfs
				// status.phase: Bound
				// status.pvcName: <PVC name>
				By("Checking the Nfs status after ending reconcilation")

			}
		})

		AfterEach(func() {
			// TODO(user): Cleanup logic after each test, like removing the resource instance.
			resource := &storagev1.Nfs{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())

			By("Cleanup the specific resource instance Nfs")
			Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
		})
		It("should successfully reconcile the resource", func() {
			By("Reconciling the created resource")
			controllerReconciler := &NfsReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())
			// TODO(user): Add more specific assertions depending on your controller's reconciliation logic.
			// Example: If you expect a certain status condition after reconciliation, verify it here.

			// Verify the Nfs resource spec.capacity and status.phase
			fetchedNfsResource := &storagev1.Nfs{}
			Expect(k8sClient.Get(ctx, typeNamespacedName, fetchedNfsResource)).To(Succeed())
			Expect(fetchedNfsResource.Spec.Capacity).To(Equal("1Gi"))
		})
	})
})
