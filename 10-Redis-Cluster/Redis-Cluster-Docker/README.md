## Start Docker Containers

```bash
docker-compose up -d
```
## Create the Redis Cluster

```bash
docker exec -it redis-node-1 redis-cli --cluster create \
  redis-node-1:6379 redis-node-2:6379 redis-node-3:6379 \
  redis-node-4:6379 redis-node-5:6379 redis-node-6:6379 \
  --cluster-replicas 1
```

## Check Cluster Info

```bash
docker exec -it redis-node-1 redis-cli -c -p 6379 cluster info
```

## Test Redis Cluster

```bash
docker exec -it redis-node-1 redis-cli -c -p 6379 set foo bar

docker exec -it redis-node-4 redis-cli -c -p 6379 get foo
```

