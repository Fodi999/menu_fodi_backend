# CI/CD Integration Guide

This guide helps you integrate the testing infrastructure into your CI/CD pipeline.

## GitHub Actions

### Basic Setup

Create `.github/workflows/test.yml`:

```yaml
name: Tests

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main, develop]

jobs:
  test:
    runs-on: ubuntu-latest
    
    services:
      postgres:
        image: postgres:15-alpine
        env:
          POSTGRES_PASSWORD: test_password
          POSTGRES_USER: test_user
          POSTGRES_DB: test_db
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
        ports:
          - 5432:5432

    steps:
      - uses: actions/checkout@v3
      
      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: 1.24
      
      - name: Download dependencies
        run: go mod download
      
      - name: Run unit tests
        run: make test
      
      - name: Run API tests
        run: make api
      
      - name: Run integration tests
        run: make integration
        env:
          DATABASE_URL: postgres://test_user:test_password@localhost:5432/test_db
      
      - name: Generate coverage
        run: make coverage
      
      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          files: ./coverage.out
```

## GitLab CI

Create `.gitlab-ci.yml`:

```yaml
stages:
  - test

test:unit:
  stage: test
  image: golang:1.24
  script:
    - go mod download
    - make test
  coverage: '/coverage: \d+\.\d+%/'

test:api:
  stage: test
  image: golang:1.24
  script:
    - go mod download
    - make api

test:integration:
  stage: test
  image: golang:1.24
  services:
    - postgres:15-alpine
  variables:
    POSTGRES_PASSWORD: test_password
    POSTGRES_USER: test_user
    POSTGRES_DB: test_db
    DATABASE_URL: postgres://test_user:test_password@postgres:5432/test_db
  script:
    - go mod download
    - make integration

coverage:
  stage: test
  image: golang:1.24
  script:
    - go mod download
    - make coverage
  artifacts:
    paths:
      - coverage.html
```

## Jenkins

Create `Jenkinsfile`:

```groovy
pipeline {
    agent any
    
    environment {
        GO_VERSION = '1.24'
    }
    
    stages {
        stage('Setup') {
            steps {
                sh 'go version'
                sh 'go mod download'
            }
        }
        
        stage('Unit Tests') {
            steps {
                sh 'make test'
            }
        }
        
        stage('API Tests') {
            steps {
                sh 'make api'
            }
        }
        
        stage('Integration Tests') {
            steps {
                sh 'make integration'
            }
        }
        
        stage('Coverage') {
            steps {
                sh 'make coverage'
                publishHTML([
                    reportDir: '.',
                    reportFiles: 'coverage.html',
                    reportName: 'Code Coverage'
                ])
            }
        }
    }
    
    post {
        always {
            junit 'test-results.xml'
        }
        failure {
            echo 'Pipeline failed'
        }
    }
}
```

## Local Development with Pre-commit

Create `.pre-commit-config.yaml`:

```yaml
repos:
  - repo: local
    hooks:
      - id: go-unit-tests
        name: Go Unit Tests
        entry: make test
        language: system
        pass_filenames: false
        stages: [commit]
      
      - id: go-fmt
        name: Go Format
        entry: go fmt ./...
        language: system
        pass_filenames: false
        types: [go]
```

Install:
```bash
pip install pre-commit
pre-commit install
```

## Docker-based Testing

Create `docker-compose.test.yml`:

```yaml
version: '3.8'

services:
  postgres:
    image: postgres:15-alpine
    environment:
      POSTGRES_PASSWORD: test_password
      POSTGRES_USER: test_user
      POSTGRES_DB: test_db
    ports:
      - "5433:5432"
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U test_user"]
      interval: 10s
      timeout: 5s
      retries: 5

  tests:
    build:
      context: .
      dockerfile: Dockerfile.test
    depends_on:
      postgres:
        condition: service_healthy
    environment:
      DATABASE_URL: postgres://test_user:test_password@postgres:5432/test_db
    command: make all
```

Create `Dockerfile.test`:

```dockerfile
FROM golang:1.24-alpine

WORKDIR /app

RUN apk add --no-cache make git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

CMD ["make", "all"]
```

Run tests:
```bash
docker-compose -f docker-compose.test.yml up --build
```

## Coverage Reports

### Local

```bash
make coverage
# Opens coverage.html in browser
```

### Codecov Integration

Add to `.codecov.yml`:

```yaml
coverage:
  precision: 2
  round: down
  range: "80..100"

ignore:
  - "tests"
  - "mocks"

comment:
  layout: "reach,diff,flags,tree"
  behavior: default
```

### Coveralls Integration

```bash
go get github.com/mattn/goveralls
goveralls -coverprofile=coverage.out -service=github
```

## Performance Benchmarks

Create `.github/workflows/benchmark.yml`:

```yaml
name: Benchmarks

on:
  push:
    branches: [main]

jobs:
  benchmark:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: 1.24
      - run: go mod download
      - run: go test ./tests/unit/... -bench=. -benchmem
```

## Best Practices

### ✅ DO
- Run tests on every push
- Fail pipeline if coverage drops below threshold
- Run integration tests with Docker
- Generate coverage reports
- Use matrix builds for multiple Go versions
- Cache dependencies to speed up CI

### ❌ DON'T
- Run integration tests on every branch
- Ignore test failures
- Skip coverage checks
- Leave long-running tests in unit suite
- Test in production environment

## Monitoring

### Coverage Badges

Add to README.md:

```markdown
[![codecov](https://codecov.io/gh/dmitrijfomin/menu-fodifood/branch/main/graph/badge.svg)](https://codecov.io/gh/dmitrijfomin/menu-fodifood)
```

### Status Badges

```markdown
[![Tests](https://github.com/dmitrijfomin/menu-fodifood/workflows/Tests/badge.svg)](https://github.com/dmitrijfomin/menu-fodifood/actions)
```

## Troubleshooting

### Tests timeout in CI

Increase timeout:
```bash
go test ./tests/... -timeout 10m
```

### Docker not available

Use service containers instead of testcontainers in CI.

### Dependencies download fails

Add step:
```yaml
- name: Go cache
  uses: actions/cache@v3
  with:
    path: |
      ~/.cache/go-build
      ~/go/pkg/mod
    key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
```

## Resources

- [GitHub Actions Go Setup](https://github.com/actions/setup-go)
- [GitLab CI/CD](https://docs.gitlab.com/ee/ci/)
- [Jenkins Go Plugin](https://plugins.jenkins.io/go/)
- [Codecov Documentation](https://docs.codecov.com/)
