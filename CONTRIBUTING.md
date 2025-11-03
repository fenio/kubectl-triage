# Contributing to kubectl-triage

Thank you for your interest in contributing to kubectl-triage! We welcome contributions from the community.

## Getting Started

### Prerequisites

- Go 1.21 or later
- kubectl configured with access to a Kubernetes cluster
- Docker (optional, for testing with Kind)

### Development Setup

1. **Fork and clone the repository**

```bash
git clone https://github.com/YOUR_USERNAME/kubectl-triage.git
cd kubectl-triage
```

2. **Install dependencies**

```bash
go mod download
```

3. **Build the binary**

```bash
make bin
```

4. **Run tests**

```bash
make test
```

## Development Workflow

### Making Changes

1. Create a new branch for your changes:
```bash
git checkout -b feature/your-feature-name
```

2. Make your changes and ensure tests pass:
```bash
make test
make fmt
make vet
```

3. Test your changes with a real cluster:
```bash
./bin/kubectl-triage <pod-name> -n <namespace>
```

### Testing with Example Pods

We provide example test pods that cover various failure scenarios:

```bash
# Deploy test pods
kubectl apply -f examples/test-pods.yaml

# Test your changes
./bin/kubectl-triage crashloop-pod -n triage-test
./bin/kubectl-triage multi-container-pod -n triage-test

# Clean up
kubectl delete namespace triage-test
```

### Code Style

- Follow standard Go conventions and idioms
- Run `make fmt` to format your code
- Run `make vet` to check for common issues
- Add tests for new functionality
- Update documentation as needed

## Submitting Changes

### Pull Request Process

1. **Commit your changes**
```bash
git add .
git commit -m "Brief description of your changes"
```

2. **Push to your fork**
```bash
git push origin feature/your-feature-name
```

3. **Create a Pull Request**
   - Go to https://github.com/Lc-Lin/kubectl-triage
   - Click "New Pull Request"
   - Select your fork and branch
   - Fill in the PR template with:
     - Description of changes
     - Motivation and context
     - Testing performed
     - Related issues (if any)

### PR Guidelines

- **Keep PRs focused**: One feature or fix per PR
- **Write clear commit messages**: Explain what and why, not just what
- **Include tests**: Add or update tests for your changes
- **Update docs**: Update README.md or doc/USAGE.md if needed
- **Check CI**: Ensure all tests pass

## Types of Contributions

### Bug Reports

When filing an issue, please include:
- kubectl-triage version (`kubectl triage --version`)
- Kubernetes version (`kubectl version`)
- Operating system
- Steps to reproduce
- Expected vs actual behavior
- Relevant logs or output

### Feature Requests

We welcome feature requests! Please:
- Check if the feature already exists or is planned
- Explain the use case and why it's valuable
- Describe how you envision it working
- Consider kubectl-triage's philosophy: "The 5-Second Triage"
  - Features should maintain speed and simplicity
  - Focus on signal over noise

### Documentation

Documentation improvements are always welcome:
- Fix typos or unclear wording
- Add examples or use cases
- Improve installation instructions
- Translate documentation (future)

### Code Contributions

Areas where contributions are especially welcome:
- Additional failure detection patterns
- Performance improvements
- Test coverage improvements
- Support for new Kubernetes features
- Bug fixes

## Design Philosophy

When contributing, keep kubectl-triage's core principles in mind:

1. **Speed**: "5-Second Triage" - keep it fast
2. **Signal over noise**: Show only relevant information
3. **Opinionated**: Smart defaults, minimal configuration
4. **CLI-native**: Not a TUI, works with pipes and scripts
5. **Focused**: Does one thing well - diagnose pod failures

## Questions?

- **Usage questions**: See [USAGE.md](doc/USAGE.md)
- **Development questions**: Open an issue with the "question" label
- **Security issues**: Email maintainers privately (see SECURITY.md if available)

## Code of Conduct

Be respectful, inclusive, and professional. We're all here to make kubectl-triage better.

## License

By contributing, you agree that your contributions will be licensed under the Apache License 2.0.

---

Thank you for contributing to kubectl-triage! 🎉
