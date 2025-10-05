# Order Service

## Fungsi General
- Titik awal transaksi e-commerce.
- Mengelola lifecycle order (`PENDING`, `CONFIRMED`, `FAILED`).
- Memulai Saga Pattern.
- Menjadi hub untuk komunikasi antar service.

## Tanggung Jawab
- Terima request order dari API Gateway / Client.
- Simpan order di **Order DB**.
- Publish event `ORDER_CREATED` ke **Event Bus / Queue** → Inventory Service.
- Listen event dari Inventory & Payment Service:
  - `STOCK_RESERVED` → lanjut ke Payment Service
  - `STOCK_FAILED` → cancel order (`FAILED`)
  - `PAYMENT_SUCCESS` → konfirmasi order (`CONFIRMED`)
  - `PAYMENT_FAILED` → cancel order (`FAILED`)
- Menggunakan **gRPC** untuk direct call ke Inventory / Payment Service jika diperlukan (sinkron).
- Menyimpan semua event ke **distributed log** untuk audit, replay, dan recovery.

## Saga Role
- Starter & coordinator untuk compensating transaction jika step berikutnya gagal.
- Mengendalikan rollback order saat ada step gagal di downstream service.

## DB / Event / Komunikasi
- **Database:** PostgreSQL / MySQL
- **Event Bus / Queue:** Kafka / RabbitMQ (distributed log)
- **gRPC:** Direct call ke Inventory / Payment Service untuk sinkronisasi (opsional)
- **Distributed Log:** Menyimpan semua event order → replay & recovery.
