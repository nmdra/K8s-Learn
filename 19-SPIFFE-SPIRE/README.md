# SPIFFE / SPIRE on Kubernetes

> Learn workload identity, mTLS between services, and how to operate a SPIRE deployment in a real cluster.

## What Is SPIFFE/SPIRE?

**SPIFFE** is an open standard (part of CNCF) for securely identifying workloads in dynamic, heterogeneous environments. Instead of relying on network location (IP addresses) or secrets passed at runtime, SPIFFE assigns each workload a cryptographically verifiable identity called a **SVID** (SPIFFE Verifiable Identity Document).

**SPIRE** is the reference implementation of SPIFFE. It acts as an identity control plane that:

- Attests workloads using platform-specific plugins (e.g., Kubernetes node/pod selectors)
- Issues X.509 SVIDs and JWT SVIDs
- Rotates certificates automatically
- Federates across trust domains

### Core Concepts

| Concept              | Description                                                                                        |
| -------------------- | -------------------------------------------------------------------------------------------------- |
| **Trust Domain**     | A logical grouping of workloads under one SPIRE Server (e.g., `spiffe://example.org`)              |
| **SPIFFE ID**        | A URI that uniquely identifies a workload, e.g., `spiffe://example.org/ns/default/sa/frontend`     |
| **SVID**             | A short-lived X.509 certificate or JWT token that carries the SPIFFE ID                            |
| **SPIRE Server**     | Central authority that issues SVIDs and manages registration entries                                |
| **SPIRE Agent**      | Node-level daemon that attests workloads and delivers SVIDs via the Workload API                   |
| **Workload API**     | A Unix socket API (gRPC) that workloads use to fetch their SVIDs                                   |
| **Attestation**      | The process of verifying that a workload is who it claims to be                                    |
| **Registration Entry** | A policy rule on the SPIRE Server mapping selectors to a SPIFFE ID                              |

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│                   Kubernetes Cluster                    │
│                                                         │
│  ┌─────────────────────────────────────────────────┐   │
│  │              SPIRE Server (StatefulSet)          │   │
│  │  - CA / Root of Trust                           │   │
│  │  - Registration entries datastore               │   │
│  │  - Issues X.509 SVIDs to Agents                 │   │
│  └──────────────────┬──────────────────────────────┘   │
│                     │ mTLS (node attestation)           │
│     ┌───────────────┼───────────────┐                  │
│     │               │               │                   │
│  ┌──┴──┐         ┌──┴──┐         ┌──┴──┐              │
│  │Agent│         │Agent│         │Agent│  (DaemonSet)  │
│  └──┬──┘         └──┬──┘         └──┬──┘              │
│     │               │               │                   │
│  Workload         Workload        Workload              │
│  API socket       API socket      API socket            │
│     │               │                                   │
│  ┌──┴──────┐   ┌────┴────┐                             │
│  │ Pod A   │   │  Pod B  │  ← fetch SVIDs via socket   │
│  │frontend │   │backend  │                              │
│  └─────────┘   └─────────┘                             │
│                                                         │
└─────────────────────────────────────────────────────────┘
```

## Project Structure

```
19-SPIFFE-SPIRE/
├── README.md
├── spire-infra/                  # Phase 2 — SPIRE Server & Agent
│   ├── spire-namespace.yaml
│   ├── server-account.yaml
│   ├── server-cluster-role.yaml
│   ├── server-configmap.yaml
│   ├── server-statefulset.yaml
│   ├── server-service.yaml
│   ├── spiffe-csi-driver.yaml
│   ├── agent-account.yaml
│   ├── agent-cluster-role.yaml
│   ├── agent-configmap.yaml
│   └── agent-daemonset.yaml
├── workloads/                    # Phase 3 & 4 — Sample workloads
│   ├── frontend-deployment.yaml
│   ├── backend-deployment.yaml
│   ├── frontend-service.yaml
│   └── backend-service.yaml
├── registration/                 # Phase 3 — Registration entries
│   └── register-entries.sh
├── sample-app/                   # Phase 4 — Go mTLS example
│   ├── go.mod
│   ├── frontend/
│   │   └── main.go
│   └── backend/
│       └── main.go
└── advanced/                     # Phase 5 — Advanced topics
    ├── cluster-spiffeid.yaml
    └── federation.sh
```

## Deployment Order

> [!IMPORTANT]
> Do NOT apply these manifests until instructed. This section is for reference.

```bash
# Phase 2 — SPIRE Infrastructure
kubectl apply -f spire-infra/spire-namespace.yaml
kubectl apply -f spire-infra/server-account.yaml
kubectl apply -f spire-infra/spiffe-csi-driver.yaml
kubectl apply -f spire-infra/server-cluster-role.yaml
kubectl apply -f spire-infra/server-configmap.yaml
kubectl apply -f spire-infra/server-statefulset.yaml
kubectl apply -f spire-infra/server-service.yaml
kubectl apply -f spire-infra/agent-account.yaml
kubectl apply -f spire-infra/agent-cluster-role.yaml
kubectl apply -f spire-infra/agent-configmap.yaml
kubectl apply -f spire-infra/agent-daemonset.yaml

# Verify
kubectl get pods -n spire

# Phase 3 — Register workloads
chmod +x registration/register-entries.sh
./registration/register-entries.sh

# Phase 4 — Deploy sample workloads
kubectl apply -f workloads/
```

## Quick Reference: Common Commands

```bash
# List all SPIRE agents
kubectl exec -n spire spire-server-0 -- \
  /opt/spire/bin/spire-server agent list

# List all registration entries
kubectl exec -n spire spire-server-0 -- \
  /opt/spire/bin/spire-server entry show

# Create an entry
kubectl exec -n spire spire-server-0 -- \
  /opt/spire/bin/spire-server entry create \
  -spiffeID spiffe://example.org/myservice \
  -parentID spiffe://example.org/spire/agent/k8s_psat/cluster/node-name \
  -selector k8s:ns:default \
  -selector k8s:sa:myservice

# Delete an entry by ID
kubectl exec -n spire spire-server-0 -- \
  /opt/spire/bin/spire-server entry delete -entryID <id>

# Fetch X.509 SVID (from inside a workload or agent pod)
/opt/spire/bin/spire-agent api fetch x509 \
  -socketPath /run/spire/sockets/agent.sock \
  -write /tmp/

# Fetch JWT SVID
/opt/spire/bin/spire-agent api fetch jwt \
  -socketPath /run/spire/sockets/agent.sock \
  -audience my-api

# Get server bundle (root CA)
kubectl exec -n spire spire-server-0 -- \
  /opt/spire/bin/spire-server bundle show

# Check server health
kubectl exec -n spire spire-server-0 -- \
  /opt/spire/bin/spire-server healthcheck -shallow
```

## Troubleshooting

| Symptom                              | Likely Cause                                  | Fix                                                                 |
| ------------------------------------ | --------------------------------------------- | ------------------------------------------------------------------- |
| Agent not attesting                  | PSAT token not projected into agent pod       | Check `serviceAccountToken` volume in agent DaemonSet               |
| Workload gets no SVID                | No matching registration entry                | Run `entry show` and verify selectors match pod metadata            |
| Certificate expired immediately      | Clock skew between nodes                      | Sync NTP; check `default_x509_svid_ttl` config                     |
| Agent can't reach server             | Network policy blocking port 8081             | Allow agent → server on port 8081 in NetworkPolicy                  |
| mTLS handshake fails                 | Trust bundle mismatch                         | Run `bundle show` on both sides; refresh if federation is stale     |
| `no identity issued` from Workload API | Workload socket path wrong                  | Confirm CSI driver or hostPath mount matches `agent.conf` socket path |

## Learning Resources

| Resource                           | Type             | Link                                                       |
| ---------------------------------- | ---------------- | ---------------------------------------------------------- |
| SPIFFE Specification               | RFC-style docs   | https://github.com/spiffe/spiffe                           |
| SPIRE Documentation                | Official docs    | https://spiffe.io/docs/latest/spire-about/                 |
| SPIRE Tutorials Repo               | Hands-on examples| https://github.com/spiffe/spire-tutorials                  |
| go-spiffe Library                  | Go SDK           | https://github.com/spiffe/go-spiffe                        |
| SPIFFE/SPIRE Book (O'Reilly)       | In-depth book    | "Solving the Bottom Turtle" (free PDF at spiffe.io)        |
| CNCF Slack `#spiffe`               | Community        | https://slack.cncf.io                                      |
| SPIRE Controller Manager           | Auto-registration| https://github.com/spiffe/spire-controller-manager         |
