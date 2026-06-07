#!/bin/bash
# Register SPIRE workload entries
# Run this AFTER the SPIRE server is running and agents are attested.

set -euo pipefail

SPIRE_SERVER_POD="spire-server-0"
SPIRE_NAMESPACE="spire"
TRUST_DOMAIN="example.org"

echo "==> Registering frontend workload..."
kubectl exec -n "${SPIRE_NAMESPACE}" "${SPIRE_SERVER_POD}" -- \
  /opt/spire/bin/spire-server entry create \
  -spiffeID "spiffe://${TRUST_DOMAIN}/ns/default/sa/frontend" \
  -parentID "spiffe://${TRUST_DOMAIN}/spire/agent/k8s_psat/demo-cluster" \
  -selector k8s:ns:default \
  -selector k8s:sa:frontend

echo "==> Registering backend workload..."
kubectl exec -n "${SPIRE_NAMESPACE}" "${SPIRE_SERVER_POD}" -- \
  /opt/spire/bin/spire-server entry create \
  -spiffeID "spiffe://${TRUST_DOMAIN}/ns/default/sa/backend" \
  -parentID "spiffe://${TRUST_DOMAIN}/spire/agent/k8s_psat/demo-cluster" \
  -selector k8s:ns:default \
  -selector k8s:sa:backend

echo "==> Listing all entries..."
kubectl exec -n "${SPIRE_NAMESPACE}" "${SPIRE_SERVER_POD}" -- \
  /opt/spire/bin/spire-server entry show

echo "==> Done!"
