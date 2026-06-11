# Phase 6.4 - PostgreSQL Backup and Restore Proof

## 1. Purpose

This document records the Phase 6.4 PostgreSQL backup and restore verification.

The goal is to prove that the platform can create logical PostgreSQL backups and restore them into temporary databases without overwriting or deleting production databases.

## 2. Scope

Databases verified:

- order_db
- inventory_db
- payment_db
- notification_db

Backup type:

- Logical backup
- pg_dump custom format
- Restore validation with pg_restore

## 3. Runtime Context

Git revision at proof time:

- Commit: 409fa06
- Branch state: HEAD equals origin/main before proof

PostgreSQL runtime:

- Pod: postgresql-0
- Namespace: db
- Tooling: psql, pg_dump, pg_restore available inside the PostgreSQL pod
- PostgreSQL version: 18.4
- PVC: data-postgresql-0
- PVC size: 30Gi
- StorageClass: local-path

## 4. Backup Location

Raw backup dumps were stored locally under:

- .local-notes/backup/phase6-postgres-backup-restore-20260611211958

These raw dump files are not committed to Git.

Only this evidence document is committed.

## 5. Database Size Snapshot

- inventory_db	56 MB
- notification_db	30 MB
- order_db	84 MB
- payment_db	73 MB
- postgres	7678 kB

## 6. Backup Files and Sizes

- order_db: 7180563 bytes
- inventory_db: 7352340 bytes
- payment_db: 12887754 bytes
- notification_db: 2420477 bytes

## 7. Restore Verification Summary

- restore_check_order_db_20260611212008	user_tables	3
- restore_check_order_db_20260611212008	estimated_rows	102633
- restore_check_inventory_db_20260611212018	user_tables	4
- restore_check_inventory_db_20260611212018	estimated_rows	102636
- restore_check_payment_db_20260611212025	user_tables	3
- restore_check_payment_db_20260611212025	estimated_rows	102523
- restore_check_notification_db_20260611212033	user_tables	1
- restore_check_notification_db_20260611212033	estimated_rows	34211

## 8. Checksum Evidence

SHA256 checksums were generated for all dump files.

- 405a9c6496b7c1c6656ea4f009fcb166ba1f394c1caad9f26217b72b3e1d371c  .local-notes/backup/phase6-postgres-backup-restore-20260611211958/order_db.dump
- 7c25f1a4b868e6e22ecbac46ce8d4084dec69a642e7c3bd3239b3bbcd14cd97d  .local-notes/backup/phase6-postgres-backup-restore-20260611211958/inventory_db.dump
- 4a0e042579f893157e0cbf6b1b656d7ce9e572821285787dfeb204807fb02dc5  .local-notes/backup/phase6-postgres-backup-restore-20260611211958/payment_db.dump
- 553a1ffe8484865dd202a8558d6b9127dd8ed64904eb2cc97dbe50b48da832f0  .local-notes/backup/phase6-postgres-backup-restore-20260611211958/notification_db.dump

## 9. Cleanup Verification

Temporary restore databases were dropped after verification.

Remaining temporary restore databases:

- None

## 10. Observability Check

After backup and restore verification:

- PostgreSQL remained running
- PgBouncer remained running
- Loki and Alloy remained running
- Tempo and OTel Collector remained running

Active ECommerce alerts after proof:

- None

## 11. Safety Notes

This proof did not overwrite any production database.

Restore checks used temporary databases named with the restore_check prefix.

Raw dump files were stored under .local-notes and were not committed.

No secret values were printed or committed.

## 12. Limitations

This is a logical backup and restore proof for the current lab and pre-production prototype.

It does not prove:

- point-in-time recovery
- WAL archive recovery
- cross-cluster disaster recovery
- encrypted offsite backup storage
- scheduled backup automation

These can be added in later production-hardening phases.

## 13. Verdict

Result: PASS.

The platform has a verified PostgreSQL logical backup and restore procedure for the core transactional databases.
