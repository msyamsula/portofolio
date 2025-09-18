# 1. Save postgres into a tarball
docker save redis -o redis.tar

# 2. Copy into every kind node
for node in $(kind get nodes --name my-cluster); do
  echo ">> Loading image into $node"
  docker cp redis.tar $node:/redis.tar
  docker exec $node ctr --namespace=k8s.io images import /redis.tar
done

# 3. Cleanup (optional)
rm redis.tar
