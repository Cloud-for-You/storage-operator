/*
Copyright 2023.

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
	"fmt"
	"os"
	"regexp"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	storagev1 "github.com/Cloud-for-You/storage-operator/api/v1"
	"github.com/Cloud-for-You/storage-operator/pkg/nfsclient"
	"github.com/Cloud-for-You/storage-operator/pkg/provisioner"
	provisioning_plugin "github.com/Cloud-for-You/storage-operator/pkg/provisioner/plugins"
	sc "github.com/Cloud-for-You/storage-operator/pkg/storageclass"
	"github.com/go-logr/logr"
)

const (
	_storageClass string = "storage-operator"
	nfsFinalizer  string = "storage.cfy.cz/finalizer"
)

// NfsReconciler reconciles a Nfs object
type NfsReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Log    logr.Logger
}

//+kubebuilder:rbac:groups=storage.cfy.cz,resources=nfs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=storage.cfy.cz,resources=nfs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=storage.cfy.cz,resources=nfs/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Nfs object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *NfsReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	// TODO(user): your logic here
	log := log.FromContext(ctx)
	log.Info("Verify if a CRD of Nfs exists")

	// Fetch the Nfs instance
	nfs := &storagev1.Nfs{}
	err := r.Get(ctx, req.NamespacedName, nfs)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("Nfs resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		log.Error(err, "Failed to get Nfs")
		return ctrl.Result{}, err
	}

	// Check if the Nfs instance is Marked to be deleted
	isNfsMarkedToBeDeleted := nfs.GetDeletionTimestamp() != nil
	if isNfsMarkedToBeDeleted {
		if controllerutil.ContainsFinalizer(nfs, nfsFinalizer) {
			// Run finalization logic
			if err := r.finalizeNfs(nfs); err != nil {
				return ctrl.Result{}, err
			}

			// Remove finalizer
			controllerutil.RemoveFinalizer(nfs, nfsFinalizer)
			err := r.Update(ctx, nfs)
			if err != nil {
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}

	// Update status to Pending
	if nfs.Status.Phase == "" {
		phase := storagev1.PhasePending
		if err := r.setStatus(ctx, nfs, &phase, nil, nil, nil); err != nil {
			log.Error(err, "Failed to update Nfs status")
		}
	}

	// Create/Update/Delete PersistentVolumeClaim
	foundPVC := &v1.PersistentVolumeClaim{}
	err = r.Get(ctx, types.NamespacedName{Name: nfs.Name, Namespace: nfs.Namespace}, foundPVC)
	if err != nil && errors.IsNotFound(err) {
		pvc, err := r.pvcForNfs(nfs)
		if err != nil {
			log.Error(err, "Failed to create new PVC", "PVC.Namespace", pvc.Namespace, "PVC.Name", pvc.Name)
			return ctrl.Result{}, err
		}
		log.Info("Creating a new PVC", "PVC.Namespace", pvc.Namespace, "PVC.Name", pvc.Name)
		err = r.Create(ctx, pvc)
		if err != nil {
			log.Error(err, "Failed to create new PVC", "PVC.Namespace", pvc.Namespace, "PVC.Name", pvc.Name)
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Failed to get PVC")
		return ctrl.Result{}, err
	}

	// Create/Update/Delete PersistentVolume
	foundPV := &v1.PersistentVolume{}
	err = r.Get(ctx, types.NamespacedName{Name: nfs.Namespace + "-" + nfs.Name, Namespace: ""}, foundPV)
	if err != nil && errors.IsNotFound(err) {
		// Provolani automatizace za predpokladu, ze je na provisioner automatizace zapnuta
		// Get provisioner and parameters from StorageClass
		storageClass, err := sc.GetStorageClass("nfs")
		if err != nil {
			log.Error(err, "not found storageclass")
		}
		provisionerName := regexp.MustCompile(`[^/]+$`).FindString(storageClass.Provisioner)
		if provisioner.IsSupportProvisioner(provisionerName) {
			if nfs.Status.Automation == storagev1.AutomationError {
				return ctrl.Result{}, err
			}
			if nfs.Status.Automation == "" {
				var selectedPlugin provisioner.Plugin
				var jobParameters provisioner.JobParameters

				switch provisionerName {
				case "awx":
					selectedPlugin = &provisioning_plugin.AWXPlugin{}
				case "generic":
					selectedPlugin = &provisioning_plugin.GenericPlugin{}
				}

				jobParameters.Limit = storageClass.Parameters["hosts"]
				jobParameters.ExtraVars.K8s = true
				jobParameters.ExtraVars.ClusterName = os.Getenv("CLUSTER_NAME")
				jobParameters.ExtraVars.NamespaceName = nfs.Namespace
				jobParameters.ExtraVars.PvcName = nfs.Name
				jobParameters.ExtraVars.PvcSize = nfs.Spec.Capacity

				automation, err := selectedPlugin.Run(storageClass.Parameters["job-template-id"], jobParameters)
				if err != nil {
					automationStatus := storagev1.AutomationError
					message := fmt.Sprintf("Automation [%s]: %v", provisionerName, err.Error())
					messagePtr := &message
					if err := r.setStatus(ctx, nfs, nil, nil, &automationStatus, messagePtr); err != nil {
						log.Error(err, "Failed to update Nfs status")
					}
					return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
				}
				automationStatus := automation.Status
				if err := r.setStatus(ctx, nfs, nil, nil, &automationStatus, nil); err != nil {
					log.Error(err, "Failed to update Nfs status")
				}
				return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
			}
			if nfs.Status.Automation == storagev1.AutomationRunning {
				automationStatus := storagev1.AutomationCompleted
				if err := r.setStatus(ctx, nfs, nil, nil, &automationStatus, nil); err != nil {
					log.Error(err, "Failed to update Nfs status")
				}
				return ctrl.Result{}, err
			}
		} else {
			log.Info("Automation is not supported, skip automation")
		}

		// Overeni ze existuje export path na NFS serveru
		// Verify the existence of spec.path on the Nfs server
		if os.Getenv("CHECK_EXPORTPATH") == "true" {
			log.Info("Validate existing nfs export in NFS server")
			err := r.validateExportPath(nfs)
			if err != nil {
				// Nastavime error message v Nfs, kterou jsme ziskali z checkeru a posleme objekt do rekoncilace,
				// ktera probehne napriklad za 20s
				phase := storagev1.PhaseError
				message := err.Error() // Získání chybové zprávy jako string
				messagePtr := &message
				if err := r.setStatus(ctx, nfs, &phase, nil, nil, messagePtr); err != nil {
					log.Error(err, "Failed to update Nfs status")
				}
				return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
			}
		}

		pv := r.pvForNfs(nfs)
		log.Info("Creating a new PV", "PV.Name", pv.Name)
		err = r.Create(ctx, pv)
		if err != nil {
			log.Error(err, "Failed to create new PV", "PV.Name", pv.Name)
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Failed to get PV")
		return ctrl.Result{}, err
	}

	// Test pro resizing
	requestedSize := resource.MustParse(nfs.Spec.Capacity)
	currentSize := foundPV.Spec.Capacity[v1.ResourceStorage]

	if requestedSize.Cmp(currentSize) > 0 {
		log.Info("Volume resizing")
		if err := r.expandVolume(ctx, foundPV, requestedSize); err != nil {
			log.Error(err, "Failed to expand volume")
			return ctrl.Result{}, err
		}
		foundPVC.Status.Capacity[v1.ResourceStorage] = requestedSize
		if err := r.Status().Update(ctx, foundPVC); err != nil {
			log.Error(err, "Failed to update PVC status")
			return ctrl.Result{}, err
		}
		foundPV.Spec.Capacity[v1.ResourceStorage] = requestedSize
		if err := r.Update(ctx, foundPV); err != nil {
			log.Error(err, "Failed to update PV")
			return ctrl.Result{}, err
		}
	}

	// Add finalizer for all CR
	if !controllerutil.ContainsFinalizer(nfs, nfsFinalizer) {
		controllerutil.AddFinalizer(nfs, nfsFinalizer)
		err = r.Update(ctx, nfs)
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	// Update status.phase over PVC
	if nfs.Status.Phase != storagev1.PhaseBound {
		err = r.Get(ctx, types.NamespacedName{Name: nfs.Name, Namespace: nfs.Namespace}, foundPVC)
		if err != nil {
			return ctrl.Result{Requeue: true}, nil
		}
		phase := string(foundPVC.Status.Phase)
		pvcName := foundPVC.Name
		message := ""
		if err := r.setStatus(ctx, nfs, &phase, &pvcName, nil, &message); err != nil {
			log.Error(err, "Failed to update Nfs status")
			return ctrl.Result{Requeue: true}, nil
		}
	}

	return ctrl.Result{}, nil

}

func (r *NfsReconciler) pvcForNfs(m *storagev1.Nfs) (*v1.PersistentVolumeClaim, error) {
	sc, err := sc.GetStorageClass("nfs")
	if err != nil {
		log.Log.Error(err, "not found storageclass")
	}
	pvc := &v1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.Name,
			Namespace: m.Namespace,
		},
		Spec: v1.PersistentVolumeClaimSpec{
			AccessModes: []v1.PersistentVolumeAccessMode{v1.ReadWriteMany},
			Resources: v1.VolumeResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceStorage: resource.MustParse(m.Spec.Capacity),
				},
			},
			StorageClassName: &sc.Name,
			VolumeName:       m.Namespace + "-" + m.Name,
		},
	}
	if err := ctrl.SetControllerReference(m, pvc, r.Scheme); err != nil {
		log.Log.Error(err, "Failed to set controller reference")
		return nil, fmt.Errorf("failed to set controller reference")
	}
	return pvc, nil
}

func (r *NfsReconciler) pvForNfs(m *storagev1.Nfs) *v1.PersistentVolume {
	fsVolumeMode := v1.PersistentVolumeFilesystem
	sc, err := sc.GetStorageClass("nfs")
	if err != nil {
		log.Log.Error(err, "not found storageclass")
	}
	pv := &v1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name: m.Namespace + "-" + m.Name,
		},
		Spec: v1.PersistentVolumeSpec{
			AccessModes: []v1.PersistentVolumeAccessMode{v1.ReadWriteMany},
			Capacity: v1.ResourceList{
				v1.ResourceStorage: resource.MustParse(m.Spec.Capacity),
			},
			VolumeMode:                    &fsVolumeMode,
			StorageClassName:              sc.Name,
			PersistentVolumeReclaimPolicy: *sc.ReclaimPolicy,
			MountOptions:                  sc.MountOptions,
			PersistentVolumeSource: v1.PersistentVolumeSource{
				NFS: &v1.NFSVolumeSource{
					Server:   m.Spec.Server,
					Path:     m.Spec.Path,
					ReadOnly: false,
				},
			},
			ClaimRef: &v1.ObjectReference{
				Namespace: m.Namespace,
				Name:      m.Name,
			},
		},
	}
	return pv
}

func (r *NfsReconciler) finalizeNfs(m *storagev1.Nfs) error {
	log.Log.Info("Successfully finalize nfs")
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *NfsReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&storagev1.Nfs{}).
		Owns(&v1.PersistentVolumeClaim{}).
		Owns(&v1.PersistentVolume{}).
		Complete(r)
}

func (r *NfsReconciler) expandVolume(ctx context.Context, pv *v1.PersistentVolume, newSize resource.Quantity) error {
	// Implement the logic to expand the volume
	// .....

	// After expanding the volume, update the PV size
	pv.Spec.Capacity[v1.ResourceStorage] = newSize
	err := r.Update(ctx, pv)
	if err != nil {
		return err
	}
	return nil
}

func (r *NfsReconciler) setStatus(ctx context.Context, nfs *storagev1.Nfs, phase, pvcName, automation, message *string) error {
	// Get Actual status
	var currentNfs storagev1.Nfs
	if err := r.Get(ctx, client.ObjectKey{Name: nfs.Name, Namespace: nfs.Namespace}, &currentNfs); err != nil {
		return err
	}
	// Pokud je phase předán, aktualizujte ho
	if phase != nil {
		currentNfs.Status.Phase = *phase
	}
	// Pokud je pvcName předán, aktualizujte ho
	if pvcName != nil {
		currentNfs.Status.PVCName = *pvcName
	}
	// Pokud je automation předán, aktualizujte ho
	if automation != nil {
		currentNfs.Status.Automation = *automation
	}
	// Pokud je message předán, aktualizujte ho
	if message != nil {
		currentNfs.Status.Message = *message
	}

	if err := r.Status().Update(ctx, &currentNfs); err != nil {
		return err
	}

	return nil

}

func (r *NfsReconciler) validateExportPath(nfs *storagev1.Nfs) error {
	mount, err := nfsclient.DialMount(nfs.Spec.Server)
	if err != nil {
		return err
	}
	defer func() {
		if err := mount.Close(); err != nil {
			fmt.Printf("Chyba při zavírání mount: %v\n", err)
		}
	}()

	AllNfsExports, err := mount.Exports()
	if err != nil {
		return err
	}

	if !containsExportPath(AllNfsExports, nfs.Spec.Path) {
		return fmt.Errorf("NFS server does't export the specified directory %s", nfs.Spec.Path)
	}

	return nil
}

// funkce pro ověření přítomnosti exportovaneho Path
func containsExportPath(exports []nfsclient.Export, searchString string) bool {
	for _, export := range exports {
		if export.Directory == searchString {
			return true
		}
	}
	return false
}
