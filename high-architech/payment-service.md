# Payment Service

## Fungsi General
- Memproses pembayaran order.
- Step ketiga di Saga flow.
- Menjadi hub untuk komunikasi terkait transaksi finansial.

## Tanggung Jawab
- Listen event `STOCK_RESERVED` dari Event Bus / Queue → charge user via payment gateway.
- Publish event:
  - `PAYMENT_SUCCESS` → Order Service update `CONFIRMED`
  - `PAYMENT_FAILED` → trigger rollback (Order cancel + Inventory release)
- Listen rollback dari Order Service → refund jika sudah tercharge.
- Menggunakan **gRPC** untuk expose ChargePayment / RefundPayment untuk direct call.

## Saga Role
- Step executor & compensating transaction.
- Memastikan rollback otomatis jika step sebelumnya berhasil tapi step b
