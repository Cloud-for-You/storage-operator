package setup

import (
	"context"
	"os"

	v1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

var (
	setupLog         = ctrl.Log.WithName("setup")
	StorageClassName string
)

func StorageClass() {
	// Client config
	cfg, err := config.GetConfig()
	if err != nil {
		setupLog.Error(err, "Error loading client configuration.")
	}
	// Create client
	c, err := client.New(cfg, client.Options{})
	if err != nil {
		setupLog.Error(err, "Error creating the client.")
	}
	// Check/Create StorageClass
	StorageClassName = os.Getenv("STORAGE_CLASS_NAME")
	if StorageClassName == "" {
		StorageClassName = "storage-operator"
	}

	provisionerName := "csi.storage.cfy.cz"
	parameters := map[string]string{
		"type": "generic",
	}
	sc := &v1.StorageClass{
		ObjectMeta: metav1.ObjectMeta{
			Name: StorageClassName,
		},
		Provisioner: provisionerName,
		Parameters:  parameters,
	}

	err = c.Get(context.TODO(), client.ObjectKey{Name: StorageClassName}, &v1.StorageClass{})
	if err != nil {
		setupLog.Info("StorageClass " + StorageClassName + " does not exist and will be created.")
		err := c.Create(context.TODO(), sc)
		if err != nil {
			setupLog.Error(err, "Failed to create new StorageClass", "SC.Name", StorageClassName)
			os.Exit(1)
		}
	}
}

func GetStorageClass() string {
	return StorageClassName
}
