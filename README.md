# kubectl-triage

> The 5-Second Diagnostic Snapshot for Failed Kubernetes Pods

`kubectl-triage` is a kubectl plugin that instantly diagnoses failing pods by intelligently aggregating the three critical pieces of information you always need — all in one command, all in 5 seconds.

## The Problem

When a pod crashes, you're stuck running these three commands in sequence:

```bash
kubectl describe pod <name>    # Scroll to bottom for Events
kubectl logs <name>             # Check current logs
kubectl logs <name> --previous  # Find the actual crash (this is the key!)
```

This is tedious, repetitive, and wastes precious debugging time every single day.

## The Solution

```bash
kubectl triage <pod-name>
```

That's it. **One command. One screen. 5 seconds.**

## What It Is

- **A First-Responder Tool**: When a pod turns red, this is the first command you run
- **A Diagnostic Snapshot**: Shows the "last moment" of a pod failure in one screen
- **Opinionated & Focused**: Only shows what matters (Warning/Error events, failed containers, crash logs)
- **CLI-Native**: Runs in your shell, outputs results, then exits — integrates with your existing workflow

## What It Is NOT

- ❌ Not a TUI (like k9s) — it's a command, not an application
- ❌ Not a log follower — use `kubectl logs -f` for live streaming
- ❌ Not an interactive debugger — use `kubectl debug` for that
- ❌ Not a replacement for `kubectl describe` — it only shows failure-related info

## Installation

### Via Krew (Recommended)

```bash
kubectl krew install triage
```

### Manual Installation

Download the latest release for your platform from the [releases page](https://github.com/lichenglin/kubectl-triage/releases).

Extract and move the binary to your PATH:

```bash
tar -xzf kubectl-triage_<platform>_<arch>.tar.gz
mv kubectl-triage /usr/local/bin/
```

Verify installation:

```bash
kubectl triage --help
```

## Usage

### Basic Usage

```bash
# Triage a failing pod
kubectl triage my-crashing-pod

# Triage a pod in a specific namespace
kubectl triage my-pod -n production

# Triage using a different kubeconfig context
kubectl triage my-pod --context staging
```

### Advanced Options

```bash
# Show all containers (not just failed ones)
kubectl triage my-pod --all-containers

# Force inspection of healthy pods
kubectl triage my-pod --force

# Show more log lines (default is 50)
kubectl triage my-pod --lines=100

# Disable colored output (for CI/CD)
kubectl triage my-pod --no-color
```

## What You Get

`kubectl-triage` provides a structured diagnostic output:

```
================================================================================
🚨 TRIAGE FOR FAILED CONTAINER: 'app' (Reason: CrashLoopBackOff)
================================================================================

📋 POD STATUS
  Phase: CrashLoopBackOff | Restarts: 5 | Ready: 0/2

⚠️  CRITICAL EVENTS (Warning/Error only - Last 10)
  2m ago  | Warning | BackOff        | Back-off restarting failed container
  5m ago  | Warning | Unhealthy      | Liveness probe failed: HTTP probe failed
  8m ago  | Error   | Failed         | Error: ImagePullBackOff

🔥 PREVIOUS LOGS (Last Crash) - Last 50 lines
--------------------------------------------------------------------------------
  panic: runtime error: invalid memory address or nil pointer dereference
  [signal SIGSEGV: segmentation violation code=0x1 addr=0x0 pc=0x...]
  goroutine 1 [running]:
  main.main()
      /app/main.go:42 +0x3e
--------------------------------------------------------------------------------

🔄 CURRENT LOGS - Last 50 lines
--------------------------------------------------------------------------------
  Starting application...
  Loading configuration...
  Connecting to database...
--------------------------------------------------------------------------------

ℹ️  1 other container running normally: [istio-proxy (Running, 0 restarts)]
```

## Key Features

### 1. Smart Container Filtering

Only shows containers that have **actually failed** or **restarted**. A container is flagged if:
- Current state is `CrashLoopBackOff`, `Error`, `ImagePullBackOff`, `OOMKilled`, etc.
- `RestartCount > 0` (the golden indicator — means it crashed before)

Healthy sidecars are summarized in one line at the bottom.

### 2. Signal Over Noise: Events Filtering

**Only displays Warning and Error events**, filtering out all the "Normal" events like:
- ✅ Scheduled
- ✅ Pulling
- ✅ Pulled
- ✅ Created
- ✅ Started

You only see what went wrong, not what went right.

### 3. True Health Detection

A pod is considered "healthy" (and triage exits early) **only if**:
1. Phase is `Running`
2. All containers are `Ready`
3. `RestartCount == 0` for all containers

If a pod is Running but has restarted 5 times, `kubectl-triage` will catch it and show you the previous crash logs.

### 4. Parallel Log Collection

Fetches current and previous logs for all failed containers **in parallel** using goroutines, keeping the total execution time under 3 seconds even for multi-container pods.

### 5. Keyword Highlighting

Automatically highlights critical keywords in logs:
- 🔴 `ERROR`, `panic`, `fatal`, `exception`, `OOMKilled`, `failed`

## Philosophy: The 5-Second Triage

This tool embodies the **Platform Engineering** principle of **eliminating friction**.

Instead of:
1. Run describe → scroll → find events → note container name
2. Run logs → check current state
3. Run logs --previous → find actual crash
4. Correlate all three in your head

You get:
1. Run `kubectl triage <pod>`
2. **Done.**

It fills the gap between "personal shell aliases" and "heavy TUI tools like k9s" — providing a **team-shareable, standardized first-response tool** for pod failures.

## Real-World Scenarios

### Scenario 1: CrashLoopBackOff

```bash
$ kubectl triage web-app-7d8f9c-xkz2q

🚨 TRIAGE FOR FAILED CONTAINER: 'web-app' (Reason: CrashLoopBackOff)
📋 POD STATUS: Phase: CrashLoopBackOff | Restarts: 12
🔥 PREVIOUS LOGS: panic: database connection failed
```

**Diagnosis in 5 seconds**: Database connection issue.

### Scenario 2: OOMKilled

```bash
$ kubectl triage worker-64b8f-hs9k4

🚨 TRIAGE FOR FAILED CONTAINER: 'worker' (Reason: OOMKilled)
⚠️  CRITICAL EVENTS: OOMKilling container (memory limit exceeded)
🔥 PREVIOUS LOGS: [huge memory allocation traces]
```

**Diagnosis in 5 seconds**: Memory leak or undersized limits.

### Scenario 3: Healthy Pod (Early Exit)

```bash
$ kubectl triage api-server-abc123

✅ Pod 'api-server-abc123' is healthy (Ready 1/1, 0 restarts).
Use --force to inspect anyway.
```

**Saves time**: No need to dig into a working pod.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

### Development

```bash
# Clone the repo
git clone https://github.com/lichenglin/kubectl-triage.git
cd kubectl-triage

# Build
make bin

# Run tests
make test

# Run locally
./bin/kubectl-triage <pod-name>
```

## Roadmap

- [ ] Add JSON/YAML output format (`--output=json`)
- [ ] Support for init containers
- [ ] Configurable keyword highlighting
- [ ] Event filtering by time window
- [ ] Integration with common observability platforms

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Credits

Inspired by the daily frustration of every Kubernetes developer who's ever typed `kubectl logs --previous` for the thousandth time.

---

**Made with ❤️ by developers, for developers.**

If this tool saved you time, give it a ⭐️ on GitHub!
