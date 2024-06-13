package sc

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/Cloud-for-You/storage-operator/pkg/k8sclient"
	v1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
)

var (
	log       = ctrl.Log.WithName("storageclass")
	clientset *kubernetes.Clientset
)

func init() {
	var err error
	clientset, err = k8sclient.K8SClient()
	if err != nil {
		log.Error(err, "failed to create kubernetes clientset")
		os.Exit(1)
	}
}

func VerifyStorageClassesExists(annotation []string) bool {
	for _, annotation := range annotation {
		sc, err := GetStorageClass(annotation)
		if err != nil {
			log.Error(err, "failed to get storageclass")
			os.Exit(1)
		}
		if sc == nil {
			log.Error(err, "storageclass from annotation storage-operator.cfy.cz/storage-type="+annotation+" not found")
			return false
		}
		storageClassJSON, err := json.Marshal(sc)
		if err != nil {
			log.Error(err, "failed to marshal storageclass to json")
			os.Exit(1)
		}
		log.Info(fmt.Sprintf("found storageclass details: %s", storageClassJSON))
	}
	return true
}

func GetStorageClass(v string) (*v1.StorageClass, error) {
	storageClassList, err := clientset.StorageV1().StorageClasses().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Error(err, "Error listing storageclass")
		return nil, err
	}
	for _, sc := range storageClassList.Items {
		// overime hodnotu annotace storage-operator.cfy.cz/storage-type
		// Pokud pro dany annotation najdeme storage class vratime cely objekt, jinak vratime nil
		annotations := sc.GetAnnotations()
		if value, exists := annotations["storage-operator.cfy.cz/storage-type"]; exists && value == v {
			return &sc, nil
		}
	}
	return nil, nil
}
