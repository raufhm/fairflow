# FairFlow Kubernetes Makefile
.PHONY: help build deploy status logs clean

# Variables
SERVICES := auth group member assignment webhook analytics
NAMESPACE := fairflow

# Default target
help:
	@echo "FairFlow Kubernetes Commands"
	@echo "============================="
	@echo ""
	@echo "  make build       - Build all Docker images"
	@echo "  make deploy      - Deploy all services to Kubernetes"
	@echo "  make status      - Show pods, services, and ingress status"
	@echo "  make logs        - Tail logs (use: make logs SERVICE=auth)"
	@echo "  make restart     - Restart service (use: make restart SERVICE=auth)"
	@echo "  make clean       - Delete all resources"
	@echo ""

# Build all Docker images
build:
	@echo "Building Docker images..."
	@for service in $(SERVICES); do \
		docker build -t fairflow/$$service-service:latest -f services/$$service/Dockerfile . ; \
	done
	@echo "✓ Build complete!"

# Deploy everything to Kubernetes
deploy:
	@echo "Deploying to Kubernetes..."
	@kubectl apply -f k8s/namespace.yaml
	@kubectl apply -f k8s/configmap.yaml
	@kubectl apply -f k8s/secret.yaml
	@kubectl apply -f k8s/
	@echo "✓ Deploy complete! Check status: make status"

# Show status
status:
	@kubectl get pods,svc,ingress -n $(NAMESPACE)

# View logs
logs:
ifndef SERVICE
	@echo "Error: SERVICE not specified"
	@echo "Usage: make logs SERVICE=<name>"
	@echo "Example: make logs SERVICE=auth"
	@exit 1
endif
	@kubectl logs -f -n $(NAMESPACE) -l app=$(SERVICE)-service --tail=50

# Restart service
restart:
ifndef SERVICE
	@echo "Error: SERVICE not specified"
	@echo "Usage: make restart SERVICE=<name>"
	@echo "Example: make restart SERVICE=auth"
	@exit 1
endif
	@kubectl rollout restart deployment/$(SERVICE)-service -n $(NAMESPACE)

# Clean up
clean:
	@kubectl delete namespace $(NAMESPACE)
	@echo "✓ Cleanup complete!"
