# Noteshare Backend Setup Guide

## 🚀 Prerequisites

Sebelum menjalankan backend, pastikan Anda sudah memiliki:

- Go 1.22 atau lebih tinggi
- PostgreSQL 12 atau lebih tinggi
- Git

## 📋 Step-by-Step Setup

### 1. Clone Repository

```bash
cd c:\Users\Faris\Documents\kuliah\
git clone <your-repo-url> Noteshare-BE
cd Noteshare-BE
```

### 2. Install Dependencies

```bash
go mod download
go mod tidy
```

### 3. Setup Database PostgreSQL

#### Option A: Menggunakan PostgreSQL lokal (Windows)

```bash
# 1. Install PostgreSQL dari https://www.postgresql.org/download/windows/
# 2. Buka pgAdmin atau command line

# 3. Buat database baru
createdb -U postgres noteshare

# 4. Verifikasi koneksi
psql -U postgres -d noteshare
```

#### Option B: Menggunakan Docker (Recommended)

```bash
# 1. Install Docker dari https://www.docker.com/products/docker-desktop/

# 2. Run PostgreSQL container
docker run --name noteshare-postgres -e POSTGRES_PASSWORD=postgres -e POSTGRES_DB=noteshare -p 5432:5432 -d postgres:16

# 3. Verifikasi
docker ps
```

### 4. Setup Environment Variables

#### Langkah 1: Copy file template

```bash
cp .env.example .env
```

#### Langkah 2: Edit .env file

Buka file `.env` dan isi dengan konfigurasi Anda:

```env
# Server
PORT=8080
ENV=development

# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres        # Sesuaikan dengan password PostgreSQL Anda
DB_NAME=noteshare

# JWT
JWT_SECRET=your-super-secret-key-here
JWT_EXPIRY_HOURS=72

# File Upload
MAX_FILE_SIZE_MB=50

# Cloudinary (WAJIB untuk file upload)
CLOUDINARY_CLOUD_NAME=xxx
CLOUDINARY_API_KEY=xxx
CLOUDINARY_API_SECRET=xxx

# AI (Optional - untuk auto-generate todo lists)
ANTHROPIC_API_KEY=xxx
```

---

## 🔑 Setup Required API Keys & Services

### 1. **Cloudinary Setup** (Required for file uploads)

Cloudinary adalah layanan cloud storage untuk file uploads. Ikuti langkah ini:

1. **Daftar akun:**
   - Buka https://cloudinary.com/
   - Klik "Sign Up for Free"
   - Daftar dengan email atau GitHub

2. **Dapatkan API Credentials:**
   - Masuk ke Dashboard: https://cloudinary.com/console
   - Di halaman dashboard, Anda akan melihat:
     - Cloud Name
     - API Key
     - API Secret
   - Copy ketiga nilai tersebut ke file `.env`

3. **Verifikasi:**
   ```env
   CLOUDINARY_CLOUD_NAME=xxxxxxxxxxxx
   CLOUDINARY_API_KEY=1234567890123456789
   CLOUDINARY_API_SECRET=1234567890123456789abcdefghijklmn
   ```

**Dokumentasi:** https://cloudinary.com/documentation/how_to_integrate_cloudinary

---

### 2. **PostgreSQL Database Setup** (Required)

Sudah dijelaskan di atas. Pastikan:

- Database sudah dibuat
- User & password sudah benar
- Server berjalan di port 5432

Verifikasi koneksi:

```bash
# Windows Command Line atau PowerShell
psql -h localhost -U postgres -d noteshare -c "SELECT version();"
```

---

### 3. **JWT Secret** (Required)

Generate secure JWT secret:

```bash
# Option 1: Gunakan OpenSSL (jika ada di system)
openssl rand -hex 32

# Option 2: Gunakan online generator
# https://www.random.org/strings/?num=1&len=32&digits=on&loweralpha=on&upperalpha=on

# Output akan mirip:
# a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6q7r8s9t0

# Copy ke .env:
JWT_SECRET=a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6q7r8s9t0
```

---

### 4. **Anthropic API** (Optional - untuk AI Todo Generation)

Ini optional. Digunakan untuk fitur auto-generate todo lists dari notes.

1. **Daftar akun:**
   - Buka https://console.anthropic.com/
   - Daftar atau login

2. **Generate API Key:**
   - Di console, buka "API Keys"
   - Klik "Create Key"
   - Copy API key ke .env

3. **Set di .env:**
   ```env
   ANTHROPIC_API_KEY=sk-ant-v0-1234567890abcdefghijklmnopqrst
   ```

**Dokumentasi:** https://docs.anthropic.com/

**Note:** Jika tidak ada API key, fitur AI tidak akan berfungsi tapi backend tetap jalan.

---

## 🏃 Run Backend

### Development Mode

```bash
go run main.go
```

Output yang diharapkan:

```
2025/05/02 10:00:00 ✅ Database connected successfully
2025/05/02 10:00:00 🚀 Noteshare backend running on port 8080
```

### Production Mode

```bash
# Build executable
go build -o noteshare-backend.exe

# Run
./noteshare-backend.exe
```

Atau set ENV variable:

```bash
$env:ENV = "production"
go run main.go
```

---

## ✅ Test Backend

### 1. Health Check

```bash
# Windows PowerShell atau Terminal lainnya
curl http://localhost:8080/health

# Output yang diharapkan:
# {
#   "status": "ok",
#   "service": "noteshare-backend",
#   "version": "1.0.0"
# }
```

### 2. Register User

```bash
curl -X POST http://localhost:8080/api/v1/auth/register `
  -H "Content-Type: application/json" `
  -d '{
    "name": "John Doe",
    "email": "john@example.com",
    "password": "password123"
  }'

# Output yang diharapkan:
# {
#   "success": true,
#   "message": "Registrasi berhasil",
#   "data": {
#     "user": { ... },
#     "token": "eyJhbGciOiJIUzI1NiIs..."
#   }
# }
```

### 3. Login

```bash
curl -X POST http://localhost:8080/api/v1/auth/login `
  -H "Content-Type: application/json" `
  -d '{
    "email": "john@example.com",
    "password": "password123"
  }'
```

### 4. Get Profile (Authenticated)

```bash
curl -X GET http://localhost:8080/api/v1/auth/me `
  -H "Authorization: Bearer YOUR_TOKEN_HERE"
```

---

## 📚 API Endpoints

### Authentication

- `POST /api/v1/auth/register` - Register user
- `POST /api/v1/auth/login` - Login
- `GET /api/v1/auth/me` - Get profile (Auth required)
- `PUT /api/v1/auth/me` - Update profile (Auth required)

### Courses

- `GET /api/v1/courses` - Get public courses
- `POST /api/v1/courses` - Create course (Auth required)
- `GET /api/v1/courses/:id` - Get single course
- `PUT /api/v1/courses/:id` - Update course (Auth required)
- `DELETE /api/v1/courses/:id` - Delete course (Auth required)

### Notes

- `GET /api/v1/notes` - Get public notes
- `GET /api/v1/notes/my` - Get my notes (Auth required)
- `POST /api/v1/notes` - Upload note (Auth required)
- `GET /api/v1/notes/:id` - Get single note
- `GET /api/v1/notes/:id/download` - Download note file
- `PUT /api/v1/notes/:id` - Update note (Auth required)
- `DELETE /api/v1/notes/:id` - Delete note (Auth required)

### Todo Lists

- `GET /api/v1/todos/my` - Get my todo lists (Auth required)
- `POST /api/v1/notes/:id/todos/generate` - AI-generate todos (Auth required)
- `GET /api/v1/notes/:id/todos` - Get todos for note (Auth required)
- `GET /api/v1/todos/:id` - Get single todo list (Auth required)
- `POST /api/v1/todos/:id/items` - Add todo item manually (Auth required)
- `PATCH /api/v1/todos/:id/items` - Toggle todo item (Auth required)
- `DELETE /api/v1/todos/:id` - Delete todo list (Auth required)

---

## 🐛 Troubleshooting

### Error: "Failed to connect to database"

**Solusi:**

1. Verifikasi PostgreSQL sedang berjalan:
   ```bash
   psql -U postgres
   ```
2. Verifikasi credentials di `.env`
3. Pastikan database `noteshare` sudah dibuat

### Error: "Cloudinary not initialized"

**Solusi:**

1. Verifikasi credentials di `.env`:
   - CLOUDINARY_CLOUD_NAME
   - CLOUDINARY_API_KEY
   - CLOUDINARY_API_SECRET
2. Test di console Cloudinary

### Error: "Port 8080 already in use"

**Solusi:**

```bash
# Ganti PORT di .env
PORT=8081

# Atau kill process yang sedang pakai port 8080
# Windows:
netstat -ano | findstr :8080
taskkill /PID <PID> /F
```

---

## 📦 Dependencies

Semua dependencies sudah di-list di `go.mod`:

- `github.com/gin-gonic/gin` - Web framework
- `gorm.io/gorm` - ORM
- `gorm.io/driver/postgres` - PostgreSQL driver
- `github.com/golang-jwt/jwt/v5` - JWT authentication
- `github.com/cloudinary/cloudinary-go/v2` - Cloudinary SDK
- `golang.org/x/crypto` - Password hashing (bcrypt)

Install semua dengan:

```bash
go mod download
```

---

## 🔒 Security Notes

Untuk production:

1. Ubah `JWT_SECRET` ke value yang sangat kuat
2. Ubah database password
3. Set `ENV=production` di .env
4. Gunakan HTTPS
5. Setup firewall dan network security
6. Jangan commit `.env` file ke git

---

## 📞 Support

Jika ada pertanyaan atau error, cek:

1. Dokumentasi service yang digunakan
2. Error message di console
3. Pastikan semua API keys valid

Goodluck! 🚀
