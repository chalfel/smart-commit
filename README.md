# Smart Commit

A Go CLI tool that generates intelligent commit messages using GitHub Copilot CLI.

## Prerequisites

- Go 1.20 or later
- Git installed and configured
- **GitHub Copilot CLI installed (`gh copilot`)** - This is required and the tool will exit with an error if not found

## Installation

### 1. Install GitHub Copilot CLI first (if not already installed)

```bash
gh extension install github/gh-copilot
```

### 2. Install Smart Commit

```bash
# Clone the repository
git clone https://github.com/chalfel/smart-commit.git
cd smart-commit

# Build and install globally
go install
```

## Usage

Simply run:

```bash
smart-commit
```

This will:
1. Add all changes to staging (`git add .`)
2. Generate a commit message using GitHub Copilot
3. Commit the changes with the generated message
4. Push the changes to the remote repository

## How it works

The tool uses GitHub Copilot CLI to analyze your staged changes and generate a contextually relevant commit message.
If GitHub Copilot CLI is not available, it falls back to a basic commit message.

## License

MIT
