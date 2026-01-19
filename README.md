# Buddy CLI

## Setup

```bash
# Create environment files
cp env_example .env.my
cp env_example .env.sg

# Edit .env.my and .env.sg with your credentials
```

## Build

```bash
make deps        # Install dependencies
make build       # Build both binaries
make deploy      # Build and install to ~/bin
```

## Verification

```bash
./bin/mybuddy --help
./bin/sgbuddy --help
```

## Commands

```bash
make build-my    # Build mybuddy only
make build-sg    # Build sgbuddy only
make lint        # Run linters
make test        # Run tests
make help        # Show all targets
```
