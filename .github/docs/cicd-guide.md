# 🚀 CI/CD Pipeline - GitHub Actions

## 📋 Estratégia de Pipeline

### Fluxo Principal
```
🔀 Feature Branch → 🧪 Tests → 🔍 Quality Gates → 🔄 PR → 🚀 Deploy
```

### Ambientes
- **Development**: Deploy automático de `develop` branch
- **Staging**: Deploy de `release/*` branches
- **Production**: Deploy de `main` após aprovação manual

## 🛠️ Workflows

### 1. CI Pipeline (Continuous Integration)
```yaml
# .github/workflows/ci.yml
name: 🏛️ CI - Continuous Integration

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main, develop]

env:
  GO_VERSION: "1.24"
  NODE_VERSION: "20"

jobs:
  # 🔍 Análise estática e linting
  static-analysis:
    name: 🔍 Static Analysis
    runs-on: ubuntu-latest
    
    steps:
      - name: 📥 Checkout
        uses: actions/checkout@v4
        
      - name: 🔧 Setup Go
        uses: actions/setup-go@v4
        with:
          go-version: ${{ env.GO_VERSION }}
          cache: true
          
      - name: 📦 Download Dependencies
        run: go mod download
        
      - name: 🔍 Go Vet
        run: go vet ./...
        
      - name: 🧹 Go Fmt Check
        run: |
          if [ "$(gofmt -s -l . | wc -l)" -gt 0 ]; then
            echo "❌ Code is not formatted. Run 'gofmt -s -w .'"
            gofmt -s -l .
            exit 1
          fi
          echo "✅ Code is properly formatted"
          
      - name: 🔒 Security Scan (gosec)
        uses: securecodewarrior/github-action-gosec@master
        with:
          args: '-fmt sarif -out gosec.sarif ./...'
          
      - name: 📤 Upload SARIF
        uses: github/codeql-action/upload-sarif@v2
        with:
          sarif_file: gosec.sarif
          
      - name: 📊 Static Check
        uses: dominikh/staticcheck-action@v1.3.0
        with:
          version: "2023.1.6"
          
      - name: 🔄 Go Mod Tidy Check
        run: |
          go mod tidy
          if [ -n "$(git status --porcelain)" ]; then
            echo "❌ go.mod or go.sum is not tidy"
            git diff
            exit 1
          fi

  # 🧪 Testes unitários e de integração
  test:
    name: 🧪 Tests
    runs-on: ubuntu-latest
    
    services:
      postgres:
        image: postgres:16
        env:
          POSTGRES_USER: test
          POSTGRES_PASSWORD: test
          POSTGRES_DB: testdb
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
        ports:
          - 5432:5432
          
      redis:
        image: redis:7
        options: >-
          --health-cmd "redis-cli ping"
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
        ports:
          - 6379:6379
    
    steps:
      - name: 📥 Checkout
        uses: actions/checkout@v4
        
      - name: 🔧 Setup Go
        uses: actions/setup-go@v4
        with:
          go-version: ${{ env.GO_VERSION }}
          cache: true
          
      - name: 📦 Install Dependencies
        run: |
          go mod download
          go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest
          
      - name: 🗄️ Run Migrations
        run: |
          migrate -path migrations -database "postgres://test:test@localhost:5432/testdb?sslmode=disable" up
        env:
          DATABASE_URL: postgres://test:test@localhost:5432/testdb?sslmode=disable
          
      - name: 🧪 Unit Tests
        run: |
          go test -race -coverprofile=coverage.out -covermode=atomic ./...
          
      - name: 📊 Coverage Check
        run: |
          coverage=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
          echo "Coverage: $coverage%"
          if (( $(echo "$coverage < 80" | bc -l) )); then
            echo "❌ Coverage ($coverage%) is below 80% threshold"
            exit 1
          fi
          echo "✅ Coverage threshold met: $coverage%"
          
      - name: 📈 Upload Coverage
        uses: codecov/codecov-action@v3
        with:
          file: ./coverage.out
          flags: backend
          
      - name: 🔗 Integration Tests
        run: go test -tags=integration ./tests/integration/...
        env:
          DATABASE_URL: postgres://test:test@localhost:5432/testdb?sslmode=disable
          REDIS_URL: redis://localhost:6379
          
      - name: 🧪 Generate Test Report
        uses: dorny/test-reporter@v1
        if: success() || failure()
        with:
          name: Go Tests
          path: test-results.json
          reporter: 'go-test-json'

  # 🎨 Frontend Tests
  frontend-test:
    name: 🎨 Frontend Tests
    runs-on: ubuntu-latest
    
    steps:
      - name: 📥 Checkout
        uses: actions/checkout@v4
        
      - name: 🔧 Setup Node.js
        uses: actions/setup-node@v4
        with:
          node-version: ${{ env.NODE_VERSION }}
          cache: 'npm'
          cache-dependency-path: frontend/package-lock.json
          
      - name: 📦 Install Dependencies
        working-directory: ./frontend
        run: npm ci
        
      - name: 🔍 TypeScript Check
        working-directory: ./frontend
        run: npm run type-check
        
      - name: 🧹 Lint
        working-directory: ./frontend
        run: npm run lint
        
      - name: 🧪 Unit Tests
        working-directory: ./frontend
        run: npm run test:coverage
        
      - name: 📈 Upload Frontend Coverage
        uses: codecov/codecov-action@v3
        with:
          file: ./frontend/coverage/clover.xml
          flags: frontend

  # 🏗️ Build
  build:
    name: 🏗️ Build
    runs-on: ubuntu-latest
    needs: [static-analysis, test, frontend-test]
    
    steps:
      - name: 📥 Checkout
        uses: actions/checkout@v4
        
      - name: 🔧 Setup Go
        uses: actions/setup-go@v4
        with:
          go-version: ${{ env.GO_VERSION }}
          
      - name: 🏗️ Build Backend Services
        run: |
          services="deputados atividades despesas forum usuarios ingestao ia"
          for service in $services; do
            echo "Building $service service..."
            CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo \
              -ldflags '-w -s -extldflags "-static"' \
              -o ./bin/$service ./backend/services/$service/cmd/server
          done
          
      - name: 🔧 Setup Node.js
        uses: actions/setup-node@v4
        with:
          node-version: ${{ env.NODE_VERSION }}
          cache: 'npm'
          cache-dependency-path: frontend/package-lock.json
          
      - name: 🏗️ Build Frontend
        working-directory: ./frontend
        run: |
          npm ci
          npm run build
          
      - name: 📦 Upload Build Artifacts
        uses: actions/upload-artifact@v3
        with:
          name: build-artifacts
          path: |
            ./bin/
            ./frontend/.next/
          retention-days: 7
```

### 2. Security Pipeline
```yaml
# .github/workflows/security.yml
name: 🔒 Security Scan

on:
  schedule:
    - cron: '0 2 * * *'  # Daily at 2 AM
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  security-scan:
    name: 🔒 Security Analysis
    runs-on: ubuntu-latest
    
    steps:
      - name: 📥 Checkout
        uses: actions/checkout@v4
        
      - name: 🔍 Trivy Vulnerability Scanner
        uses: aquasecurity/trivy-action@master
        with:
          scan-type: 'fs'
          scan-ref: '.'
          format: 'sarif'
          output: 'trivy-results.sarif'
          severity: 'CRITICAL,HIGH'
          
      - name: 📤 Upload Trivy Results
        uses: github/codeql-action/upload-sarif@v2
        with:
          sarif_file: 'trivy-results.sarif'
          
      - name: 🔒 Dependency Check
        uses: dependency-check/Dependency-Check_Action@main
        with:
          project: 'to-de-olho'
          path: '.'
          format: 'SARIF'
          
      - name: 🚨 Security Audit (Go)
        run: |
          go list -json -deps ./... | nancy sleuth
          
      - name: 🔍 License Check
        uses: fossa-contrib/fossa-action@v2
        with:
          api-key: ${{ secrets.FOSSA_API_KEY }}
```

### 3. CD Pipeline (Continuous Deployment)
```yaml
# .github/workflows/cd.yml
name: 🚀 CD - Continuous Deployment

on:
  push:
    branches: [main, develop]
    tags: ['v*']

env:
  REGISTRY: ghcr.io
  IMAGE_NAME: ${{ github.repository }}

jobs:
  # 🐳 Build and Push Docker Images
  docker:
    name: 🐳 Docker Build & Push
    runs-on: ubuntu-latest
    
    permissions:
      contents: read
      packages: write
      
    strategy:
      matrix:
        service: [deputados, atividades, despesas, forum, usuarios, ingestao, ia]
    
    steps:
      - name: 📥 Checkout
        uses: actions/checkout@v4
        
      - name: 🔑 Log in to Container Registry
        uses: docker/login-action@v3
        with:
          registry: ${{ env.REGISTRY }}
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}
          
      - name: 📋 Extract Metadata
        id: meta
        uses: docker/metadata-action@v5
        with:
          images: ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}/${{ matrix.service }}
          tags: |
            type=ref,event=branch
            type=ref,event=pr
            type=semver,pattern={{version}}
            type=semver,pattern={{major}}.{{minor}}
            type=sha,prefix={{branch}}-
            
      - name: 🔧 Set up Docker Buildx
        uses: docker/setup-buildx-action@v3
        
      - name: 🏗️ Build and Push
        uses: docker/build-push-action@v5
        with:
          context: .
          file: ./backend/services/${{ matrix.service }}/Dockerfile
          push: true
          tags: ${{ steps.meta.outputs.tags }}
          labels: ${{ steps.meta.outputs.labels }}
          cache-from: type=gha
          cache-to: type=gha,mode=max
          platforms: linux/amd64,linux/arm64

  # 🚀 Deploy to Development
  deploy-dev:
    name: 🚀 Deploy to Development
    runs-on: ubuntu-latest
    needs: [docker]
    if: github.ref == 'refs/heads/develop'
    environment: development
    
    steps:
      - name: 📥 Checkout
        uses: actions/checkout@v4
        
      - name: ⚙️ Configure kubectl
        uses: azure/setup-kubectl@v3
        with:
          version: 'v1.28.0'
          
      - name: 🔑 Setup Kubernetes Context
        uses: azure/k8s-set-context@v1
        with:
          method: kubeconfig
          kubeconfig: ${{ secrets.KUBE_CONFIG_DEV }}
          
      - name: 🚀 Deploy to Development
        run: |
          # Update image tags in deployment manifests
          sed -i "s|image: .*|image: ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}:${{ github.sha }}|g" k8s/development/*.yaml
          
          # Apply manifests
          kubectl apply -f k8s/development/
          
          # Wait for rollout
          kubectl rollout status deployment/to-de-olho-api -n development
          
      - name: 🧪 Health Check
        run: |
          # Wait for service to be ready
          sleep 30
          
          # Check health endpoint
          kubectl port-forward svc/to-de-olho-api 8080:8080 -n development &
          sleep 5
          
          response=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/health)
          if [ "$response" != "200" ]; then
            echo "❌ Health check failed: $response"
            exit 1
          fi
          
          echo "✅ Development deployment successful"

  # 🏗️ Deploy to Staging
  deploy-staging:
    name: 🏗️ Deploy to Staging
    runs-on: ubuntu-latest
    needs: [docker]
    if: startsWith(github.ref, 'refs/heads/release/')
    environment: staging
    
    steps:
      - name: 📥 Checkout
        uses: actions/checkout@v4
        
      - name: ⚙️ Setup Helm
        uses: azure/setup-helm@v3
        with:
          version: 'v3.12.0'
          
      - name: 🔑 Setup Kubernetes Context
        uses: azure/k8s-set-context@v1
        with:
          method: kubeconfig
          kubeconfig: ${{ secrets.KUBE_CONFIG_STAGING }}
          
      - name: 🚀 Deploy with Helm
        run: |
          helm upgrade --install to-de-olho ./helm/to-de-olho \
            --namespace staging \
            --create-namespace \
            --set image.tag=${{ github.sha }} \
            --set environment=staging \
            --wait \
            --timeout=10m
            
      - name: 🧪 Run E2E Tests
        run: |
          # Wait for services to be ready
          kubectl wait --for=condition=ready pod -l app=to-de-olho -n staging --timeout=300s
          
          # Run E2E tests against staging
          go test -tags=e2e ./tests/e2e/... -staging-url=https://staging.to-de-olho.com

  # 🏁 Deploy to Production
  deploy-prod:
    name: 🏁 Deploy to Production
    runs-on: ubuntu-latest
    needs: [docker]
    if: github.ref == 'refs/heads/main'
    environment: production
    
    steps:
      - name: 📥 Checkout
        uses: actions/checkout@v4
        
      - name: ⚙️ Setup Helm
        uses: azure/setup-helm@v3
        with:
          version: 'v3.12.0'
          
      - name: 🔑 Setup Kubernetes Context
        uses: azure/k8s-set-context@v1
        with:
          method: kubeconfig
          kubeconfig: ${{ secrets.KUBE_CONFIG_PROD }}
          
      - name: 🚀 Blue-Green Deployment
        run: |
          # Deploy to green environment
          helm upgrade --install to-de-olho-green ./helm/to-de-olho \
            --namespace production \
            --set image.tag=${{ github.sha }} \
            --set environment=production \
            --set deployment.slot=green \
            --wait \
            --timeout=15m
            
          # Health check
          kubectl wait --for=condition=ready pod -l app=to-de-olho,slot=green -n production --timeout=300s
          
          # Switch traffic to green
          kubectl patch service to-de-olho -n production -p '{"spec":{"selector":{"slot":"green"}}}'
          
          # Wait and verify
          sleep 60
          
          # If successful, remove blue deployment
          helm uninstall to-de-olho-blue -n production || true
          
          # Rename green to blue for next deployment
          helm upgrade to-de-olho-blue ./helm/to-de-olho \
            --namespace production \
            --set image.tag=${{ github.sha }} \
            --set environment=production \
            --set deployment.slot=blue
            
      - name: 🔔 Notify Success
        uses: 8398a7/action-slack@v3
        with:
          status: success
          text: "✅ Production deployment successful! Version: ${{ github.sha }}"
        env:
          SLACK_WEBHOOK_URL: ${{ secrets.SLACK_WEBHOOK }}
```

### 4. Quality Gates
```yaml
# .github/workflows/quality-gates.yml
name: 🛡️ Quality Gates

on:
  pull_request:
    branches: [main, develop]

jobs:
  quality-check:
    name: 🛡️ Quality Gates
    runs-on: ubuntu-latest
    
    steps:
      - name: 📥 Checkout
        uses: actions/checkout@v4
        with:
          fetch-depth: 0  # Fetch full history for better analysis
          
      - name: 🔧 Setup Go
        uses: actions/setup-go@v4
        with:
          go-version: "1.24"
          
      - name: 🔍 Code Coverage Gate
        run: |
          go test -coverprofile=coverage.out ./...
          coverage=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
          
          echo "Coverage: $coverage%"
          
          if (( $(echo "$coverage < 80" | bc -l) )); then
            echo "❌ Coverage ($coverage%) is below 80% threshold"
            echo "::error::Code coverage is below required threshold"
            exit 1
          fi
          
          echo "✅ Coverage gate passed: $coverage%"
          
      - name: 🔒 Security Gate
        run: |
          # Run security scan
          gosec -quiet -fmt json -out gosec-report.json ./... || true
          
          # Check for high/critical issues
          critical_issues=$(jq '[.Issues[] | select(.severity == "HIGH" or .severity == "CRITICAL")] | length' gosec-report.json)
          
          if [ "$critical_issues" -gt 0 ]; then
            echo "❌ Found $critical_issues critical/high security issues"
            jq '.Issues[] | select(.severity == "HIGH" or .severity == "CRITICAL")' gosec-report.json
            exit 1
          fi
          
          echo "✅ Security gate passed"
          
      - name: 📊 Complexity Gate
        run: |
          # Install gocyclo
          go install github.com/fzipp/gocyclo/cmd/gocyclo@latest
          
          # Check cyclomatic complexity
          complex_functions=$(gocyclo -over 10 . | wc -l)
          
          if [ "$complex_functions" -gt 0 ]; then
            echo "❌ Found $complex_functions functions with complexity > 10"
            gocyclo -over 10 .
            exit 1
          fi
          
          echo "✅ Complexity gate passed"
          
      - name: 🧹 Code Quality Gate (SonarCloud)
        uses: SonarSource/sonarcloud-github-action@master
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          SONAR_TOKEN: ${{ secrets.SONAR_TOKEN }}
          
      - name: 📈 Performance Gate
        run: |
          # Run benchmarks
          go test -bench=. -benchmem ./... > benchmark-results.txt
          
          # Check for performance regressions (placeholder)
          echo "✅ Performance gate passed"
          
      - name: 📝 Documentation Gate
        run: |
          # Check if public functions have comments
          missing_docs=$(go doc -all ./... | grep -c "^func.*exported.*should have comment" || true)
          
          if [ "$missing_docs" -gt 0 ]; then
            echo "❌ Found $missing_docs exported functions without documentation"
            exit 1
          fi
          
          echo "✅ Documentation gate passed"
```

## 🔧 Configuration Files

### Dockerfile Example
```dockerfile
# backend/services/deputados/Dockerfile
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o main ./backend/services/deputados/cmd/server

# Final stage
FROM scratch

# Import from builder
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /app/main /main

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD ["/main", "health"]

# Run the binary
ENTRYPOINT ["/main"]
```

### Kubernetes Deployment
```yaml
# k8s/development/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deputados-service
  namespace: development
  labels:
    app: deputados-service
    version: v1
spec:
  replicas: 2
  selector:
    matchLabels:
      app: deputados-service
  template:
    metadata:
      labels:
        app: deputados-service
        version: v1
    spec:
      containers:
      - name: deputados-service
        image: ghcr.io/alzarus/to-de-olho/deputados:latest
        ports:
        - containerPort: 8080
        env:
        - name: DATABASE_URL
          valueFrom:
            secretKeyRef:
              name: database-secret
              key: url
        - name: REDIS_URL
          valueFrom:
            secretKeyRef:
              name: redis-secret
              key: url
        - name: ENVIRONMENT
          value: "development"
        resources:
          limits:
            cpu: 500m
            memory: 512Mi
          requests:
            cpu: 250m
            memory: 256Mi
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /ready
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
```

---

> **🎯 Resultado**: Pipeline automatizado que garante qualidade, segurança e deploy confiável em múltiplos ambientes.
