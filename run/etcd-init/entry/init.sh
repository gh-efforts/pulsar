echo "etcd init pulsar conf......"

export PROJECT_NAME=pulsar
export ETCDCTL_API=3
export ETCDCTL_ENDPOINTS=http://etcd:2380

echo "start put key to ${PROJECT_NAME}"

etcdctl put /${PROJECT_NAME}/data/mongo/uri "mongodb://mongodb:27017/"
etcdctl put /${PROJECT_NAME}/data/redis '{"addr":"redis:6379", "db":0, "password":"","userName":""}'

echo "finish etcd-init"
