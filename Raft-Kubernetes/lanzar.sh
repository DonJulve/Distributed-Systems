export PATH=$PATH:$(pwd) # Lo ponemos en el path

echo "Cerrando Cluster"
./cerrar_cluster.sh

echo "Creando Cluster"
./kind-with-registry.sh

echo -e "\nCompilando servidor"
rm  ./DockerFiles/servidor/srvraft >/dev/null 2>&1
cd ./raft
CGO_ENABLED=0 go build -o ./../DockerFiles/servidor/srvraft ./cmd/srvraft/main.go
cd ./../DockerFiles/servidor
docker build . -t localhost:5001/srvraft:latest
docker push localhost:5001/srvraft:latest

cd ./../..

echo -e "\nCompilando cliente"
rm  ./DockerFiles/cliente/cltraft >/dev/null 2>&1
cd ./raft
CGO_ENABLED=0 go build -o ./../DockerFiles/cliente/cltraft ./pkg/cltraft/cltraft.go
cd ./../DockerFiles/cliente
docker build . -t localhost:5001/cltraft:latest
docker push localhost:5001/cltraft:latest

cd ./../..

echo -e "\nCompilando barrera"
rm  ./DockerFiles/barrier/barrier >/dev/null 2>&1
cd ./barrier
CGO_ENABLED=0 go build -o ./../DockerFiles/barrier/barrier barrier.go
cd ./../DockerFiles/barrier
docker build . -t localhost:5001/barrier:latest
docker push localhost:5001/barrier:latest

cd ./../..

./go_pods.sh


