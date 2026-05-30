# Payment Transactional Outbox Smoke Test

## Thoi diem

- Test type: single-order smoke test
- Feature: Payment Transactional Outbox
- Payment image: payment-outbox-20260530171950
- Smoke order id: 8160bb8b-5264-444c-96dc-2b79691f7b08

## Muc tieu

Xac nhan payment-service khong con publish truc tiep payment.completed tu business logic.

Luong moi:

payments status terminal
-> DB trigger insert payment_outbox_events
-> payment outbox worker publish Kafka
-> order-service saga monitor consume payment.completed
-> order COMPLETED

## Ket qua

Order:

- ID: 8160bb8b-5264-444c-96dc-2b79691f7b08
- Final status: COMPLETED

Payment:

- Status: COMPLETED
- Transaction ID: cod_03f562be-cc01-4101-80e6-16a662a1e80f
- Paid at: 2026-05-30 10:30:27

Payment outbox:

- Event type: payment.completed
- Message key: 8160bb8b-5264-444c-96dc-2b79691f7b08
- Status: PUBLISHED
- Attempts: 1

Kafka:

- payment-service-group: lag 0
- notification-service-group: lag 0
- order-service-saga-monitor: lag 0
- read-model-service-group: lag 0

## Ket luan

Payment Transactional Outbox hoat dong dung.

Phase 1 dat yeu cau correctness. Buoc tiep theo la chay lai benchmark 20 RPS de so sanh voi baseline truoc do.
