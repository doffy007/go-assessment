# E-Commerce Core Services – General Overview

## 1. Order Service
- Titik awal transaksi e-commerce.
- Mengelola lifecycle order (`PENDING`, `CONFIRMED`, `FAILED`).
- Memulai Saga Pattern.
- Mengkoordinasikan rollback jika terjadi kegagalan di service lain.

## 2. Inventory Service
- Mengelola stok produk secara real-time.
- Step kedua dalam transaksi Saga.
- Menyediakan kompensasi (release stock) jika pembayaran gagal.

## 3. Payment Service
- Memproses pembayaran order.
- Step ketiga dalam transaksi Saga.
- Menyediakan kompensasi (refund) jika order dibatalkan atau gagal.

## 4. Notification Service
- Mengirim notifikasi ke user ketika order berhasil atau gagal.
- Step terakhir dalam Saga flow.
- Read-only, tidak ada kompensasi, hanya side-effect (notif).

## 5. Event Bus / Distributed Log
- Backbone komunikasi antar service.
- Menyimpan semua event untuk audit, replay, dan recovery.
- Mendukung penerapan Saga Pattern berbasis event-driven.
