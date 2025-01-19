# ETCD

> https://etcd.io/docs/v3.5/
> https://etcd.io/docs/v2.3/docker_guide/
> **Etcd Playgroud:** http://play.etcd.io/play
> https://raft.github.io/


## Create ETCD cluster

```bash
docker-compose up -d
```

## Interact with etcd cluster

1. Get Version

```bash
docker exec etcd1 etcdctl version
```

2. List Members

```bash
etcdctl --endpoints=http://etcd1:2379 member list
```

3. Check Leader

```bash
docker exec etcd1 etcdctl --endpoints=http://etcd1:2379,http://etcd2:2379,http://etcd3:2379 endpoint status --write-out=table
```

4. Watch Value

```bash
docker exec etcd1 etcdctl --endpoints=http://etcd1:2379 watch --prefix foo
```

5. PUT Value

```bash
docker exec etcd1 etcdctl --endpoints=http://etcd1:2379 put bar "Hello etcd
```



