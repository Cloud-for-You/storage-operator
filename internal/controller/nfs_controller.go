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
	"github.com/Cloud-for-You/storage-operator/pkg/setup"
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

	// Check existing mountPath
	mount, err := nfsclient.DialMount(nfs.Spec.Server)
	if err != nil {
		log.Error(err, "unable to dial MOUNT service")
	}
	defer mount.Close()
	AllNfsExports, err := mount.Exports()
	if err != nil {
		log.Error(err, "unable to export volumes")
	}

	if !containsExportPath(AllNfsExports, nfs.Spec.Path) {
		// Nastavime status na nejaky Error a zajistime novou rekoncilaci za cca 10s
		statusUpdate := storagev1.NfsStatus{
			Phase:   "Pending",
			Message: "The NFS server does not export the specified directory " + nfs.Spec.Path + ".",
		}
		nfs.Status = statusUpdate
		if err := r.Status().Update(ctx, nfs); err != nil {
			log.Error(err, "Failed to update Nfs status")
		}
		return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
	}

	// Create/Update/Delete PersistentVolumeClaim
	foundPVC := &v1.PersistentVolumeClaim{}
	err = r.Get(ctx, types.NamespacedName{Name: nfs.Name, Namespace: nfs.Namespace}, foundPVC)
	if err != nil && errors.IsNotFound(err) {
		pvc := r.pvcForNfs(nfs)
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

	// Add finalizer for all CR
	if !controllerutil.ContainsFinalizer(nfs, nfsFinalizer) {
		controllerutil.AddFinalizer(nfs, nfsFinalizer)
		err = r.Update(ctx, nfs)
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	// Update status.phase over PVC
	if nfs.Status.Phase == "" {
		statusUpdate := storagev1.NfsStatus{Phase: "Pending"}
		nfs.Status = statusUpdate
		if err := r.Status().Update(ctx, nfs); err != nil {
			log.Error(err, "Failed to update Nfs status")
		}
		return ctrl.Result{Requeue: true}, nil
	} else {
		err = r.Get(ctx, types.NamespacedName{Name: nfs.Name, Namespace: nfs.Namespace}, foundPVC)
		if err != nil {
			return ctrl.Result{Requeue: true}, nil
		}
		statusUpdate := storagev1.NfsStatus{Phase: string(foundPVC.Status.Phase)}
		nfs.Status = statusUpdate
		if err := r.Status().Update(ctx, nfs); err != nil {
			log.Error(err, "Failed to update Nfs status")
		}
	}

	return ctrl.Result{}, nil
}

func (r *NfsReconciler) pvcForNfs(m *storagev1.Nfs) *v1.PersistentVolumeClaim {
	storageClassName := setup.GetStorageClass()
	pvc := &v1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.Name,
			Namespace: m.Namespace,
		},
		Spec: v1.PersistentVolumeClaimSpec{
			AccessModes: []v1.PersistentVolumeAccessMode{v1.ReadWriteMany},
			Resources: v1.VolumeResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceName(v1.ResourceStorage): resource.MustParse("1Gi"),
				},
			},
			StorageClassName: &storageClassName,
			VolumeName:       m.Namespace + "-" + m.Name,
		},
	}
	ctrl.SetControllerReference(m, pvc, r.Scheme)
	return pvc
}

func (r *NfsReconciler) pvForNfs(m *storagev1.Nfs) *v1.PersistentVolume {
	fsVolumeMode := v1.PersistentVolumeFilesystem
	storageClassName := setup.GetStorageClass()
	pv := &v1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name: m.Namespace + "-" + m.Name,
		},
		Spec: v1.PersistentVolumeSpec{
			AccessModes: []v1.PersistentVolumeAccessMode{v1.ReadWriteMany},
			Capacity: v1.ResourceList{
				v1.ResourceName(v1.ResourceStorage): resource.MustParse("1Gi"),
			},
			VolumeMode:                    &fsVolumeMode,
			StorageClassName:              storageClassName,
			PersistentVolumeReclaimPolicy: v1.PersistentVolumeReclaimRetain,
			MountOptions:                  []string{"nfsvers=4", "hard", "intr"},
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
	ctrl.SetControllerReference(m, pv, r.Scheme)
	return pv
}

func (r *NfsReconciler) finalizeNfs(m *storagev1.Nfs) error {
	log.Log.Info("Successfuly finalize nfs")
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

// funkce pro ověření přítomnosti exportovaneho Path
func containsExportPath(exports []nfsclient.Export, searchString string) bool {
	for _, export := range exports {
		if export.Directory == searchString {
			return true
		}
	}
	return false
}
