# Contributing to Grit

Thank you for your interest in contributing to Grit! This document outlines the development process, testing requirements, and quality standards.

## 🚀 Quick Start

1. **Fork and Clone**
   ```bash
   git clone https://github.com/yourusername/grit.git
   cd grit
   ```

2. **Install Dependencies**
   ```bash
   go mod download
   ```

3. **Run Tests**
   ```bash
   make test
   ```

4. **Build**
   ```bash
   make build
   ```

## 📋 Required Checks for Merge

All pull requests must pass the following automated checks before being merged:

### ✅ **CI Pipeline Jobs**

#### **1. Lint and Format (`lint`)**
- **golangci-lint** with comprehensive rule set
- **Go formatting** check (`gofmt`)
- **Import organization** check (`goimports`)
- **Dependency verification** (`go mod verify`)

**Required Status**: ✅ **MUST PASS**

#### **2. Tests (`test`)**
- **Unit tests** across Go versions 1.23 and 1.24
- **Race condition detection** (`-race` flag)
- **Code coverage** minimum 60% required
- **Coverage upload** to Codecov

**Required Status**: ✅ **MUST PASS**

#### **3. Build Verification (`build`)**
- **Cross-platform builds** for:
  - Linux (amd64, arm64)
  - macOS (amd64, arm64)  
  - Windows (amd64)
- **Artifact generation** for all platforms

**Required Status**: ✅ **MUST PASS**

#### **4. Security Scan (`security`)**
- **Gosec security scanner** (SARIF upload)
- **Vulnerability detection** (`govulncheck`)
- **Dependency security audit**

**Required Status**: ✅ **MUST PASS**

#### **5. Integration Tests (`integration`)**
- **CLI command validation**
- **Workspace operation tests**
- **Real-world scenario testing**

**Required Status**: ✅ **MUST PASS**

#### **6. Code Quality (`quality`)**
- **Ineffective assignment detection**
- **Unused code detection** (`staticcheck`)
- **Spelling check** (`misspell`)
- **Module tidiness** verification

**Required Status**: ✅ **MUST PASS**

#### **7. Dependency Audit (`dependencies`)**
- **Vulnerability scanning**
- **Dependency report generation**
- **License compliance**

**Required Status**: ✅ **MUST PASS**

## 🛠 Development Workflow

### **1. Branch Strategy**
- **Main branch**: `main` (protected, requires PR)
- **Development branch**: `develop` (integration branch)
- **Feature branches**: `feature/your-feature-name`
- **Bug fix branches**: `fix/bug-description`

### **2. Commit Standards**
Follow [Conventional Commits](https://www.conventionalcommits.org/):

```
type(scope): description

feat(cli): add graph visualization command
fix(build): resolve dependency caching issue
docs(readme): update installation instructions
test(graph): add dependency tree tests
refactor(format): improve output formatting
```

**Types**: `feat`, `fix`, `docs`, `style`, `refactor`, `test`, `chore`

### **3. Pull Request Process**

1. **Create Feature Branch**
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Make Changes and Test**
   ```bash
   # Run local checks
   make lint
   make test
   make build
   ```

3. **Commit and Push**
   ```bash
   git add .
   git commit -m "feat: add amazing new feature"
   git push origin feature/your-feature-name
   ```

4. **Open Pull Request**
   - Use the PR template
   - Link related issues
   - Add screenshots for UI changes
   - Ensure all CI checks pass

5. **Code Review**
   - Address reviewer feedback
   - Keep commits clean
   - Rebase if needed

6. **Merge**
   - Squash merge for feature branches
   - Merge commit for releases

## 🧪 Testing Standards

### **Unit Tests**
- **Coverage requirement**: Minimum 60%
- **Test file naming**: `*_test.go`
- **Table-driven tests** preferred
- **Mock external dependencies**

Example:
```go
func TestFormatHeader(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
    }{
        {"simple header", "Test", "═══ Test ═══"},
        {"empty header", "", "═══  ═══"},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            f := output.New()
            // Test implementation
        })
    }
}
```

### **Integration Tests**
- **CLI command testing**
- **Workspace scenario validation**
- **Error handling verification**

### **Benchmark Tests**
```go
func BenchmarkBuildCommand(b *testing.B) {
    // Benchmark implementation
}
```

## 🎯 Code Quality Standards

### **Linting Rules**
The project uses `.golangci.yml` with:
- **40+ enabled linters**
- **Cognitive complexity** limit: 7
- **Function length** limit: 100 lines
- **Line length** limit: 140 characters

### **Code Organization**
```
grit/
├── cmd/           # CLI commands
├── pkg/           # Reusable packages
│   ├── grit/      # Core functionality
│   └── output/    # Formatting utilities
├── .github/       # CI/CD workflows
└── docs/          # Documentation
```

### **Error Handling**
- **Always handle errors explicitly**
- **Provide contextual error messages**
- **Use wrapped errors** for debugging

```go
if err != nil {
    return fmt.Errorf("failed to load packages: %w", err)
}
```

### **Documentation**
- **Public functions** must have doc comments
- **Package documentation** in `doc.go`
- **Examples** for complex functionality

## 🚨 Branch Protection Rules

The following branches are protected and require:

### **Main Branch (`main`)**
- ✅ **Pull request required**
- ✅ **Dismiss stale reviews** when new commits are pushed
- ✅ **Require review from code owners**
- ✅ **All CI checks must pass**:
  - `lint` - Linting and formatting
  - `test` - Unit tests (Go 1.23, 1.24)
  - `build` - Cross-platform builds
  - `security` - Security scanning
  - `integration` - Integration tests
  - `quality` - Code quality checks
  - `dependencies` - Dependency audit
- ✅ **Up-to-date branch required**
- ✅ **Conversation resolution required**
- ❌ **Admin enforcement** (maintainers can bypass)

### **Develop Branch (`develop`)**
- ✅ **Pull request required**
- ✅ **CI checks must pass**
- ✅ **Up-to-date branch required**

## 🎁 Release Process

### **Versioning**
Following [Semantic Versioning](https://semver.org/):
- **MAJOR**: Breaking changes
- **MINOR**: New features (backward compatible)  
- **PATCH**: Bug fixes (backward compatible)

### **Release Steps**
1. **Update version** in relevant files
2. **Update CHANGELOG.md**
3. **Create release PR** to `main`
4. **Tag release** after merge:
   ```bash
   git tag -a v1.2.3 -m "Release v1.2.3"
   git push origin v1.2.3
   ```
5. **GitHub Actions** automatically:
   - Creates GitHub release
   - Builds cross-platform binaries
   - Generates completion scripts
   - Creates Docker images
   - Updates package managers

## 🛡 Security

### **Reporting Vulnerabilities**
- **Email**: security@grit-monorepo.dev
- **Response time**: 48 hours
- **Disclosure**: Coordinated disclosure process

### **Security Checks**
- **Gosec** static analysis
- **govulncheck** vulnerability scanning
- **Dependency audit** for known CVEs
- **SARIF upload** to GitHub Security

## 🏆 Recognition

Contributors are recognized in:
- **GitHub contributors graph**
- **Release notes** for significant contributions
- **Hall of Fame** section in README
- **Annual contributor spotlight**

## 📞 Getting Help

- **GitHub Issues**: Bug reports and feature requests
- **GitHub Discussions**: Questions and community help
- **Discord**: Real-time community chat
- **Email**: maintainers@grit-monorepo.dev

## 📜 License

By contributing to Grit, you agree that your contributions will be licensed under the same license as the project.

---

**Thank you for contributing to Grit!** 🚀

Your efforts help make Grit the best monorepo tool for developers worldwide.