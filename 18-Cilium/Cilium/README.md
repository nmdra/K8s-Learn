# 🐝 Cilium — eBPF-Powered Kubernetes Networking, Observability & Security

> A project-based guide to learn [Cilium](https://cilium.io/) by building and securing a real-world microservices application on Kubernetes.

---

## Table of Contents

- [What is Cilium?](#what-is-cilium)
- [Key Features](#key-features)
- [Demo Project — Online Bookstore](#demo-project--online-bookstore)
  - [Architecture](#architecture)
  - [Project Structure](#project-structure)
- [Part 1 — Cluster Setup & Cilium Installation](#part-1--cluster-setup--cilium-installation)
  - [Prerequisites](#prerequisites)
  - [Install Cilium CLI & Hubble CLI](#install-cilium-cli--hubble-cli)
  - [Create the Kind Cluster](#create-the-kind-cluster)
  - [Install Cilium with Helm](#install-cilium-with-helm)
  - [Verify Installation](#verify-installation)
- [Part 2 — Deploy the Bookstore Application](#part-2--deploy-the-bookstore-application)
  - [Step 1: Create the Namespace](#step-1-create-the-namespace)
  - [Step 2: Deploy All Microservices](#step-2-deploy-all-microservices)
  - [Step 3: Verify Deployment](#step-3-verify-deployment)
  - [Step 4: Test Connectivity (No Policies)](#step-4-test-connectivity-no-policies)
- [Part 3 — Network Policies (Zero Trust)](#part-3--network-policies-zero-trust)
  - [Step 1: Default Deny All Traffic](#step-1-default-deny-all-traffic)
  - [Step 2: Allow Frontend → API Gateway](#step-2-allow-frontend--api-gateway)
  - [Step 3: Allow API Gateway → Backend Services](#step-3-allow-api-gateway--backend-services)
  - [Step 4: L7 HTTP Policy (Read-Only Catalog)](#step-4-l7-http-policy-read-only-catalog)
  - [Step 5: DNS Egress Policy (External APIs Only)](#step-5-dns-egress-policy-external-apis-only)
  - [Step 6: Test the Policies](#step-6-test-the-policies)
  - [CiliumNetworkPolicy vs NetworkPolicy](#ciliumnetworkpolicy-vs-networkpolicy)
- [Part 4 — Observability with Hubble](#part-4--observability-with-hubble)
  - [Observe Live Traffic Flows](#observe-live-traffic-flows)
  - [Investigate Dropped Traffic](#investigate-dropped-traffic)
  - [Hubble UI — Service Map](#hubble-ui--service-map)
  - [Hubble Metrics with Prometheus](#hubble-metrics-with-prometheus)
- [Part 5 — kube-proxy Replacement & Load Balancing](#part-5--kube-proxy-replacement--load-balancing)
  - [Verify kube-proxy Replacement](#verify-kube-proxy-replacement)
  - [Inspect Cilium Service Table](#inspect-cilium-service-table)
  - [Maglev Consistent Hashing](#maglev-consistent-hashing)
- [Part 6 — Transparent Encryption (WireGuard)](#part-6--transparent-encryption-wireguard)
- [Part 7 — Ingress & Gateway API](#part-7--ingress--gateway-api)
- [Part 8 — Bandwidth Management](#part-8--bandwidth-management)
- [Part 9 — Cluster Mesh (Multi-Cluster)](#part-9--cluster-mesh-multi-cluster)
- [Troubleshooting & Debugging](#troubleshooting--debugging)
- [Resources](#resources)
- [Sub-Projects](#sub-projects)

---

## What is Cilium?

Cilium is an open-source project that provides **eBPF-based networking, security, and observability** for cloud-native environments. It operates at the Linux kernel level using eBPF programs, bypassing traditional iptables for significantly better performance and programmability.

```
┌──────────────────────────────────────────┐
│              Kubernetes Cluster          │
│                                          │
│  ┌──────────┐  ┌──────────┐              │
│  │  Pod A   │  │  Pod B   │  ← Endpoints │
│  └────┬─────┘  └────┬─────┘              │
│       │              │                   │
│  ┌────▼──────────────▼─────┐             │
│  │     Cilium Agent (eBPF) │  ← Datapath │
│  └─────────────────────────┘             │
│       │                                  │
│  ┌────▼────────────────────┐             │
│  │   Linux Kernel (eBPF)   │             │
│  └─────────────────────────┘             │
└──────────────────────────────────────────┘
```

## Key Features

| Feature                    | Description                                                       |
| :------------------------- | :---------------------------------------------------------------- |
| **eBPF Datapath**          | High-performance networking directly in the Linux kernel           |
| **kube-proxy Replacement** | Native eBPF-based service load balancing (no iptables)            |
| **Network Policies**       | L3/L4/L7 policies with identity-based enforcement                |
| **Hubble**                 | Deep observability — flow logs, service maps, metrics             |
| **Service Mesh**           | Sidecar-free service mesh using eBPF (mTLS, traffic management)  |
| **Cluster Mesh**           | Multi-cluster connectivity with global services                   |
| **Bandwidth Manager**      | EDT-based rate limiting for fair bandwidth allocation             |
| **Transparent Encryption** | IPsec or WireGuard-based node-to-node encryption                 |
| **BGP Control Plane**      | Native BGP speaker for advertising Pod CIDRs & service IPs       |
| **Tetragon**               | Runtime security & observability based on eBPF                   |

---

## Demo Project — Online Bookstore

We will learn every Cilium feature by building and progressively securing a realistic **online bookstore** microservices application.

### Architecture

```
                        ┌──────────────────────────┐
                        │        Internet          │
                        └────────────┬─────────────┘
                                     │
                          ┌──────────▼──────────┐
                          │   Cilium Ingress /   │
                          │     Gateway API      │
                          └──────────┬───────────┘
                                     │
                  ┌──────────────────▼──────────────────┐
                  │          API Gateway (:8080)         │
                  │       (routes & auth proxy)          │
                  └───┬──────────┬──────────┬───────────┘
                      │          │          │
           ┌──────────▼──┐ ┌────▼─────┐ ┌──▼───────────┐
           │  Catalog     │ │  Orders  │ │  Payments    │
           │  Service     │ │  Service │ │  Service     │
           │  (:5000)     │ │  (:5001) │ │  (:5002)     │
           └──────┬───────┘ └────┬─────┘ └──┬───────────┘
                  │              │           │
           ┌──────▼───────┐ ┌───▼──────┐   │  ┌───────────────┐
           │  Catalog DB  │ │ Orders DB│   └─▶│ Stripe API    │
           │  (postgres)  │ │ (postgres│      │ (external)    │
           └──────────────┘ └──────────┘      └───────────────┘
```

**Scenario:** The bookstore must enforce:
1. **Zero-trust networking** — deny all traffic by default, allow only explicitly
2. **L7 policy** — the Catalog API is read-only for public users (`GET` only)
3. **DNS egress control** — only the Payments service can reach `api.stripe.com`
4. **Encryption** — all node-to-node traffic encrypted via WireGuard
5. **Observability** — full flow visibility and service dependency map via Hubble
6. **Rate limiting** — bandwidth caps on bulk-export endpoints

### Project Structure

```
18-Cilium/
├── README.md
├── cluster/
│   └── kind-config.yaml              # Kind cluster with Cilium-ready config
├── app/
│   ├── namespace.yaml                 # bookstore namespace
│   ├── catalog-service.yaml           # Deployment + Service
│   ├── orders-service.yaml
│   ├── payments-service.yaml
│   ├── api-gateway.yaml
│   └── databases.yaml                 # PostgreSQL StatefulSets
├── policies/
│   ├── 00-default-deny.yaml           # Default deny all
│   ├── 01-allow-frontend-to-gw.yaml   # Frontend → API Gateway
│   ├── 02-allow-gw-to-backends.yaml   # API Gateway → microservices
│   ├── 03-allow-backends-to-db.yaml   # Services → their own DB
│   ├── 04-l7-catalog-readonly.yaml    # L7: GET-only on Catalog
│   ├── 05-dns-egress-payments.yaml    # DNS: Payments → stripe.com
│   └── 06-deny-cross-service.yaml     # Prevent catalog ↔ orders directly
├── ingress/
│   └── gateway.yaml                   # Cilium Gateway API resources
├── encryption/
│   └── wireguard-values.yaml          # Helm values for WireGuard
├── bandwidth/
│   └── rate-limit-annotation.yaml     # Bandwidth manager annotations
└── ebpf-go-basics/                    # eBPF sub-project
    └── README.md
```

---

## Part 1 — Cluster Setup & Cilium Installation

### Prerequisites

- Docker Desktop or Docker Engine
- [Kind](https://kind.sigs.k8s.io/) v0.20+
- `kubectl` configured
- `helm` v3+
- Linux kernel ≥ 5.4 (5.10+ recommended)

### Install Cilium CLI & Hubble CLI

```bash
# Cilium CLI
CILIUM_CLI_VERSION=$(curl -s https://raw.githubusercontent.com/cilium/cilium-cli/main/stable.txt)
curl -L --fail --remote-name-all \
  https://github.com/cilium/cilium-cli/releases/download/${CILIUM_CLI_VERSION}/cilium-linux-amd64.tar.gz
sudo tar xzvfC cilium-linux-amd64.tar.gz /usr/local/bin
rm cilium-linux-amd64.tar.gz

# Hubble CLI
HUBBLE_VERSION=$(curl -s https://raw.githubusercontent.com/cilium/hubble/master/stable.txt)
curl -L --fail --remote-name-all \
  https://github.com/cilium/hubble/releases/download/${HUBBLE_VERSION}/hubble-linux-amd64.tar.gz
sudo tar xzvfC hubble-linux-amd64.tar.gz /usr/local/bin
rm hubble-linux-amd64.tar.gz
```

### Create the Kind Cluster

```yaml
# cluster/kind-config.yaml
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
networking:
  disableDefaultCNI: true    # Let Cilium manage CNI
  kubeProxyMode: none         # Cilium replaces kube-proxy
nodes:
  - role: control-plane
    extraPortMappings:
      - containerPort: 80     # For Cilium Ingress
        hostPort: 80
      - containerPort: 443
        hostPort: 443
  - role: worker
  - role: worker
```

```bash
kind create cluster --config cluster/kind-config.yaml --name cilium-bookstore
```

### Install Cilium with Helm

```bash
helm repo add cilium https://helm.cilium.io/
helm repo update

helm install cilium cilium/cilium --version 1.17.0 \
  --namespace kube-system \
  --set kubeProxyReplacement=true \
  --set hubble.relay.enabled=true \
  --set hubble.ui.enabled=true \
  --set hubble.metrics.enableOpenMetrics=true \
  --set hubble.metrics.enabled="{dns,drop,tcp,flow,port-distribution,icmp,httpV2:exemplars=true;labelsContext=source_ip\,source_namespace\,source_workload\,destination_ip\,destination_namespace\,destination_workload\,traffic_direction}" \
  --set ingressController.enabled=true \
  --set ingressController.loadbalancerMode=shared \
  --set encryption.enabled=true \
  --set encryption.type=wireguard
```

### Verify Installation

```bash
# Wait for all Cilium pods to be ready
cilium status --wait

# Quick connectivity validation
cilium connectivity test

# Check component versions
kubectl -n kube-system exec ds/cilium -- cilium version
```

---

## Part 2 — Deploy the Bookstore Application

### Step 1: Create the Namespace

```yaml
# app/namespace.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: bookstore
  labels:
    project: bookstore
```

```bash
kubectl apply -f app/namespace.yaml
```

### Step 2: Deploy All Microservices

```yaml
# app/catalog-service.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: catalog
  namespace: bookstore
spec:
  replicas: 2
  selector:
    matchLabels:
      app: catalog
      tier: backend
  template:
    metadata:
      labels:
        app: catalog
        tier: backend
    spec:
      containers:
        - name: catalog
          image: nginx:alpine         # Replace with real app image
          ports:
            - containerPort: 5000
          env:
            - name: SERVICE_NAME
              value: "catalog"
---
apiVersion: v1
kind: Service
metadata:
  name: catalog
  namespace: bookstore
  labels:
    app: catalog
spec:
  selector:
    app: catalog
  ports:
    - port: 5000
      targetPort: 5000
      name: http
```

```yaml
# app/orders-service.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: orders
  namespace: bookstore
spec:
  replicas: 2
  selector:
    matchLabels:
      app: orders
      tier: backend
  template:
    metadata:
      labels:
        app: orders
        tier: backend
    spec:
      containers:
        - name: orders
          image: nginx:alpine
          ports:
            - containerPort: 5001
---
apiVersion: v1
kind: Service
metadata:
  name: orders
  namespace: bookstore
  labels:
    app: orders
spec:
  selector:
    app: orders
  ports:
    - port: 5001
      targetPort: 5001
      name: http
```

```yaml
# app/payments-service.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: payments
  namespace: bookstore
spec:
  replicas: 1
  selector:
    matchLabels:
      app: payments
      tier: backend
  template:
    metadata:
      labels:
        app: payments
        tier: backend
    spec:
      containers:
        - name: payments
          image: nginx:alpine
          ports:
            - containerPort: 5002
---
apiVersion: v1
kind: Service
metadata:
  name: payments
  namespace: bookstore
  labels:
    app: payments
spec:
  selector:
    app: payments
  ports:
    - port: 5002
      targetPort: 5002
      name: http
```

```yaml
# app/api-gateway.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: api-gateway
  namespace: bookstore
spec:
  replicas: 2
  selector:
    matchLabels:
      app: api-gateway
      tier: frontend
  template:
    metadata:
      labels:
        app: api-gateway
        tier: frontend
    spec:
      containers:
        - name: gateway
          image: nginx:alpine
          ports:
            - containerPort: 8080
---
apiVersion: v1
kind: Service
metadata:
  name: api-gateway
  namespace: bookstore
  labels:
    app: api-gateway
spec:
  selector:
    app: api-gateway
  ports:
    - port: 8080
      targetPort: 8080
      name: http
```

```yaml
# app/databases.yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: catalog-db
  namespace: bookstore
spec:
  serviceName: catalog-db
  replicas: 1
  selector:
    matchLabels:
      app: catalog-db
      tier: database
  template:
    metadata:
      labels:
        app: catalog-db
        tier: database
    spec:
      containers:
        - name: postgres
          image: postgres:16-alpine
          ports:
            - containerPort: 5432
          env:
            - name: POSTGRES_DB
              value: catalog
            - name: POSTGRES_USER
              value: catalog_user
            - name: POSTGRES_PASSWORD
              value: catalog_pass     # Use Secrets in production
---
apiVersion: v1
kind: Service
metadata:
  name: catalog-db
  namespace: bookstore
spec:
  selector:
    app: catalog-db
  ports:
    - port: 5432
      targetPort: 5432
  clusterIP: None   # Headless for StatefulSet
---
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: orders-db
  namespace: bookstore
spec:
  serviceName: orders-db
  replicas: 1
  selector:
    matchLabels:
      app: orders-db
      tier: database
  template:
    metadata:
      labels:
        app: orders-db
        tier: database
    spec:
      containers:
        - name: postgres
          image: postgres:16-alpine
          ports:
            - containerPort: 5432
          env:
            - name: POSTGRES_DB
              value: orders
            - name: POSTGRES_USER
              value: orders_user
            - name: POSTGRES_PASSWORD
              value: orders_pass
---
apiVersion: v1
kind: Service
metadata:
  name: orders-db
  namespace: bookstore
spec:
  selector:
    app: orders-db
  ports:
    - port: 5432
      targetPort: 5432
  clusterIP: None
```

```bash
kubectl apply -f app/
```

### Step 3: Verify Deployment

```bash
kubectl -n bookstore get pods -o wide
kubectl -n bookstore get svc

# Check Cilium endpoints for the bookstore namespace
kubectl -n kube-system exec ds/cilium -- cilium endpoint list | grep bookstore
```

### Step 4: Test Connectivity (No Policies)

Before applying any policies, everything can talk to everything — this is the **default-allow** state:

```bash
# API Gateway → Catalog (should succeed)
kubectl -n bookstore exec deploy/api-gateway -- \
  curl -s -o /dev/null -w "%{http_code}" http://catalog:5000/

# Catalog → Orders (should succeed — but we don't want this!)
kubectl -n bookstore exec deploy/catalog -- \
  curl -s -o /dev/null -w "%{http_code}" http://orders:5001/

# Catalog → Orders DB (should succeed — DANGEROUS!)
kubectl -n bookstore exec deploy/catalog -- \
  nc -zv orders-db 5432
```

> [!WARNING]
> Without network policies, **any Pod can reach any other Pod** in the cluster — including databases. This is the problem Cilium solves.

---

## Part 3 — Network Policies (Zero Trust)

We will progressively lock down the bookstore using identity-based Cilium Network Policies.

```
                ALLOWED TRAFFIC AFTER POLICIES
                ==============================

  Internet ──▶ Ingress ──▶ API Gateway ──┬──▶ Catalog ──▶ Catalog DB
                                         ├──▶ Orders  ──▶ Orders DB
                                         └──▶ Payments ──▶ api.stripe.com

  Everything else: ❌ DENIED
```

### Step 1: Default Deny All Traffic

```yaml
# policies/00-default-deny.yaml
apiVersion: cilium.io/v2
kind: CiliumNetworkPolicy
metadata:
  name: default-deny
  namespace: bookstore
spec:
  endpointSelector: {}     # Matches ALL pods in the namespace
  ingress:
    - {}                    # This section being empty = deny all ingress
  egress:
    - {}                    # Deny all egress
```

> [!CAUTION]
> After applying this, **all traffic** in the `bookstore` namespace will be blocked. Apply the allow-rules immediately after.

```bash
kubectl apply -f policies/00-default-deny.yaml

# Verify: this should now FAIL
kubectl -n bookstore exec deploy/api-gateway -- \
  curl -s --connect-timeout 3 http://catalog:5000/
# Expected: connection timeout
```

### Step 2: Allow Frontend → API Gateway

Allow external traffic (from Cilium Ingress) to reach the API Gateway:

```yaml
# policies/01-allow-frontend-to-gw.yaml
apiVersion: cilium.io/v2
kind: CiliumNetworkPolicy
metadata:
  name: allow-ingress-to-gateway
  namespace: bookstore
spec:
  endpointSelector:
    matchLabels:
      app: api-gateway
  ingress:
    - fromEntities:
        - world               # Traffic from outside the cluster
        - cluster              # Traffic from other namespaces (e.g., ingress controller)
      toPorts:
        - ports:
            - port: "8080"
              protocol: TCP
  egress:
    - toEndpoints:
        - matchLabels:
            app: catalog
      toPorts:
        - ports:
            - port: "5000"
              protocol: TCP
    - toEndpoints:
        - matchLabels:
            app: orders
      toPorts:
        - ports:
            - port: "5001"
              protocol: TCP
    - toEndpoints:
        - matchLabels:
            app: payments
      toPorts:
        - ports:
            - port: "5002"
              protocol: TCP
    # Allow DNS resolution
    - toEndpoints:
        - matchLabels:
            k8s:io.kubernetes.pod.namespace: kube-system
            k8s-app: kube-dns
      toPorts:
        - ports:
            - port: "53"
              protocol: ANY
```

### Step 3: Allow API Gateway → Backend Services

Allow each backend service to reach only its own database:

```yaml
# policies/02-allow-gw-to-backends.yaml

# --- Catalog Service ---
apiVersion: cilium.io/v2
kind: CiliumNetworkPolicy
metadata:
  name: allow-catalog-policy
  namespace: bookstore
spec:
  endpointSelector:
    matchLabels:
      app: catalog
  ingress:
    - fromEndpoints:
        - matchLabels:
            app: api-gateway
      toPorts:
        - ports:
            - port: "5000"
              protocol: TCP
  egress:
    - toEndpoints:
        - matchLabels:
            app: catalog-db
      toPorts:
        - ports:
            - port: "5432"
              protocol: TCP
    - toEndpoints:
        - matchLabels:
            k8s:io.kubernetes.pod.namespace: kube-system
            k8s-app: kube-dns
      toPorts:
        - ports:
            - port: "53"
              protocol: ANY
---
# --- Orders Service ---
apiVersion: cilium.io/v2
kind: CiliumNetworkPolicy
metadata:
  name: allow-orders-policy
  namespace: bookstore
spec:
  endpointSelector:
    matchLabels:
      app: orders
  ingress:
    - fromEndpoints:
        - matchLabels:
            app: api-gateway
      toPorts:
        - ports:
            - port: "5001"
              protocol: TCP
  egress:
    - toEndpoints:
        - matchLabels:
            app: orders-db
      toPorts:
        - ports:
            - port: "5432"
              protocol: TCP
    - toEndpoints:
        - matchLabels:
            k8s:io.kubernetes.pod.namespace: kube-system
            k8s-app: kube-dns
      toPorts:
        - ports:
            - port: "53"
              protocol: ANY
---
# --- Database Ingress ---
apiVersion: cilium.io/v2
kind: CiliumNetworkPolicy
metadata:
  name: allow-db-ingress
  namespace: bookstore
spec:
  endpointSelector:
    matchLabels:
      tier: database
  ingress:
    - fromEndpoints:
        - matchLabels:
            tier: backend
```

### Step 4: L7 HTTP Policy (Read-Only Catalog)

This is where Cilium shines vs standard K8s NetworkPolicy. We restrict the Catalog API to **GET requests only** — preventing unauthorized writes:

```yaml
# policies/04-l7-catalog-readonly.yaml
apiVersion: cilium.io/v2
kind: CiliumNetworkPolicy
metadata:
  name: catalog-l7-readonly
  namespace: bookstore
spec:
  endpointSelector:
    matchLabels:
      app: catalog
  ingress:
    - fromEndpoints:
        - matchLabels:
            app: api-gateway
      toPorts:
        - ports:
            - port: "5000"
              protocol: TCP
          rules:
            http:
              - method: GET
                path: "/api/v1/books"
              - method: GET
                path: "/api/v1/books/[0-9]+"
              - method: GET
                path: "/health"
```

```bash
# Test: GET should succeed
kubectl -n bookstore exec deploy/api-gateway -- \
  curl -s -o /dev/null -w "%{http_code}" http://catalog:5000/api/v1/books
# Expected: 200

# Test: POST should be DENIED by Cilium L7 policy
kubectl -n bookstore exec deploy/api-gateway -- \
  curl -s -o /dev/null -w "%{http_code}" -X POST http://catalog:5000/api/v1/books
# Expected: 403 (Access denied by Cilium proxy)
```

### Step 5: DNS Egress Policy (External APIs Only)

Only the Payments service should be able to reach Stripe's API. No other service may make external calls:

```yaml
# policies/05-dns-egress-payments.yaml
apiVersion: cilium.io/v2
kind: CiliumNetworkPolicy
metadata:
  name: payments-dns-egress
  namespace: bookstore
spec:
  endpointSelector:
    matchLabels:
      app: payments
  ingress:
    - fromEndpoints:
        - matchLabels:
            app: api-gateway
      toPorts:
        - ports:
            - port: "5002"
              protocol: TCP
  egress:
    # Allow DNS resolution
    - toEndpoints:
        - matchLabels:
            k8s:io.kubernetes.pod.namespace: kube-system
            k8s-app: kube-dns
      toPorts:
        - ports:
            - port: "53"
              protocol: ANY
          rules:
            dns:
              - matchPattern: "*.stripe.com"
    # Allow HTTPS to Stripe only (resolved via DNS above)
    - toFQDNs:
        - matchPattern: "*.stripe.com"
      toPorts:
        - ports:
            - port: "443"
              protocol: TCP
```

```bash
# Test: Payments → stripe.com should succeed
kubectl -n bookstore exec deploy/payments -- \
  curl -s --connect-timeout 5 -o /dev/null -w "%{http_code}" https://api.stripe.com/v1
# Expected: 401 (Stripe rejects unauthenticated — but connection WORKS)

# Test: Payments → google.com should FAIL
kubectl -n bookstore exec deploy/payments -- \
  curl -s --connect-timeout 5 https://www.google.com
# Expected: timeout / connection refused

# Test: Catalog → anywhere external should FAIL
kubectl -n bookstore exec deploy/catalog -- \
  curl -s --connect-timeout 5 https://api.stripe.com/v1
# Expected: timeout
```

### Step 6: Test the Policies

Run the full test matrix:

```bash
echo "=== Policy Verification Matrix ==="

echo "1. API Gateway → Catalog (GET /api/v1/books) — should PASS"
kubectl -n bookstore exec deploy/api-gateway -- \
  curl -s -o /dev/null -w "%{http_code}" http://catalog:5000/api/v1/books

echo "2. API Gateway → Catalog (POST) — should FAIL (L7 deny)"
kubectl -n bookstore exec deploy/api-gateway -- \
  curl -s -o /dev/null -w "%{http_code}" -X POST http://catalog:5000/api/v1/books

echo "3. API Gateway → Orders — should PASS"
kubectl -n bookstore exec deploy/api-gateway -- \
  curl -s -o /dev/null -w "%{http_code}" http://orders:5001/

echo "4. Catalog → Orders (cross-service) — should FAIL"
kubectl -n bookstore exec deploy/catalog -- \
  curl -s --connect-timeout 3 -o /dev/null -w "%{http_code}" http://orders:5001/

echo "5. Catalog → Orders DB (cross-db) — should FAIL"
kubectl -n bookstore exec deploy/catalog -- \
  nc -zv -w 3 orders-db 5432

echo "6. Payments → stripe.com — should PASS"
kubectl -n bookstore exec deploy/payments -- \
  curl -s --connect-timeout 5 -o /dev/null -w "%{http_code}" https://api.stripe.com/v1

echo "7. Catalog → stripe.com — should FAIL"
kubectl -n bookstore exec deploy/catalog -- \
  curl -s --connect-timeout 5 -o /dev/null -w "%{http_code}" https://api.stripe.com/v1
```

### CiliumNetworkPolicy vs NetworkPolicy

| Aspect              | `NetworkPolicy` (K8s)            | `CiliumNetworkPolicy`                |
| :------------------ | :------------------------------- | :----------------------------------- |
| L3/L4 support       | ✅                                | ✅                                    |
| L7 support          | ❌                                | ✅ (HTTP, gRPC, Kafka, DNS)          |
| Identity-based      | ❌ (IP-based)                     | ✅ (label-based identity)            |
| DNS policies        | ❌                                | ✅ (FQDN-based egress)              |
| Cluster-wide scope  | ❌                                | ✅ (`CiliumClusterwideNetworkPolicy`) |
| Entity selectors    | ❌                                | ✅ (`world`, `cluster`, `host`)      |

---

## Part 4 — Observability with Hubble

With policies in place, let's use Hubble to see what's happening in real time.

### Observe Live Traffic Flows

```bash
# Port-forward Hubble Relay
cilium hubble port-forward &

# Watch ALL bookstore traffic
hubble observe --namespace bookstore -f

# Filter: only HTTP traffic
hubble observe --namespace bookstore --protocol http -f

# Filter: only traffic from api-gateway
hubble observe --namespace bookstore --from-pod bookstore/api-gateway -f

# Filter: show only specific verdicts
hubble observe --namespace bookstore --verdict FORWARDED -f
hubble observe --namespace bookstore --verdict DROPPED -f
```

### Investigate Dropped Traffic

When a policy blocks traffic, Hubble shows the exact reason:

```bash
# Show all denied flows with drop reason
hubble observe --namespace bookstore --verdict DROPPED -o json | jq '{
  src: .source.namespace + "/" + .source.pod_name,
  dst: .destination.namespace + "/" + .destination.pod_name,
  port: .l4.TCP.destination_port,
  verdict: .verdict,
  drop_reason: .drop_reason_desc
}'
```

Example output:
```json
{
  "src": "bookstore/catalog-7f8b9c-x4k2l",
  "dst": "bookstore/orders-5d6e7f-m9n3p",
  "port": 5001,
  "verdict": "DROPPED",
  "drop_reason": "Policy denied"
}
```

### Hubble UI — Service Map

```bash
# Launch Hubble UI
cilium hubble ui &
# Open http://localhost:12000 in your browser
# Select "bookstore" namespace to see the service dependency graph
```

The UI shows:
- **Green lines** → Allowed (forwarded) traffic
- **Red lines** → Denied (dropped) traffic
- **Service map** → Visual graph of which services talk to which

### Hubble Metrics with Prometheus

Hubble exports metrics that can be scraped by Prometheus:

```bash
# Check Hubble metrics endpoint
kubectl -n kube-system port-forward ds/cilium 9965:9965 &
curl -s http://localhost:9965/metrics | grep hubble

# Key metrics:
# hubble_flows_processed_total        — Total flow events
# hubble_drop_total                   — Dropped packets (by reason)
# hubble_tcp_flags_total              — TCP flag distribution
# hubble_http_requests_total          — HTTP request count (L7)
# hubble_http_request_duration_seconds — HTTP latency histogram
# hubble_dns_queries_total            — DNS query count
```

---

## Part 5 — kube-proxy Replacement & Load Balancing

Since we installed Cilium with `kubeProxyReplacement=true`, Cilium handles all Service load balancing natively via eBPF — no iptables rules.

### Verify kube-proxy Replacement

```bash
# Confirm kube-proxy is disabled
kubectl -n kube-system get pods | grep kube-proxy
# Expected: no pods (or terminated)

# Check Cilium's kube-proxy replacement status
kubectl -n kube-system exec ds/cilium -- cilium status | grep KubeProxyReplacement
# Expected: True [eth0 ...]
```

### Inspect Cilium Service Table

```bash
# List all services managed by Cilium eBPF
kubectl -n kube-system exec ds/cilium -- cilium service list

# Example output shows the bookstore services:
# ID   Frontend             Service Type   Backend
# 1    10.96.0.10:53        ClusterIP      ...
# 5    10.96.45.12:5000     ClusterIP      10.244.1.5:5000, 10.244.2.3:5000
# ...
```

### Maglev Consistent Hashing

For production workloads, enable Maglev for better load distribution:

```bash
helm upgrade cilium cilium/cilium --version 1.17.0 \
  --namespace kube-system \
  --reuse-values \
  --set loadBalancer.algorithm=maglev
```

---

## Part 6 — Transparent Encryption (WireGuard)

All pod-to-pod traffic across nodes is encrypted transparently via WireGuard (already enabled in our Helm install):

```bash
# Verify WireGuard status
kubectl -n kube-system exec ds/cilium -- cilium encrypt status
# Expected: Encryption: Wireguard ...

# Check WireGuard interfaces on nodes
kubectl -n kube-system exec ds/cilium -- wg show
```

```yaml
# encryption/wireguard-values.yaml
# Helm values to enable WireGuard (for reference)
encryption:
  enabled: true
  type: wireguard
  wireguard:
    userspaceFallback: false
```

> [!TIP]
> WireGuard encryption is **transparent** — no application changes needed. All traffic between nodes is automatically encrypted/decrypted by the kernel.

---

## Part 7 — Ingress & Gateway API

Expose the bookstore API externally using Cilium's built-in Gateway API support:

```yaml
# ingress/gateway.yaml
apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
metadata:
  name: bookstore-gateway
  namespace: bookstore
spec:
  gatewayClassName: cilium
  listeners:
    - name: http
      protocol: HTTP
      port: 80
---
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: bookstore-routes
  namespace: bookstore
spec:
  parentRefs:
    - name: bookstore-gateway
  rules:
    - matches:
        - path:
            type: PathPrefix
            value: /api/v1/books
      backendRefs:
        - name: api-gateway
          port: 8080
    - matches:
        - path:
            type: PathPrefix
            value: /api/v1/orders
      backendRefs:
        - name: api-gateway
          port: 8080
```

```bash
kubectl apply -f ingress/gateway.yaml

# Test external access
curl -s http://localhost/api/v1/books
```

---

## Part 8 — Bandwidth Management

Limit bandwidth for bulk-export endpoints to prevent resource starvation:

```yaml
# bandwidth/rate-limit-annotation.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: catalog
  namespace: bookstore
spec:
  template:
    metadata:
      annotations:
        kubernetes.io/egress-bandwidth: "10M"    # 10 Mbit/s egress cap
        kubernetes.io/ingress-bandwidth: "10M"   # 10 Mbit/s ingress cap
```

```bash
# Enable bandwidth manager (if not already)
helm upgrade cilium cilium/cilium --version 1.17.0 \
  --namespace kube-system \
  --reuse-values \
  --set bandwidthManager.enabled=true

# Verify bandwidth manager
kubectl -n kube-system exec ds/cilium -- cilium status | grep BandwidthManager
```

---

## Part 9 — Cluster Mesh (Multi-Cluster)

> [!NOTE]
> This requires two separate Kind clusters. Useful for learning disaster recovery and multi-region setups.

```bash
# Create a second cluster
kind create cluster --config cluster/kind-config.yaml --name cilium-bookstore-dr

# Install Cilium on the second cluster with a different cluster ID
helm install cilium cilium/cilium --version 1.17.0 \
  --namespace kube-system \
  --set cluster.name=bookstore-dr \
  --set cluster.id=2 \
  --set kubeProxyReplacement=true

# Enable Cluster Mesh on both clusters
cilium clustermesh enable --context kind-cilium-bookstore
cilium clustermesh enable --context kind-cilium-bookstore-dr

# Connect the clusters
cilium clustermesh connect \
  --context kind-cilium-bookstore \
  --destination-context kind-cilium-bookstore-dr

# Verify mesh status
cilium clustermesh status --context kind-cilium-bookstore
```

To make a service globally accessible:

```yaml
apiVersion: v1
kind: Service
metadata:
  name: catalog
  namespace: bookstore
  annotations:
    service.cilium.io/global: "true"      # Expose across clusters
    service.cilium.io/shared: "true"
```

---

## Troubleshooting & Debugging

```bash
# --- Cilium Status ---
cilium status --wait
cilium connectivity test

# --- Endpoints & Identities ---
kubectl -n kube-system exec ds/cilium -- cilium endpoint list
kubectl -n kube-system exec ds/cilium -- cilium identity list

# --- Policy Debugging ---
# Check which policies apply to a specific endpoint
kubectl -n kube-system exec ds/cilium -- cilium endpoint get <ENDPOINT_ID>
kubectl -n kube-system exec ds/cilium -- cilium policy get

# --- Monitor (real-time packet-level debug) ---
kubectl -n kube-system exec ds/cilium -- cilium monitor --type drop
kubectl -n kube-system exec ds/cilium -- cilium monitor --type policy-verdict

# --- Hubble (flow-level debug) ---
hubble observe --namespace bookstore --verdict DROPPED
hubble observe --namespace bookstore --to-pod bookstore/catalog -f

# --- BPF Maps (advanced) ---
kubectl -n kube-system exec ds/cilium -- cilium bpf ct list global
kubectl -n kube-system exec ds/cilium -- cilium bpf lb list

# --- Logs ---
kubectl -n kube-system logs ds/cilium --tail=100
kubectl -n kube-system logs deploy/cilium-operator --tail=100
kubectl -n kube-system logs deploy/hubble-relay --tail=100
```

---

## Resources

> [!IMPORTANT]
> **Official Documentation & Learning**
> - [Cilium Documentation](https://docs.cilium.io/)
> - [Cilium Getting Started Guide](https://docs.cilium.io/en/stable/gettingstarted/)
> - [Isovalent Labs (Free Hands-on)](https://isovalent.com/labs/)
> - [eBPF.io — What is eBPF?](https://ebpf.io/)
> - [Cilium GitHub Repository](https://github.com/cilium/cilium)
> - [Hubble Documentation](https://docs.cilium.io/en/stable/observability/hubble/)
> - [Cilium Network Policy Editor (visual)](https://editor.networkpolicy.io/)
> - [Cilium Cheat Sheet](https://cilium.io/blog/cheat-sheet/)

---

## Sub-Projects

| Directory                            | Description                                            |
| :----------------------------------- | :----------------------------------------------------- |
| [ebpf-go-basics](./ebpf-go-basics/) | Learn eBPF programming fundamentals using Go & `ebpf-go` |
