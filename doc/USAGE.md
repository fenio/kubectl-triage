# kubectl-triage Usage Guide

## Installation

### Via Krew (Recommended)

```shell
kubectl krew install triage
```

### Manual Installation

Download the binary for your platform from the [releases page](https://github.com/lichenglin/kubectl-triage/releases) and place it in your PATH.

## Basic Usage

### Triage a Pod

The simplest usage - provide just the pod name:

```shell
kubectl triage <pod-name>
```

Example:
```shell
kubectl triage web-app-7d8f9c-xkz2q
```

### Specify Namespace

```shell
kubectl triage <pod-name> -n <namespace>
```

Example:
```shell
kubectl triage api-server-abc123 -n production
```

### Use Different Context

```shell
kubectl triage <pod-name> --context=staging
```

### Use Different Kubeconfig

```shell
kubectl triage <pod-name> --kubeconfig=/path/to/config
```

## Advanced Options

### Show All Containers

By default, `kubectl-triage` only shows containers that have failed or restarted. Use `--all-containers` to see all containers:

```shell
kubectl triage my-pod --all-containers
```

**Use case**: When you want to inspect sidecars or init containers even if they're healthy.

### Force Inspection of Healthy Pods

By default, kubectl-triage exits early if a pod is truly healthy (Running, all Ready, 0 restarts). Use `--force` to inspect anyway:

```shell
kubectl triage healthy-pod --force
```

**Use case**: When investigating performance issues or suspicious behavior in an otherwise "healthy" pod.

### Adjust Log Lines

Default is 50 lines. Increase or decrease as needed:

```shell
# Show more context
kubectl triage my-pod --lines=100

# Show less (faster, more concise)
kubectl triage my-pod --lines=20
```

**Use case**:
- Use higher values for Java stack traces or verbose errors
- Use lower values for quick scans

### Disable Colors

For piping output or CI/CD environments:

```shell
kubectl triage my-pod --no-color
```

**Use case**: When redirecting output to files or logs where ANSI codes would be messy.

## Common Scenarios

### Scenario 1: CrashLoopBackOff

When a pod is stuck in a crash loop:

```shell
$ kubectl triage backend-api-5d7f8-kx2p9

================================================================================
🚨 TRIAGE FOR FAILED CONTAINER: 'backend-api' (Reason: CrashLoopBackOff)
================================================================================

📋 POD STATUS
  Phase: CrashLoopBackOff | Restarts: 8 | Ready: 0/1

⚠️  CRITICAL EVENTS (Warning/Error only - Last 10)
  30s ago | Warning | BackOff        | Back-off restarting failed container
  2m ago  | Warning | BackOff        | Back-off restarting failed container
  5m ago  | Warning | BackOff        | Back-off restarting failed container

🔥 PREVIOUS LOGS (Last Crash) - Last 50 lines
--------------------------------------------------------------------------------
  2024/01/15 10:23:45 Starting application...
  2024/01/15 10:23:46 Connecting to database at postgres:5432
  panic: dial tcp: lookup postgres: no such host

  goroutine 1 [running]:
  main.connectDB()
      /app/main.go:42 +0x3e
  main.main()
      /app/main.go:18 +0x25
--------------------------------------------------------------------------------
```

**Diagnosis**: DNS resolution failure for database host.

**Next steps**: Check service names, ConfigMaps, or network policies.

---

### Scenario 2: OOMKilled

When a container is killed due to memory limits:

```shell
$ kubectl triage worker-6b8f4-hs9k4

================================================================================
🚨 TRIAGE FOR FAILED CONTAINER: 'worker' (Reason: OOMKilled)
================================================================================

📋 POD STATUS
  Phase: Running | Restarts: 3 | Ready: 1/1

⚠️  CRITICAL EVENTS (Warning/Error only - Last 10)
  5m ago  | Warning | BackOff        | Back-off restarting failed container
  7m ago  | Error   | OOMKilled      | Container was OOMKilled

🔥 PREVIOUS LOGS (Last Crash) - Last 50 lines
--------------------------------------------------------------------------------
  [2024-01-15 10:20:15] Processing batch job...
  [2024-01-15 10:20:16] Loading 100000 records into memory...
  [2024-01-15 10:20:17] Allocating buffers...
  [2024-01-15 10:20:18] Memory usage: 480MB
  [2024-01-15 10:20:19] Memory usage: 510MB
  <terminated>
--------------------------------------------------------------------------------
```

**Diagnosis**: Memory limit exceeded during batch processing.

**Next steps**: Increase memory limits or optimize batch size.

---

### Scenario 3: ImagePullBackOff

When Kubernetes can't pull the container image:

```shell
$ kubectl triage frontend-app-9c7d-x2k8

================================================================================
🚨 TRIAGE FOR FAILED CONTAINER: 'frontend' (Reason: ImagePullBackOff)
================================================================================

📋 POD STATUS
  Phase: Pending | Restarts: 0 | Ready: 0/1

⚠️  CRITICAL EVENTS (Warning/Error only - Last 10)
  1m ago  | Warning | Failed         | Failed to pull image "myregistry.io/frontend:v2.1.0": rpc error: code = Unknown desc = Error response from daemon: pull access denied
  2m ago  | Warning | Failed         | Error: ErrImagePull
  3m ago  | Warning | BackOff        | Back-off pulling image "myregistry.io/frontend:v2.1.0"
```

**Diagnosis**: Image pull authentication failure.

**Next steps**: Check image name, tag, and registry credentials.

---

### Scenario 4: Liveness Probe Failures

When health checks are failing:

```shell
$ kubectl triage api-gateway-4k8f-p9x2

================================================================================
🚨 TRIAGE FOR FAILED CONTAINER: 'api-gateway' (Reason: CrashLoopBackOff)
================================================================================

📋 POD STATUS
  Phase: Running | Restarts: 5 | Ready: 0/1

⚠️  CRITICAL EVENTS (Warning/Error only - Last 10)
  30s ago | Warning | Unhealthy      | Liveness probe failed: HTTP probe failed with statuscode: 500
  1m ago  | Warning | Unhealthy      | Liveness probe failed: HTTP probe failed with statuscode: 500
  2m ago  | Warning | Killing        | Killing container with id docker://api-gateway

🔥 PREVIOUS LOGS (Last Crash) - Last 50 lines
--------------------------------------------------------------------------------
  [INFO] Server listening on :8080
  [INFO] Health check endpoint: /healthz
  [ERROR] Database connection lost: connection refused
  [ERROR] Health check failed: database unavailable
  [WARN] Returning HTTP 500 for /healthz
--------------------------------------------------------------------------------
```

**Diagnosis**: Liveness probe failing due to database connection issue.

**Next steps**: Fix database connectivity or adjust probe thresholds.

---

### Scenario 5: Multi-Container Pod (One Failed)

When one container fails in a multi-container pod:

```shell
$ kubectl triage web-pod-7f8k-m3x9

================================================================================
🚨 TRIAGE FOR FAILED CONTAINER: 'app' (Reason: Error)
================================================================================

📋 POD STATUS
  Phase: Running | Restarts: 2 | Ready: 2/3

⚠️  CRITICAL EVENTS (Warning/Error only - Last 10)
  1m ago  | Warning | BackOff        | Back-off restarting failed container app in pod web-pod-7f8k-m3x9

🔥 PREVIOUS LOGS (Last Crash) - Last 50 lines
--------------------------------------------------------------------------------
  Config file not found: /etc/config/app.yaml
  ERROR: failed to initialize application
  exit code: 1
--------------------------------------------------------------------------------

--------------------------------------------------------------------------------
ℹ️  2 other containers running normally: [istio-proxy (Running, 0 restarts), log-collector (Running, 0 restarts)]
```

**Diagnosis**: Missing config file in the main app container. Other containers (sidecars) are healthy.

**Next steps**: Check ConfigMap mount or volume configuration.

---

## Flags Reference

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--lines` | int64 | 50 | Number of log lines to display |
| `--all-containers` | bool | false | Show all containers, not just failed ones |
| `--force` | bool | false | Inspect pod even if it appears healthy |
| `--no-color` | bool | false | Disable colored output |
| `-n, --namespace` | string | default | Kubernetes namespace |
| `--context` | string | current | Kubeconfig context to use |
| `--kubeconfig` | string | ~/.kube/config | Path to kubeconfig file |

## Understanding the Output

### Section 1: Pod Status

```
📋 POD STATUS
  Phase: CrashLoopBackOff | Restarts: 5 | Ready: 0/2
```

- **Phase**: Current pod phase (Running, Pending, Failed, etc.)
- **Restarts**: Number of times this specific container has restarted
- **Ready**: How many containers are ready out of total (e.g., 1/2 means 1 out of 2 containers is ready)

### Section 2: Critical Events

```
⚠️  CRITICAL EVENTS (Warning/Error only - Last 10)
  2m ago  | Warning | BackOff        | Back-off restarting failed container
```

- **Only Warning and Error events are shown** (Normal events are filtered out)
- Events are sorted by time (most recent first)
- Limited to last 10 events to keep output concise

### Section 3: Previous Logs

```
🔥 PREVIOUS LOGS (Last Crash) - Last 50 lines
```

- Logs from the **previous run** of the container (before it crashed)
- This is often the most important diagnostic information
- Only shown if the container has restarted at least once
- Lines containing ERROR, panic, fatal, etc. are **highlighted in red**

### Section 4: Current Logs

```
🔄 CURRENT LOGS - Last 50 lines
```

- Logs from the **current run** of the container
- Useful to see if the container is stuck in a restart loop or progressing differently

### Section 5: Healthy Containers Summary

```
ℹ️  2 other containers running normally: [istio-proxy (Running, 0 restarts), log-collector (Running, 0 restarts)]
```

- One-line summary of containers that are healthy
- Shows you didn't miss anything
- Confirms the problem is isolated to specific containers

## Tips and Best Practices

### 1. Make it Your First Command

When you see a pod failure alert, make `kubectl triage` your reflex:

```shell
# Instead of this workflow:
kubectl get pods | grep -i error
kubectl describe pod <name> | less
kubectl logs <name>
kubectl logs <name> --previous

# Just do this:
kubectl triage <pod-name>
```

### 2. Combine with kubectl get

```shell
# Find failing pods
kubectl get pods | grep -v Running

# Triage them
kubectl triage <failing-pod-name>
```

### 3. Pipe to Less for Long Outputs

```shell
kubectl triage my-pod | less
```

### 4. Save Output for Sharing

```shell
kubectl triage my-pod > triage-report.txt
```

Share the report with your team for collaborative debugging.

### 5. Use --no-color for Logs

When saving to files or posting in Slack:

```shell
kubectl triage my-pod --no-color | pbcopy  # macOS
kubectl triage my-pod --no-color | xclip   # Linux
```

## Troubleshooting

### "Pod not found"

```
Error: failed to get pod my-pod in namespace default: pods "my-pod" not found
```

**Solution**: Check pod name and namespace. Use `-n` to specify the correct namespace.

### "No previous logs"

```
🔥 PREVIOUS LOGS
  (No previous logs: previous terminated container "app" in pod "my-pod" not found)
```

**Meaning**: The container has never crashed before. This is its first run.

### "Forbidden: User cannot get pods"

```
Error: failed to get pod my-pod: pods "my-pod" is forbidden: User "john" cannot get resource "pods" in API group "" in the namespace "production"
```

**Solution**: Check your RBAC permissions. You need at least `get`, `list` permissions on pods and events.

## Further Reading

- [kubectl-triage README](https://github.com/lichenglin/kubectl-triage)
- [Kubernetes Debugging Pods](https://kubernetes.io/docs/tasks/debug-application-cluster/debug-application/)
- [Understanding Pod Lifecycle](https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle/)

---

**Need help?** File an issue on [GitHub](https://github.com/lichenglin/kubectl-triage/issues).
