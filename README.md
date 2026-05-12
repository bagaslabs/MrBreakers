# MrBreakers 🚀

**MrBreakers** adalah alat pengujian beban (Load Testing) berbasis Go yang dirancang untuk melakukan pengujian stabilitas jaringan dan simulasi beban koneksi tinggi pada server target melalui jalur proxy SOCKS5.

Program ini mendukung protokol **TCP** dan **UDP (UDP Associate)** dengan fitur rekoneksi otomatis untuk menguji ketahanan server dalam menangani handshake proxy secara massal.

## 🌟 Fitur Utama
- **Support SOCKS5**: Mendukung penuh proxy SOCKS5 dengan otentikasi username & password.
- **UDP Associate**: Implementasi manual handshake UDP SOCKS5 untuk menembus batasan jaringan.
- **High Concurrency**: Mendukung ribuan worker (goroutines) secara bersamaan.
- **Real-time Stats**: Menampilkan statistik uptime, paket sukses, paket gagal, dan jumlah koneksi aktif.
- **Connection-Only Mode**: Mode khusus untuk menguji stress pada bagian handshake server tanpa membebani bandwidth secara berlebih (Rekoneksi loop).

## 🛠️ Instalasi
1. Pastikan Anda sudah menginstal [Go](https://golang.org/dl/).
2. Clone repository ini:
   ```bash
   git clone https://github.com/bagaslabs/MrBreakers.git
   cd MrBreakers
   ```
3. Jalankan program:
   ```bash
   go run cmd/main.go
   ```

## ⚙️ Konfigurasi
Edit file `config.json` untuk menyesuaikan target pengujian:
- `mode`: "tcp" atau "udp"
- `host`: IP target server
- `port`: Port target server
- `connections`: Jumlah koneksi simultan (Worker)
- `interval_ms`: Jeda waktu sebelum melakukan rekoneksi

## ⚠️ Disclaimer
Alat ini dibuat untuk tujuan **Edukasi** dan **Internal Stress Testing** saja. Segala bentuk penyalahgunaan terhadap infrastruktur pihak lain tanpa izin adalah tanggung jawab penuh pengguna. Penulis tidak bertanggung jawab atas kerugian yang ditimbulkan.

---
Developed with ❤️ by [bagaslabs](https://github.com/bagaslabs)
