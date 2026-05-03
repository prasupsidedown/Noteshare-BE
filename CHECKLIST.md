# 🚀 Noteshare Backend - Quick Checklist

Gunakan checklist ini untuk memastikan semua sudah disetup dengan benar sebelum menjalankan backend.

## ✅ Pre-Setup Checklist

- [ ] Go 1.22+ sudah terinstall (`go version`)
- [ ] PostgreSQL sudah terinstall dan berjalan
- [ ] Git sudah terinstall
- [ ] Visual Studio Code atau editor favorit siap

## ✅ Database Setup

- [ ] PostgreSQL service sudah berjalan
- [ ] Database `noteshare` sudah dibuat
- [ ] Username & password sudah benar (default: postgres/postgres)

**Verifikasi:**

```bash
psql -U postgres -d noteshare -c "SELECT version();"
```

## ✅ Cloudinary Setup (WAJIB untuk file upload)

Jika belum punya akun:

1. [ ] Buka https://cloudinary.com/
2. [ ] Sign up dengan email atau GitHub
3. [ ] Verifikasi email
4. [ ] Masuk ke dashboard: https://cloudinary.com/console
5. [ ] Copy credentials:
   - [ ] CLOUDINARY_CLOUD_NAME
   - [ ] CLOUDINARY_API_KEY
   - [ ] CLOUDINARY_API_SECRET

## ✅ Environment Setup

1. [ ] Copy `.env.example` ke `.env`
2. [ ] Isi database credentials:
   - [ ] DB_HOST=localhost
   - [ ] DB_PORT=5432
   - [ ] DB_USER=postgres
   - [ ] DB_PASSWORD=<your password>
   - [ ] DB_NAME=noteshare

3. [ ] Generate JWT secret:

   ```bash
   # Copy-paste hasil ini ke JWT_SECRET di .env
   openssl rand -hex 32
   ```

4. [ ] Isi Cloudinary credentials:
   - [ ] CLOUDINARY_CLOUD_NAME
   - [ ] CLOUDINARY_API_KEY
   - [ ] CLOUDINARY_API_SECRET

5. [ ] (Optional) Setup Anthropic API untuk AI features:
   - [ ] Daftar di https://console.anthropic.com/
   - [ ] Generate API key
   - [ ] Isi ANTHROPIC_API_KEY di .env

## ✅ Project Setup

```bash
# Di folder Noteshare-BE
cd c:\Users\Faris\Documents\kuliah\Noteshare-BE

# Download dependencies
go mod download
go mod tidy

# Verifikasi dependencies
go mod graph
```

- [ ] `go mod download` berhasil
- [ ] Tidak ada error di console

## ✅ Run Backend

```bash
go run main.go
```

Expected output:

```
✅ Database connected successfully
🚀 Noteshare backend running on port 8080
```

- [ ] Backend berjalan tanpa error
- [ ] Database connection success
- [ ] Port 8080 listen dengan baik

## ✅ Test Backend

### Health Check

```bash
curl http://localhost:8080/health
```

- [ ] Response: `{"status":"ok","service":"noteshare-backend",...}`

### Register Test

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test User",
    "email": "test@example.com",
    "password": "password123"
  }'
```

- [ ] Response: 201 Created dengan token
- [ ] Copy token untuk testing

### Get Profile Test

```bash
curl http://localhost:8080/api/v1/auth/me \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"
```

- [ ] Response: 200 OK dengan user data

## ✅ Troubleshooting

Jika ada error, cek:

- [ ] PostgreSQL sudah running?

  ```bash
  psql -U postgres
  ```

- [ ] Database credentials benar di `.env`?
  - Host: localhost
  - Port: 5432
  - User: postgres
  - Password: <sesuai yang diset>
  - Database: noteshare

- [ ] Cloudinary credentials benar?
  - Buka https://cloudinary.com/console
  - Verifikasi ketiga credentials

- [ ] JWT_SECRET sudah diisi? (jangan kosong)

- [ ] Port 8080 tidak sedang dipakai aplikasi lain?

  ```bash
  netstat -ano | findstr :8080
  ```

- [ ] File `.env` sudah dibuat dari `.env.example`?

## ✅ Important Notes

⚠️ **JANGAN:**

- [ ] Commit file `.env` ke git
- [ ] Share JWT_SECRET public
- [ ] Gunakan API key di public repository
- [ ] Set MAX_FILE_SIZE_MB terlalu besar (>100MB)

📝 **INGAT:**

- [ ] Backup credentials Cloudinary & Anthropic
- [ ] Ubah JWT_SECRET sebelum production
- [ ] Monitor Cloudinary usage (ada free tier limit)
- [ ] Backup database secara berkala

---

**Status:** Siap untuk development! 🎉

Jika semua checklist sudah diisi ✅, backend siap dijalankan:

```bash
go run main.go
```

Happy coding! 🚀
