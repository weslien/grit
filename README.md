# GRIT
A go based monorepo tool. Convention over configuration.
## Installation
```bash
brew install grit
```

## Usage
```bash
grit [command] [options]
```
## Commands
### Initialize a new repository
Run in the root of a `git` repository
```bash
grit init
```

### fix up a grit repository
Run in the root of a `grit` repository
```bash
grit fixup
``` 
### Create a new package type
Run in the root of a `grit` repository
```bash
grit new type [name] 
```

### Create a new package
Run in the root of a `grit` repository
```bash
grit new [type] [name] 
```
### Create a new package from a template
Run in the root of a `grit` repository
```bash
grit new [type] [name] --template [github repo]
```
### Initialize a new template repository
Run in the root of a git repository
```bash
grit init-template
```

### Builds
Builds are run in the root of a `grit` repository. By default, the build system will build all dirty packages utilizing the build cache. To build all packages using detault settings, run the following command:
```bash
grit build
```
This will calculate which packages need to be built and build them according to the dependency graph. If the graph has changed, you may need to run `grit fixup` to update the graph.

To build a specific package and all of it's dirty dependencies, run the following command:
```bash
grit build [type] [name]
```

To bypass the build cache, run the following command:
```bash
grit build --no-cache
```

## Features
- [x] Package types
- [x] Package templates
- [x] Package dependencies
- [x] Package versioning
- [x] Package publishing
- [x] Package management
- [x] Polyglot
- [x] Build caching
- [x] CI/CD
- [x] Code generation
- [x] Prompt instructions included

## Conventions

The root of a grit repository contains a `grit.yaml` file. Among other things, this file contains the list of package types in the repository. Each type has it's own configuration, which can override the default configuration.

Packages of a [type] are located in the `packages/[type]` directory. A package of type `lib` with name `foo` would be located at `packages/lib/foo`.

A type is created merely by adding the type configuration to the `grit.yaml` file.

The input for builds are located in the `src` directory in the package's directory. This is where source code, static assets, configuration, and other files should be located, as this is where the build system will look for source code.

The build output of each package is located in the `build/[type]/[name]` directory.

Coverage reports are located in the `coverage/[type]/[name]` directory.

A package also contains the following directories:
- `.prompt`: Prompt instructions to coding agents specifically for that package.
- `.mod`: GritQL modifiers to align code to standards.
- `.dev`: settings and tools for local development.
- `.ops`: settings and tools for deployment and operations.

All packages define the following build targets:
- `build`: Build the package
- `install`: Install dependencies
- `run`: Run the package
- `mod`: Run GritQL modifiers to align code to standards
- `test`: Run tests
- `lint`: Lint the package
- `coverage`: Run tests and generate coverage reports
- `clean`: Clean the package (removing build output and coverage reports)
- `all`: Run all targets

Package configuration is located in the package's `grit.yaml` file.

Since each package is a completely separate module in any language using any build system, you need to map the package's build targets to the build system's build targets. This is done by defining the `build` property in the package's `grit.yaml` file. The `build` property is a map of build targets to build system targets. For example, the config for a `go` package would look like the following:
```yaml
targets:
  build: go build
  install: go install
  run: go run
  test: go test
  lint: golangci-lint run
  coverage: go test -coverprofile=coverage.out
  ```