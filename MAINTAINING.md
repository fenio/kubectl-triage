# Maintainer's Guide

This guide is for maintainers of kubectl-triage. It covers:
- Releasing new versions
- Building and testing
- Publishing to Krew
- Managing contributions

---

## 🚀 Release to GitHub

### Step 1: Push to GitHub

```bash
# Push the main branch
git push origin main

# Push the tag (this triggers GitHub Actions)
git push origin v0.1.0
```

**What happens:**
- GitHub Actions workflow (`.github/workflows/release.yml`) will automatically trigger
- GoReleaser will build binaries for all platforms:
  - Linux: amd64, 386
  - macOS: amd64
  - Windows: amd64, 386
- Binaries will be attached to the GitHub Release page
- Release will be created automatically with release notes

### Step 2: Verify GitHub Actions

1. Go to: `https://github.com/Lc-Lin/kubectl-triage/actions`
2. Watch the release workflow complete (~5-10 minutes)
3. Check for any errors

### Step 3: Verify Release

1. Go to: `https://github.com/Lc-Lin/kubectl-triage/releases`
2. Verify v0.1.0 release is created
3. Download and test binaries for your platform

---

## 📦 Install Locally (Before Publishing to Krew)

### Option 1: Copy Binary to PATH

```bash
# Copy the built binary
sudo cp ./bin/kubectl-triage /usr/local/bin/

# Verify it works
kubectl triage --help
```

### Option 2: Test with Krew (Local)

```bash
# Install locally for testing
kubectl krew install --manifest=./deploy/krew/plugin.yaml --archive=./bin/kubectl-triage

# Use it
kubectl triage <pod-name>
```

---

## 🎯 Submit to Krew Plugin Index

### Prerequisites

1. ✅ GitHub release v0.1.0 is published
2. ✅ Binaries are available on GitHub releases page
3. ✅ Plugin manifest is ready (`deploy/krew/plugin.yaml`)

### Steps

#### 1. Update Plugin Manifest SHA256

After GitHub release is published, download the binaries and calculate SHA256:

```bash
# Download release artifacts
curl -LO https://github.com/Lc-Lin/kubectl-triage/releases/download/v0.1.0/kubectl-triage_linux_amd64.tar.gz
curl -LO https://github.com/Lc-Lin/kubectl-triage/releases/download/v0.1.0/kubectl-triage_darwin_amd64.tar.gz
curl -LO https://github.com/Lc-Lin/kubectl-triage/releases/download/v0.1.0/kubectl-triage_windows_amd64.zip

# Calculate SHA256
sha256sum kubectl-triage_linux_amd64.tar.gz
sha256sum kubectl-triage_darwin_amd64.tar.gz
sha256sum kubectl-triage_windows_amd64.zip
```

Update the `sha256` fields in `deploy/krew/plugin.yaml` with the calculated values.

#### 2. Fork Krew Index

```bash
# Fork https://github.com/kubernetes-sigs/krew-index on GitHub
# Clone your fork
git clone https://github.com/YOUR_USERNAME/krew-index.git
cd krew-index
```

#### 3. Add Plugin Manifest

```bash
# Create plugin directory
mkdir -p plugins

# Copy your plugin manifest
cp /path/to/kubectl-triage/deploy/krew/plugin.yaml plugins/triage.yaml

# Commit
git add plugins/triage.yaml
git commit -m "Add kubectl-triage plugin

kubectl-triage is a fast diagnostic tool for failed Kubernetes pods.

It provides a 5-second diagnostic snapshot by intelligently aggregating:
- Pod status and container states
- Critical events (Warning/Error only)
- Previous crash logs (if container restarted)
- Current container logs

Only failed/restarted containers are shown by default, keeping output focused.

GitHub: https://github.com/Lc-Lin/kubectl-triage
"

# Push to your fork
git push origin main
```

#### 4. Create Pull Request

1. Go to: `https://github.com/kubernetes-sigs/krew-index`
2. Click "New Pull Request"
3. Select your fork
4. Title: `Add kubectl-triage plugin`
5. Description: Use the commit message above
6. Submit PR

#### 5. Wait for Review

- Krew maintainers will review your PR
- They may request changes
- Once approved and merged, your plugin will be available via:
  ```bash
  kubectl krew install triage
  ```

---

## 📖 Installation (After Krew Approval)

### Via Krew (Recommended)

```bash
# Update krew index
kubectl krew update

# Install kubectl-triage
kubectl krew install triage

# Use it
kubectl triage <pod-name>
```

### Manual Installation

#### Linux/macOS

```bash
# Download the binary
curl -LO https://github.com/Lc-Lin/kubectl-triage/releases/download/v0.1.0/kubectl-triage_$(uname -s)_amd64.tar.gz

# Extract
tar -xzf kubectl-triage_$(uname -s)_amd64.tar.gz

# Move to PATH
sudo mv kubectl-triage /usr/local/bin/

# Verify
kubectl triage --help
```

#### Windows (PowerShell)

```powershell
# Download
Invoke-WebRequest -Uri "https://github.com/Lc-Lin/kubectl-triage/releases/download/v0.1.0/kubectl-triage_windows_amd64.zip" -OutFile "kubectl-triage.zip"

# Extract
Expand-Archive kubectl-triage.zip -DestinationPath .

# Add to PATH or use directly
.\kubectl-triage.exe --help
```

---

## 🧪 Testing with Real Clusters

### Prerequisites

- Access to a Kubernetes cluster
- `kubectl` configured with valid context
- Some pods to test with (ideally including failed ones)

### Test Scenarios

#### 1. Test with Healthy Pod

```bash
# Find a healthy pod
kubectl get pods -A | grep Running | head -1

# Test early exit
kubectl triage <healthy-pod-name> -n <namespace>

# Expected output:
# ✅ Pod '<name>' is healthy (Ready 1/1, 0 restarts).
# Use --force to inspect anyway.
```

#### 2. Test with Failed Pod

```bash
# Find a failed pod
kubectl get pods -A | grep -E "CrashLoopBackOff|Error|ImagePullBackOff"

# Run triage
kubectl triage <failed-pod-name> -n <namespace>

# Expected output:
# - Pod status section
# - Critical events (Warning/Error only)
# - Previous logs (crash logs)
# - Current logs
```

#### 3. Test with --all-containers

```bash
# Test with a multi-container pod
kubectl triage <pod-name> -n <namespace> --all-containers
```

#### 4. Test with --lines

```bash
# Show more log context
kubectl triage <pod-name> -n <namespace> --lines=100
```

#### 5. Test with --force

```bash
# Force inspection of healthy pod
kubectl triage <healthy-pod-name> -n <namespace> --force
```

#### 6. Test with --no-color

```bash
# Disable colors (useful for CI/CD)
kubectl triage <pod-name> -n <namespace> --no-color
```

---

## 📊 Post-Release Checklist

- [ ] GitHub release v0.1.0 published
- [ ] All platform binaries available
- [ ] SHA256 checksums calculated
- [ ] Plugin manifest updated with SHA256s
- [ ] PR submitted to krew-index
- [ ] README badges updated (if applicable)
- [ ] Announcement posted (Twitter, Reddit, etc.)
- [ ] Added to awesome-kubectl-plugins list

---

## 🔄 Future Releases

For subsequent releases (v0.2.0, etc.):

```bash
# Make your changes
git add .
git commit -m "Description of changes"

# Tag new version
git tag -a v0.2.0 -m "Release v0.2.0: <description>"

# Push
git push origin main
git push origin v0.2.0
```

Then follow the same process:
1. GitHub Actions builds automatically
2. Update plugin manifest SHA256s
3. Submit PR to krew-index with updated version

---

## 🐛 Troubleshooting

### GitHub Actions Fails

- Check `.github/workflows/release.yml` for errors
- Verify GoReleaser configuration (`.goreleaser.yml`)
- Check build logs for missing dependencies

### Krew PR Rejected

Common reasons:
- Missing or incorrect SHA256 checksums
- Invalid plugin manifest format
- Plugin name conflicts
- License issues

### Binary Doesn't Work

- Check platform (amd64 vs arm64)
- Verify execute permissions: `chmod +x kubectl-triage`
- Check Go version compatibility

---

## 📝 Release Notes Template

For GitHub releases, use this template:

```markdown
## kubectl-triage v0.1.0

The 5-Second Diagnostic Snapshot for Failed Kubernetes Pods

### What's New

- Initial release of kubectl-triage
- Smart container filtering (shows only failed/restarted containers)
- Event noise reduction (Warning/Error only)
- Parallel log collection (previous + current logs)
- Keyword highlighting (ERROR, panic, fatal)
- Early exit for healthy pods
- Beautiful structured output with color-coded sections

### Installation

#### Via Krew (Coming Soon)
```bash
kubectl krew install triage
```

#### Manual Installation
Download the binary for your platform below and add to PATH.

### Usage

```bash
# Basic usage
kubectl triage <pod-name>

# With namespace
kubectl triage <pod-name> -n production

# Show all containers
kubectl triage <pod-name> --all-containers

# More log lines
kubectl triage <pod-name> --lines=100
```

### Full Documentation

See [README.md](https://github.com/Lc-Lin/kubectl-triage/blob/main/README.md) for complete documentation.

### Checksums

See release assets for SHA256 checksums.
```

---

## 🎉 You're Ready to Release!

All code is committed, tagged, and ready. Just run:

```bash
git push origin main --tags
```

And watch your plugin come to life! 🚀
