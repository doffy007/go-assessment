# Event Bus / Distributed Log

## Fungsi General
- Backbone komunikasi antar microservices.
- Menyimpan semua event transaksi → mendukung Saga Pattern.
- Memungkinkan replay, audit, dan recovery.

## Tugas Utama
- Publish / subscribe event:
  - `ORDER_CREATED` → `STOCK_RESERVED` / `STOCK_FAILED` → `PAYMENT_SUCCESS` / `PAYMENT_FAILED` → `ORDER_CONFIRMED` / `ORDER_CANCELLED`
- Audit trail & replay / recovery

## Teknologi
- Kafka atau RabbitMQ
- Durability & high throughput
