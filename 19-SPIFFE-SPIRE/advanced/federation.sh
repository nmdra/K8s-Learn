#!/bin/bash
# Federation example — configure SPIRE Server A to trust Server B
# Run this from inside the SPIRE Server pod or via kubectl exec

set -euo pipefail

# Replace these values with your actual federation targets
REMOTE_BUNDLE_URL="https://spire-server-b:8443"
REMOTE_TRUST_DOMAIN="cluster-b.example.org"
REMOTE_SPIFFE_ID="spiffe://cluster-b.example.org/spire/server"

echo "==> Creating federation with ${REMOTE_TRUST_DOMAIN}..."
/opt/spire/bin/spire-server federation create \
  -bundleEndpointURL "${REMOTE_BUNDLE_URL}" \
  -bundleEndpointProfile https_spiffe \
  -endpointSPIFFEID "${REMOTE_SPIFFE_ID}" \
  -trustDomainID "spiffe://${REMOTE_TRUST_DOMAIN}"

echo "==> Listing federations..."
/opt/spire/bin/spire-server federation list

echo "==> Done!"
