kubectl delete pod cliente
kubectl delete service raft-service
kubectl delete statefulset raft
kubectl delete statefulset barrier
echo "--------- Esperar un poco para dar tiempo que terminen Pods previos"
sleep 1
kubectl create -f pods_go.yaml
