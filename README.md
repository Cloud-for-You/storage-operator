# storage-operator
Zjednodušuje vytváření vazby PersistentVolumeClaim a PersistentVolume v k8s clusteru. Pro vytvoření vazby postačí definovat CustomResource, který obsahuje popis, co se má v daném namespace připojit. Aktu8ln2 je podporov8no pouze NFS.

## Container Environment
Environment | Default | Popis
---|---|---
CHECK_EXPORTPATH | false | Zapne kontrolu exportu na NFS serveru. Pokud není export dostupný, zástane stav ve stavu Pending a bude zařazen do rekoncilační fronty pro další rekoncilaci.

### Example CustomResource
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