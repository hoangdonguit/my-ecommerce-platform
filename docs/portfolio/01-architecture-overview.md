# 01 - Architecture Overview

## Purpose

This project builds a cloud-native e-commerce order-processing platform to evaluate how microservices, event-driven communication, Kubernetes, GitOps, observability, and resilience testing work together in a distributed system.

The system focuses on the order workflow rather than simple CRUD operations.

## Why microservices?

The order-processing workflow is separated into business domains:

- order-service
- inventory-service
- payment-service
- notification-service
- read-model-service
- web-gateway
- ecommerce-dashboard

Each service can be deployed, scaled, observed, and evolved independently.

## Why Kafka?

Kafka is used as the event backbone for asynchronous communication between services. It helps decouple services, buffer load spikes, and make consumer lag observable.

Kafka is not the source of truth. PostgreSQL remains the transactional source of truth for business data.

## Why Kubernetes and GitOps?

Kubernetes/K3s provides workload orchestration, service discovery, self-healing, autoscaling, and deployment isolation. ArgoCD GitOps keeps the runtime cluster aligned with manifests stored in Git.

## Lab environment

The final lab cluster uses three K3s nodes:

- vm1-gateway: 192.168.100.27, 4 CPU, about 8GB RAM
- vm2-mesh: 192.168.100.240, 4 CPU, about 8GB RAM
- vm3-gitops: 192.168.100.24, 4 CPU, about 8GB RAM

The DoneGitOps workstation is outside the cluster and is used for kubectl access, Git operations, k6 benchmark execution, and dashboard access.
