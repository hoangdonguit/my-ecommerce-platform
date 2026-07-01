# 05 - Security and Limitations

## Implemented security foundations

The system includes:

- Kubernetes Secrets for runtime values
- API key protection at the web-gateway
- Istio mTLS for internal service-to-service traffic
- Istio AuthorizationPolicy for internal API access control
- NetworkPolicy for selected namespace/service traffic
- restricted admin exposure
- health probes and workload resource boundaries
- secret hygiene in scripts and connector bootstrap jobs

## Important limitation

This project does not claim complete Zero Trust or production readiness. It implements a Zero-Trust-oriented security foundation for an academic cloud-native prototype.

## Current limitations

- The cluster is a lab/VM environment.
- PostgreSQL is still single-primary.
- Kafka, Redis, MongoDB, ClickHouse, and observability components do not have full production HA.
- Payment is simulated/COD.
- Notification is in-app/log based.
- Kubernetes Secrets are used instead of Vault or a cloud secret manager.
- DLQ/replay exists for selected paths but is not standardized across the full pipeline.
- Soak test is 60 minutes, not multi-day.
- Backup/restore and disaster recovery need more production-grade validation.
