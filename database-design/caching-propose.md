# Proposal Caching untuk User Feed

## Tujuan
Meningkatkan performa baca dan mengurangi beban database untuk query user feed dengan menambahkan lapisan caching menggunakan Redis di atas SQL view.

---

## Arsitektur Saat Ini
- **Database**: PostgreSQL  
- **Query User Feed**: Join kompleks antara posts, comments, reactions, dan user relationships.  
- **Masalah**: Beban baca yang tinggi dapat memperlambat respon database.

---

## Strategi Caching yang Diusulkan

1. **SQL View**  
   - Buat SQL view yang menggabungkan semua data yang dibutuhkan untuk user feed.  
   - Memastikan logika data tersentralisasi dan konsisten.  

2. **Redis Cache**  
   - Simpan hasil SQL view per user di Redis.  
   - Gunakan format key: `user_feed:{user_id}`.  
   - Set TTL yang sesuai (misal 5–10 menit) untuk mencegah data kadaluarsa.  

3. **Invalidasi Cache**  
   - Setiap ada perubahan pada tabel dasar (posts, comments, reactions, follows):  
     - Hapus atau refresh key Redis terkait.  
     - Bisa dipicu oleh:
       - Event di aplikasi (setelah insert/update/delete).  
       - Cron job untuk refresh periodik jika diperlukan.  

4. **Integrasi API**  
   - Endpoint API fetch feed:
     1. Cek terlebih dahulu di Redis cache.  
     2. Jika tidak ada, query SQL view, simpan hasil ke Redis, lalu kembalikan ke client.  

---

## Manfaat
- Respon lebih cepat untuk permintaan user feed.  
- Mengurangi beban PostgreSQL untuk query baca yang berat.  
- Logika feed tersentralisasi di SQL view sehingga data konsisten.  

---

## Pertimbangan
- Perlu memastikan invalidasi cache dilakukan dengan benar agar tidak menyajikan data kadaluarsa.  
- Perlu memonitor penggunaan memori Redis.  
- Untuk trafik tinggi, pertimbangkan sharding cache per user atau menggunakan Redis cluster.  

---

## Peningkatan di Masa Depan
- Implementasi pub/sub atau event bus untuk push update feed secara near real-time.  
- Gunakan partial update untuk perubahan incremental agar tidak selalu refresh full cache.  
- Pantau metrik performa untuk menyesuaikan TTL dan strategi caching.
