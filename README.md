# Mini Project - GitOps with ArgoCD & Canary Deployment

A demo project showcasing GitOps practices with ArgoCD, canary deployments using Argo Rollouts, and Gateway API with Traefik.

## Architecture

- **ArgoCD**: GitOps continuous delivery
- **Argo Rollouts**: Progressive delivery with canary deployments
- **Traefik**: Ingress controller with Gateway API support
- **Prometheus**: Metrics for canary analysis
- **Gateway API**: Kubernetes-native traffic routing
- **Demo App**: Simple Go application for testing

## Prerequisites

- Kubernetes cluster (v1.21+)
- kubectl
- helm (v3+)

## Quick Start

### 1. Install Gateway API CRDs

```bash
kubectl apply -f https://github.com/kubernetes-sigs/gateway-api/releases/download/v1.4.0/standard-install.yaml
```

### 2. Install ArgoCD

```bash
helm repo add argo https://argoproj.github.io/argo-helm
helm repo update

helm upgrade --install argocd argo/argo-cd \
  --namespace argocd \
  --create-namespace \
  --values argocd/helm-values/values.yaml \
  --wait
```

### 3. Deploy All Applications

```bash
kubectl apply -f argocd/
```

This will deploy:
- ArgoCD Applications (Traefik, Argo Rollouts, Prometheus, Demo App)

### 4. Configure Domain Access

Wait for Traefik to be ready, then add hosts:

```bash
# Wait for Traefik LoadBalancer IP
kubectl wait --for=condition=available deployment/traefik -n traefik-system --timeout=120s

# Get Traefik IP and add to /etc/hosts
TRAEFIK_IP=$(kubectl get svc -n traefik-system traefik -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
echo "$TRAEFIK_IP argocd.local traefik.local demo-app.example.com" | sudo tee -a /etc/hosts
```

### 5. Access Services

Get ArgoCD admin password:

```bash
kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath='{.data.password}' | base64 -d && echo
```

Access URLs:
- **ArgoCD**: http://argocd.local
- **Traefik Dashboard**: http://traefik.local/dashboard/
- **Demo App**: http://demo-app.example.com

## Project Structure

```
.
├── app/                          # Demo application source code
│   ├── Dockerfile
│   ├── main.go
│   └── go.mod
├── argocd/                       # ArgoCD manifests & config
│   ├── argo-rollouts-app.yaml   # Argo Rollouts (argoproj.github.io/argo-helm)
│   ├── prometheus-app.yaml      # Prometheus (prometheus-community.github.io)
│   ├── traefik-app.yaml         # Traefik (traefik.github.io/charts)
│   ├── demo-app.yaml            # Demo App (helm/demo-app)
│   └── helm-values/
│       └── values.yaml          # ArgoCD Helm values (includes HTTPRoute)
├── helm/
│   └── demo-app/                # Demo app Helm chart
│       ├── Chart.yaml
│       ├── values.yaml
│       └── templates/
│           ├── rollout.yaml     # Argo Rollout with canary strategy
│           ├── analysis.yaml    # AnalysisTemplate for success rate
│           ├── service.yaml     # Stable + Canary services
│           └── httproute.yaml
└── .github/
    └── workflows/
        └── build.yaml           # CI: Build & push Docker image
```

## Canary Deployment

The demo app uses Argo Rollouts for progressive canary deployment with automatic analysis:

### Deployment Strategy

| Step | Weight | Action |
|------|--------|--------|
| 1 | 20% | Pause 60s |
| 2 | 40% | Pause 60s |
| 3 | 60% | Pause 60s |
| 4 | 80% | Pause 60s |
| 5 | 100% | Promote to stable |

Background analysis runs continuously throughout the deployment, checking success rate every minute.

### Analysis Template

The analysis uses Prometheus to check the success rate:
- **Metric**: HTTP request success rate (non-5xx / total)
- **Threshold**: >= 99% success rate
- **Interval**: 1 minute
- **Failure Limit**: 2 (rollback if exceeded)

If no traffic is present, the analysis returns success by default (`or vector(1)`).

Traffic splitting is managed through Gateway API HTTPRoute.

### Testing Canary Rollback

The demo app supports fault injection via `errorRate` value:

```bash
# Set errorRate to 50 in helm/demo-app/values.yaml to simulate 50% errors
# This will trigger analysis failure and automatic rollback

# Generate traffic to trigger the analysis
while true; do curl -s http://demo-app.local; sleep 0.5; done
```

## Features

- ✅ GitOps with ArgoCD
- ✅ Automated canary deployments with Argo Rollouts
- ✅ Prometheus metrics & analysis for automatic rollback
- ✅ Traffic management with Gateway API
- ✅ Traefik as Gateway controller with Dashboard
- ✅ ArgoCD Rollout Extension for UI visualization
- ✅ Image tag set to Git commit SHA
- ✅ Fault injection for testing rollback (`errorRate`)

## Useful Commands

### Watch Rollout Progress

```bash
kubectl argo rollouts get rollout demo-app -n demo-app --watch
```

### Trigger a New Deployment

```bash
# ArgoCD will automatically sync and deploy using the commit SHA
git commit -am "Update application"
git push
```

### Manual Rollout Control

```bash
# Promote canary to stable
kubectl argo rollouts promote demo-app -n demo-app

# Skip all steps and promote immediately
kubectl argo rollouts promote demo-app -n demo-app --full

# Abort rollout
kubectl argo rollouts abort demo-app -n demo-app

# Restart rollout
kubectl argo rollouts restart demo-app -n demo-app
```

### View ArgoCD Applications

```bash
kubectl get applications -n argocd
```

## Cleanup

```bash
# Delete all ArgoCD applications
kubectl delete -f argocd/

# Uninstall ArgoCD
helm uninstall argocd -n argocd

# Delete Gateway API CRDs
kubectl delete -f https://github.com/kubernetes-sigs/gateway-api/releases/download/v1.4.0/standard-install.yaml

# Delete namespaces
kubectl delete namespace argocd traefik-system demo-app argo-rollouts monitoring
```

## Troubleshooting

### Gateway not working

```bash
kubectl get gateway -A
kubectl describe gateway traefik-gateway -n traefik-system
```

### HTTPRoute not attached

```bash
kubectl get httproute -A
kubectl describe httproute demo-app -n demo-app
```

### Rollout stuck

```bash
kubectl argo rollouts status demo-app -n demo-app
kubectl describe rollout demo-app -n demo-app
```

### Analysis failing

```bash
kubectl get analysisrun -n demo-app
kubectl describe analysisrun <name> -n demo-app
```

Common issues:
- **No traffic**: Analysis requires traffic to calculate success rate. Send requests during deployment.
- **Prometheus not reachable**: Check Prometheus service is running in monitoring namespace.
- **No metrics**: Ensure the app exposes `/metrics` endpoint and has `prometheus.io/scrape: "true"` annotation.

### Skip analysis and promote manually

```bash
kubectl argo rollouts promote demo-app -n demo-app --full
```

## License

MIT
