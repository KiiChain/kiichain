# Contribution Guidelines

Contributions to KiiChain are welcome. As a contributor, here are the guidelines we would like you to follow:

- [Code of Conduct](#code-of-conduct)
- [Issues and Bugs](#found-a-bug)
- [Feature Requests](#missing-a-feature)
- [Local Development Setup](#local-development-setup)
- [Submission Guidelines](#submission-guidelines)
- [Coding Rules](#coding-rules)
- [Commit Message Guidelines](#commit-message-convention)

## Code of Conduct

Please read and follow our [Code of Conduct][coc].

## Found a Bug?

If you find a bug in the source code [submit a bug report issue](#submitting-an-issue).
Even better, you can [submit a Pull Request](#submitting-a-pull-request-pr) with a fix.

## Missing a Feature?

You can *request* a new feature by [submitting a feature request issue](#submitting-an-issue).
If you would like to *implement* a new feature:

- For a **Major Feature**, first [open an issue](#submitting-an-issue) and outline your proposal so that it can be discussed.
- **Small Features** can be crafted and directly [submitted as a Pull Request](#submitting-a-pull-request-pr).

## Local Development Setup

### Prerequisites

| Tool | Version | Required |
|------|---------|----------|
| [Go](https://golang.org/dl/) | 1.24+ | Yes |
| [Git](https://git-scm.com/) | any recent | Yes |
| [GNU Make](https://www.gnu.org/software/make/) | any recent | Recommended |
| [Docker](https://docs.docker.com/get-docker/) | any recent | For e2e tests only |

Verify your setup:

```bash
go version       # Should show 1.24+
git --version
make --version   # Optional but recommended
docker --version # Only needed for e2e tests
```

### Clone and Build

```bash
# 1. Fork the repository on GitHub, then clone your fork
git clone https://github.com/<your-username>/kiichain.git
cd kiichain

# 2. Add the upstream remote
git remote add upstream https://github.com/KiiChain/kiichain.git

# 3. Create a feature branch
git checkout -b feat/my-change

# 4. Build the binary
make build          # Output: ./build/kiichaind

# Or install to $GOPATH/bin
make install
```

If you don't have GNU Make:

```bash
go build -o ./build/kiichaind ./cmd/kiichaind
```

### Running Tests

**Unit tests:**

```bash
make test-unit              # Run all unit tests (5 min timeout)
make test-unit-cover        # Run with coverage report
make test-unit-cover-html   # Generate HTML coverage report
```

Or directly with `go test`:

```bash
go test ./... -count=1 -timeout=5m
```

**Race condition detection:**

```bash
make test-race
```

**End-to-end tests** (requires Docker):

```bash
# Build the required Docker images first
docker build -t kiichain/kiichaind-e2e -f Dockerfile .
cd tests/e2e/docker && docker build -t kiichain/hermes-e2e:1.0.0 -f hermes.Dockerfile .
cd ../../..

# Run e2e tests
make test-e2e
```

### Linting and Formatting

```bash
make lint       # Run golangci-lint
make lint-fix   # Auto-fix lint issues where possible
make format     # Format code
```

### Protobuf

If you modify `.proto` files under `proto/`:

```bash
make proto-gen          # Regenerate Go code from proto definitions
make proto-format       # Format .proto files
make proto-lint         # Lint .proto files
```

### EVM Precompiles

If you modify Solidity interfaces under `precompiles/`:

```bash
make compile-evm-precompiles
```

## Submission Guidelines

### Submitting an Issue

Before you submit an issue, please search the [issue tracker][issues]. An issue for your problem might already exist and the discussion might inform you of workarounds readily available.

For bug reports, it is important that we can reproduce and confirm it. For this, we need you to provide a minimal reproduction instruction (this is part of the bug report issue template).

You can file new issues by selecting from our [new issue templates][new-issue] and filling out the issue template.

### Submitting a Pull Request (PR)

Before you submit your Pull Request (PR) consider the following guidelines:

1. All Pull Requests should be based off of and opened against the `main` branch.

2. Search [Existing PRs][prs] for an open or closed PR that relates to your submission.
   You don't want to duplicate existing efforts.

3. Be sure that an issue exists describing the problem you're fixing, or the design for the feature you'd like to add.

4. [Fork](https://docs.github.com/en/github/getting-started-with-github/fork-a-repo) the [repository][github].

5. In your forked repository, make your changes in a new git branch created off of the `main` branch.

6. Make your changes, **including test cases and documentation updates where appropriate**.

7. Follow our [coding rules](#coding-rules).

8. Run all tests and checks locally, as described in the [development guide][developing], and ensure they pass. This saves CI hours and ensures you only commit clean code.

9. Commit your changes using a descriptive commit message that follows our [commit message conventions](#commit-message-convention).

10. Push your branch to GitHub.

11. In GitHub, send a pull request to `KiiBlockchain:main`.

#### Reviewing a Pull Request

The Kiijs team reserves the right not to accept pull requests from community members who haven't been good citizens of the community. Such behavior includes not following our [code of conduct][coc] and applies within or outside the managed channels.

When you contribute a new feature, the maintenance burden is transferred to the core team. This means that the benefit of the contribution must be compared against the cost of maintaining the feature.

#### Addressing review feedback

If we ask for changes via code reviews then:

1. Make the required updates to the code.

2. Re-run the tests and checks to ensure they are still passing.

3. Create a new commit and push to your GitHub repository (this will update your Pull Request).

#### After your pull request is merged

After your pull request is merged, you can safely delete your branch and pull the changes from the (upstream) repository.

## Coding Rules

To ensure consistency throughout the source code, keep these rules in mind as you are working:

- All code must pass our code quality checks (linters, formatters, etc).
- All features **must be tested** via unit-tests and if applicable integration-tests. Bug fixes also require tests, because the presence of bugs usually indicates insufficient test coverage. Tests help to: 

    1. Prove that your code works correctly, and
    2. Guard against future breaking changes and lower the maintenance cost. 

- All public features **must be documented**.
- All files must include a license header. 
- Keep API compatibility in mind when you change any code under `Kiijs`. Above version `1.0.0`, breaking changes can happen across versions with different left digit. Below version `1.0.0`, they can happen across versions with different middle digit. Reviewers of your pull request will comment on any API compatibility issues.

## Commit Message Convention

Please follow the [Conventional Commits v1.0.0][convcommit]. The commit types must be one of the following:

- **build**: Changes that affect the build system or external dependencies
- **ci**: Changes to our CI configuration files and scripts
- **docs**: Changes to the documentation
- **feat**: A new feature
- **fix**: A bug fix
- **nfunc**: Code that improves some non-functional characteristic, such as performance, security, ...
- **refactor**: A code change that neither fixes a bug nor adds a feature
- **test**: Adding missing tests or correcting existing tests

[coc]: ./CODE_OF_CONDUCT.md
[issues]: https://github.com/KiiChain/kiichain/issues
[new-issue]: https://github.com/KiiChain/kiichain/issues/new/choose
[prs]: https://github.com/KiiChain/kiichain/pulls
[convcommit]: https://www.conventionalcommits.org/en/v1.0.0/
[github]: https://github.com/KiiChain/kiichain
