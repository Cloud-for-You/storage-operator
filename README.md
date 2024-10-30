# storage-operator
Zjednodušuje vytváření vazby PersistentVolumeClaim a PersistentVolume v k8s clusteru. Pro vytvoření vazby postačí definovat CustomResource, který obsahuje popis, co se má v daném namespace připojit. Aktuálně je podporováno pouze NFS.

## Container Environment
Environment | Default | Vyžadováno | Popis
--- | --- | --- | ---
CHECK_EXPORTPATH | false | Ano | Zapne kontrolu exportu na NFS serveru. Pokud není export dostupný, zástane stav ve stavu Pending a bude zařazen do rekoncilační fronty pro další rekoncilaci.
CLUSTER_NAME | | Ne | Nutné nastavit, pokud vyžadujeme zapnutou automatizaci.
AWX_URL | | Ne | Pouze pokud je ve storageClass zapnuta automatizace pomocí AWX. 
AWX_USERNAME | | Ne | Pouze pokud je ve storageClass zapnuta automatizace pomocí AWX.
AWX_PASSWORD | | Ne | Pouze  pokud je ve storageClass zapnuta automatizace pomocí AWX.

## StorageClass
Pro správné fungování operátor vyžaduje v clusteru storageClass, která musí obsahovat annotaci ***storage-operator.cfy.cz/storage-type: nfs***. Pro automatizované provisionování nfs exportu je možné povolit automatizaci parametrem provisioner. Aktuálně je podporována pouze jediná automatizace a to provolání RestAPI Ansible Toweru a spuštění existující template. Tato template je uvedena ve storageClass jako ***parameters.job-template-id*** a dále je nutné specifikovat jméno endpointu pro provisionování PVC ***parameters.hosts***. Tento parametr se dále přenáší jako hosts pro spouštěný Ansible Playbook.

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
  hosts: "ansible_hosts"
  job-template-id: "105"
  volume-group: vg_nfs
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

## AWX Automation
Je možné vytvářet persistentní úložiště pomocí AWX Automation platform. AWX má automatizaci v rámci ansible playbooku. Parametry, které jsou možné pro toto nastavit jsou.


Jméno parametru | Popis | Vyžadováno
--- | --- | ---
hosts | Jména hostů nebo groups, na které bude playbook delegován | Ano
job-template-id | ID template, pod kterou je v AWX automatizace vedena | Ano
volume-group | Jméno volume group (LVM), na kterou je persistentní svazek umístěn | Ne

Pokud není parametr vyžadován je nutné jej specifikovat jako default parametr v AWX