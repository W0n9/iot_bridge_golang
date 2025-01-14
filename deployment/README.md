# Deploy

1. Create the iot namespace
```
kubectl create namespace iot
```

2. Create the ConfigMap from local file
```
kubectl create configmap iot-config -n iot --from-file=config.yaml
```

3. deploy the iot service
```
kubectl apply -f workload.yaml
kubectl apply -f service.yaml
```