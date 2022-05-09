if ! [ -x "$(command -v etcdctl)" ];  then
  echo 'error:etcdctl is not installed.' >&2
  exit 1
fi


. /var/etcd/entry/init.sh