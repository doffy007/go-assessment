# Inventory Service

## Fungsi General
- Mengelola stok produk secara real-time.
- Step kedua di Saga flow.
- Menjadi hub untuk komunikasi antar service terkait stock.

## Tanggung Jawab
- Listen event `ORDER_CREATED` dari Event Bus / Queue → reserve stock.
- Publish event:
  - `STOCK_RESERVED` → jika sukses
  - `STOCK_FAILED` → jika gagal → trigger rollback di Order Service
- Listen event `PAYMENT_FAILED` → release stock (compensating transaction)
- Menggunakan **gRPC** untuk expose metode ReserveStock / ReleaseStock untuk direct call.

## Saga Role
- Step executor & compensating transaction.
- Trigger rollback jika step downstream gagal.

## DB / Event / Komunikasi
- **Database:** Redis / PostgreSQL
- **Event Bus / Queue:** Kafka / RabbitMQ (distributed log)
- **gRPC:** Expose ReserveStock / ReleaseStock untuk sinkronisasi.
- **Distributed Log:** Semua event dicatat → replay & recovery.
