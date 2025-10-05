# Notification Service

## Fungsi General
- Mengirim notifikasi ke user saat order berhasil atau gagal.
- Step terakhir di Saga flow, read-only.
- Menjadi hub untuk komunikasi terkait pemberitahuan transaksi.

## Tanggung Jawab
- Listen event dari Event Bus / Queue:
  - `ORDER_CONFIRMED` → send email/SMS/push
  - `ORDER_CANCELLED` → send cancellation notice
- Optional: simpan log notifikasi
- Menggunakan **gRPC** untuk expose SendNotification method jika service lain ingin trigger notifikasi secara sinkron.

## Saga Role
- Step terakhir → tidak ada compensating transaction.
- Read-only, side-effect only.

## DB / Event / Komunikasi
- **Database:** optional Notification DB / log
- **Event Bus / Queue:** subscribe (distributed log)
- **gRPC:** SendNotification method untuk sinkronisasi.
- **Distributed Log:** Semua event tersimpan → replay & audit.
