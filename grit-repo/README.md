# Grit Test Repository

This repository serves as a comprehensive test environment for the **grit** monorepo tool, featuring various package types, build systems, and target configurations to ensure feature parity and comprehensive testing coverage.

## Repository Structure

```
grit-repo/
├── grit.yaml                     # Root configuration with package types and global targets
└── packages/
    ├── lib/                      # Library packages
    │   ├── utils/                # Basic utility library (no dependencies)
    │   ├── core/                 # Core library (depends on utils)
    │   └── validation/           # Validation library (depends on utils)
    ├── app/                      # Application packages
    │   ├── web-server/           # Web server app (depends on core, validation)
    │   ├── cli-tool/             # CLI tool (depends on utils)
    │   └── worker/               # Background worker (depends on core)
    ├── service/                  # Service packages with deployment capabilities
    │   ├── api-gateway/          # API gateway (depends on core, validation)
    │   ├── auth-service/         # Authentication service (depends on core)
    │   └── payment-service/      # Payment service (depends on core, validation)
    ├── frontend/                 # Frontend packages
    │   ├── admin-dashboard/      # React admin dashboard
    │   └── user-portal/          # Vue.js user portal
    ├── python/                   # Python packages
    │   ├── data-processor/       # Data processing utilities
    │   └── ml-models/            # Machine learning models
    └── rust/                     # Rust packages
        ├── crypto-utils/         # Cryptographic utilities
        └── performance-monitor/  # Performance monitoring tools
```

## Package Types and Features

### Library Packages (`lib`)
- **No external dependencies allowed**
- Basic build, test, lint, coverage targets
- Documentation generation
- Benchmark testing

### Application Packages (`app`)
- **Can depend on library packages**
- Extended targets: run, install, docker build/run
- Integration testing
- CLI completion generation

### Service Packages (`service`)
- **Can depend on lib and app packages**
- Full deployment pipeline: Docker, Kubernetes, Helm
- Load testing, security scanning
- Database migration and seeding
- End-to-end testing

### Frontend Packages (`frontend`)
- **Independent packages (no grit dependencies)**
- Modern build tools (Vite, webpack)
- Testing: unit, component, e2e, visual
- Storybook integration
- Performance auditing (Lighthouse)

### Python Packages (`python`)
- **Standard Python tooling**
- Build system: setuptools/build
- Testing: pytest with coverage
- Code quality: black, isort, flake8, mypy
- Security: bandit
- ML-specific: training, evaluation, model export

### Rust Packages (`rust`)
- **Cargo-based build system**
- Comprehensive testing and benchmarking
- Code formatting and linting (rustfmt, clippy)
- Security auditing
- Performance profiling and optimization

## Available Targets

### Common Targets (all packages)
- `build` - Build the package
- `test` - Run tests
- `lint` - Run linters and code quality checks
- `coverage` - Generate test coverage reports
- `clean` - Clean build artifacts

### Language-Specific Targets

#### Go Packages
- `run` - Execute the application
- `install` - Install the binary
- `benchmark` - Run Go benchmarks
- `docs` - Generate documentation

#### Frontend Packages
- `dev` - Start development server
- `serve` - Serve production build
- `type-check` - TypeScript type checking
- `format` - Code formatting
- `storybook` - Run Storybook
- `e2e` - End-to-end testing

#### Service Packages
- `docker-build/run/push` - Docker operations
- `k8s-deploy/delete` - Kubernetes deployment
- `helm-install` - Helm chart installation
- `load-test` - Performance testing
- `security-scan` - Security vulnerability scanning

#### Python Packages
- `format` - Code formatting (black, isort)
- `type-check` - Type checking (mypy)
- `security` - Security analysis (bandit)
- `notebook` - Launch Jupyter Lab
- `train/evaluate/predict` - ML operations

#### Rust Packages
- `check` - Fast compilation check
- `fmt` - Code formatting
- `audit` - Security audit
- `bench` - Benchmarking
- `doc` - Documentation generation
- `flamegraph` - Performance profiling

## Testing the Grit Tool

This repository provides comprehensive test cases for:

1. **Dependency Management**: Complex dependency graphs between packages
2. **Multi-Language Support**: Go, JavaScript/TypeScript, Python, Rust
3. **Build System Integration**: Various build tools and package managers
4. **Target Execution**: Wide variety of build, test, and deployment targets
5. **Package Types**: Different package categories with specific capabilities
6. **Monorepo Operations**: Cross-package operations and dependency resolution

## Usage Examples

```bash
# Build all packages
grit build

# Test a specific package
grit test packages/lib/utils

# Run all linting
grit lint

# Build and run a service
grit build packages/service/api-gateway
grit run packages/service/api-gateway

# Frontend development
grit dev packages/frontend/admin-dashboard

# Python ML workflow
grit train packages/python/ml-models
grit evaluate packages/python/ml-models
```

This test repository ensures the grit tool can handle real-world monorepo scenarios with complex dependency relationships, multiple programming languages, and diverse build requirements.