#!/usr/bin/env bash
set -euo pipefail

PRODUCT_ID="${PRODUCT_ID:-prod-123}"
STOCK="${STOCK:-100}"
NAMESPACE="${NAMESPACE:-default}"

REDIS_POD="$(
  kubectl -n "$NAMESPACE" get pods -l app=redis \
    -o jsonpath='{.items[0].metadata.name}'
)"

KEY="flashsale:stock:${PRODUCT_ID}"

echo "===== INIT FLASH SALE STOCK ====="
echo "namespace=$NAMESPACE"
echo "redis_pod=$REDIS_POD"
echo "product_id=$PRODUCT_ID"
echo "stock=$STOCK"
echo "key=$KEY"

kubectl -n "$NAMESPACE" exec "$REDIS_POD" -- redis-cli SET "$KEY" "$STOCK"
kubectl -n "$NAMESPACE" exec "$REDIS_POD" -- redis-cli GET "$KEY"
