echo "etcd init pulsar conf......"

export PROJECT_NAME=pulsar
export ETCDCTL_API=3
export ETCDCTL_ENDPOINTS=http://etcd:2380

echo "start put key to ${PROJECT_NAME}"

etcdctl put /${PROJECT_NAME}/data/mongo/uri "mongodb://mongodb:27017/"
etcdctl put /${PROJECT_NAME}/data/redis '{"addr":"redis:6379", "db":0, "password":"","userName":""}'
etcdctl put /${PROJECT_NAME}/data/jwt/secret "test"
etcdctl put /${PROJECT_NAME}/data/nats '{"addr":"127.0.0.1","port":4222}'

echo "finish etcd-init"
