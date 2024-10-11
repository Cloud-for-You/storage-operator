# storage-operator
Zjednodušuje vytváření vazby PersistentVolumeClaim a PersistentVolume v k8s clusteru. Pro vytvoření vazby postačí definovat CustomResource, který obsahuje popis, co se má v daném namespace připojit. Aktu8ln2 je podporov8no pouze NFS.

## Container Environment
Environment | Default | Popis
---|---|---
CHECK_EXPORTPATH | false | Zapne kontrolu exportu na NFS serveru. Pokud není export dostupný, zástane stav ve stavu Pending a bude zařazen do rekoncilační fronty pro další rekoncilaci.
AWX_URL | | (optional) Pouze pokud je ve storageClass zapnuta automatizace pomocí AWX. 
AWX_USERNAME | | (optional) Pouze pokud je ve storageClass zapnuta automatizace pomocí AWX.
AWX_PASSWORD | | (optional) Pouze  pokud je ve storageClass zapnuta automatizace pomocí AWX.

## StorageClass
Pro správné fungování operátor vyžaduje v clusteru storageClass, která musí obsahovat annotaci ***storage-operator.cfy.cz/storage-type: nfs***. Pro automatizované provisionování nfs exportu je možné povolit automatizaci parametrem provisioner. Aktuálně je podporována pouze jediná automatizace a to provolání RestAPI Ansible Toweru a spuštění existující template. Tato template je uvedena ve storageClass jako ***parameters.job_template_id*** a dále je nutné specifikovat jméno endpointu pro provisionování PVC ***parameters.hosts***. Tento parametr se dále přenáší jako hosts pro spouštěný Ansible Playbook.

```yaml
allowVolumeExpansion: true
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  annotations:
    storage-operator.cfy.cz/storage-type: nfs
  name: nfs
mountOptions:
- nfsvers=4
- hard
- intr
provisioner: storage-operator.cfy.cz/awx
parameters:
  job_template_id: "105"
  hosts: "ansible_hosts"
reclaimPolicy: Retain
volumeBindingMode: Immediate
```

### Example CustomResource
Custom resource pro objekt, který chceme mít přístupný jako PersistentVolumeClaim a vytvořený claim s PersistentVolume
```yaml
apiVersion: storage.cfy.cz/v1
kind: Nfs
metadata:
  name: example-nfs
  namespace: example
spec:
  server: 172.16.4.102
  path: /volume1/nfs-exports
  capacity: 1Gi
```