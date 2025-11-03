# kubectl-triage Examples

This directory contains example files for testing and demonstrating kubectl-triage functionality.

## test-pods.yaml

A collection of test pods that cover various failure scenarios in Kubernetes. These pods are useful for:
- Testing kubectl-triage functionality
- Demonstrating different pod failure modes
- Learning how kubectl-triage handles various scenarios

### Included Test Scenarios

The file includes 7 different pod configurations:

1. **healthy-nginx** - A healthy nginx pod (Running, no restarts)
   - Use case: Test early exit for healthy pods

2. **crashloop-pod** - A pod in CrashLoopBackOff state
   - Use case: Test crash loop detection and previous log retrieval

3. **imagepull-error** - A pod with ImagePullBackOff error
   - Use case: Test handling of image pull failures

4. **oom-pod** - A pod that will be OOMKilled
   - Use case: Test out-of-memory detection

5. **multi-container-pod** - Multi-container pod with one failed container
   - Use case: Test intelligent container filtering

6. **liveness-failed** - A pod with liveness probe failures
   - Use case: Test probe failure detection

7. **healthy-multi** - A healthy multi-container pod
   - Use case: Test --all-containers flag

### Usage

#### Deploy test pods to your cluster:

```bash
kubectl apply -f examples/test-pods.yaml
```

#### Wait for pods to reach their intended states:

```bash
# Watch pods start up
kubectl get pods -n triage-test -w
```

#### Test kubectl-triage with different scenarios:

```bash
# Test healthy pod (should exit early)
kubectl triage healthy-nginx -n triage-test

# Test CrashLoopBackOff pod (full diagnostic)
kubectl triage crashloop-pod -n triage-test

# Test ImagePullBackOff
kubectl triage imagepull-error -n triage-test

# Test multi-container filtering
kubectl triage multi-container-pod -n triage-test

# Test --all-containers flag
kubectl triage healthy-multi -n triage-test --all-containers

# Test --force flag on healthy pod
kubectl triage healthy-nginx -n triage-test --force

# Test --lines flag
kubectl triage crashloop-pod -n triage-test --lines=100
```

#### Clean up:

```bash
kubectl delete namespace triage-test
```

## Notes

- These test pods are designed to fail intentionally for testing purposes
- The namespace `triage-test` is created automatically when applying the YAML
- Some pods may take a few moments to reach their intended failure states
- The OOM pod may take 10-30 seconds to trigger the out-of-memory condition
