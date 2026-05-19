# PostgreSQL Backup / Restore Runbook

## Mục tiêu

PostgreSQL là source of truth cho dữ liệu giao dịch của hệ thống. Vì vậy cần có quy trình backup và kiểm tra restore cơ bản để chứng minh hệ thống có phương án phục hồi dữ liệu.

## Thành phần được backup

Các database chính:

- order_db
- inventory_db
- payment_db
- notification_db

## Backup

Chạy lệnh:

    ./scripts/backup/postgres-backup.sh

Kết quả backup được lưu ở:

    backups/postgres/<timestamp>/*.dump

Định dạng dump sử dụng pg_dump -Fc, phù hợp để restore bằng pg_restore.

## Restore check an toàn

Để kiểm tra file backup có restore được hay không, chạy:

    ./scripts/backup/postgres-restore-check.sh backups/postgres/<timestamp>/order_db.dump

Script sẽ:

1. Tạo database tạm.
2. Restore file dump vào database tạm.
3. Liệt kê bảng và số dòng đã restore.
4. Xóa database tạm sau khi kiểm tra.

Script không ghi đè hoặc xóa database thật.

## Ý nghĩa

Quy trình này giúp chứng minh hệ thống không chỉ chạy được, mà còn có bước vận hành cơ bản để bảo vệ dữ liệu giao dịch.
