# 1. Save postgres into a tarball
docker save postgres:16.3 -o postgres.tar

# 2. Copy into every kind node
for node in $(kind get nodes --name my-cluster); do
  echo ">> Loading image into $node"
  docker cp postgres.tar $node:/postgres.tar
  docker exec $node ctr --namespace=k8s.io images import /postgres.tar
done

# 3. Cleanup (optional)
rm postgres.tar
