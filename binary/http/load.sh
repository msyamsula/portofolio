# 1. Save http into a tarball
docker save syamsuldocker:0.0.0 -o syamsuldocker:0.0.0.tar

# 2. Copy into every kind node
for node in $(kind get nodes --name my-cluster); do
  echo ">> Loading image into $node"
  docker cp syamsuldocker:0.0.0.tar $node:/syamsuldocker:0.0.0.tar
  docker exec $node ctr --namespace=k8s.io images import /syamsuldocker:0.0.0.tar
done

# 3. Cleanup (optional)
rm syamsuldocker:0.0.0.tar
