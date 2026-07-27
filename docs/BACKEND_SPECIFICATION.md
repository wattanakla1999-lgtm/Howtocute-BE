# Technical Backend Specification (Backend Spec) — Howtocute Studio เชียงใหม่

> **Document Version**: 1.1.0 (Aligned with Go/Gin/GORM Backend Implementation)  
> **Target Application**: Howtocute studio เชียงใหม่ Booking & Salon Management System  
> **Author**: AI Pair Programmer / System Architect  
> **Date**: July 27, 2026  

---

## 1. Overview & Architecture Goals

เอกสารนี้ระบุข้อกำหนดทางเทคนิค (Technical Specification) ของระบบ Backend สำหรับ **Howtocute studio เชียงใหม่** ที่ตรงตามโครงสร้างซอร์สโค้ดปัจจุบันในโปรเจกต์ (Go + Gin + GORM + PostgreSQL/Supabase)

### Key Architecture Features
- **Backend Tech Stack**: Go 1.22+, Gin Framework, GORM ORM, PostgreSQL (Supabase Database & Storage), JWT Auth (Admin & Customer LINE LIFF)
- **Base API Path**: `/api`
- **Queue & Staff Assignment**: ลูกค้าเลือก **บริการ $\rightarrow$ วันที่ $\rightarrow$ เวลา $\rightarrow$ ข้อมูลผู้จอง $\rightarrow$ ชำระมัดจำ (โอนผ่าน PromptPay/แนบสลิป)** โดยร้านค้าจะเป็นผู้กำหนด/มอบหมายช่างประจำเคาน์เตอร์ (`nail_technicians`) ในภายหลังผ่านระบบ Admin
- **Deposit & Slip Verification**: ระบบรองรับการอัปโหลดสลิปมัดจำไปยัง Supabase Storage และให้แอดมินตรวจสอบอนุมัติ (`verified`) หรือปฏิเสธ (`rejected`)
- **Shop Settings Management**: เจ้าของร้านสามารถตั้งค่าสถานะร้าน (เปิด/ปิด), เวลาเปิด-ปิดร้าน, เบอร์โทร, เลข PromptPay และจำนวนเงินมัดจำผ่าน `/api/settings`

---

## 2. Database Specification (GORM Models & PostgreSQL Schema)

### Entity-Relationship Diagram (ERD Overview)

```
[Users / Customers] ───< [Bookings] >─── [Nail Technicians]
                             │
                             ├───> [Service DBs / Services]
                             │
                             └───> [Shop Settings / Deposit Config]

[Admins] (JWT Authentication for Store Management)
```

---

### GORM Schema Definitions & Table Structure

#### 1. `bookings` Table (ตารางการจองคิว)
```sql
CREATE TABLE bookings (
    id BIGSERIAL PRIMARY KEY,
    booking_no VARCHAR(50) NOT NULL UNIQUE,          -- เช่น 'HTC-20260727-001'
    user_id BIGINT NULL REFERENCES users(id) ON DELETE SET NULL,
    service_id BIGINT NOT NULL REFERENCES service_dbs(id) ON DELETE RESTRICT,
    technician_id BIGINT NULL REFERENCES nail_technicians(id) ON DELETE SET NULL,
    start_at TIMESTAMP WITH TIME ZONE NOT NULL,       -- วันเวลาที่เริ่มรับบริการ
    end_at TIMESTAMP WITH TIME ZONE NOT NULL,         -- วันเวลาที่เสร็จบริการ (start_at + duration_minutes)
    customer_name VARCHAR(255) NOT NULL,              -- ชื่อผู้จอง
    customer_phone VARCHAR(50) NOT NULL,              -- เบอร์โทรศัพท์
    service_name VARCHAR(255) NOT NULL,               -- ชื่อบริการ ณ เวลาจอง
    price INT NOT NULL CHECK (price >= 0),            -- ราคาสรุป (บาท)
    duration_minutes INT NOT NULL CHECK (duration_minutes > 0), -- ระยะเวลา (นาที)
    status VARCHAR(20) NOT NULL DEFAULT 'pending',    -- 'pending', 'confirmed', 'in_service', 'completed', 'cancelled', 'no_show'
    payment_method VARCHAR(20) NOT NULL DEFAULT 'cash',-- 'cash', 'transfer', 'card'
    note TEXT NULL,
    cancel_reason TEXT NULL,
    deposit_amount NUMERIC(10,2) NOT NULL DEFAULT 0.00, -- จำนวนเงินมัดจำ
    deposit_status VARCHAR(20) NOT NULL DEFAULT 'none', -- 'none', 'pending', 'verified', 'rejected'
    slip_url TEXT NULL,                               -- URL รูปสลิปบน Supabase Storage
    slip_uploaded_at TIMESTAMP WITH TIME ZONE NULL,
    deposit_reject_reason TEXT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE NULL
);

CREATE INDEX idx_bookings_start_at ON bookings(start_at);
CREATE INDEX idx_bookings_status ON bookings(status);
CREATE INDEX idx_bookings_customer_phone ON bookings(customer_phone);
```

#### 2. `service_dbs` Table (ตารางรายการบริการ)
```sql
CREATE TABLE service_dbs (
    id BIGSERIAL PRIMARY KEY,
    service_id VARCHAR(50) NOT NULL,                -- รหัสบริการอ้างอิง เช่น 'SERV-001'
    service_name VARCHAR(255) NOT NULL,             -- ชื่อบริการ
    service_price INT NOT NULL,                     -- ราคาบริการ
    duration INT NOT NULL,                          -- ระยะเวลาบริการ (นาที)
    image_url TEXT NULL,                            -- รูปภาพบริการ
    img TEXT NULL,
    popular BOOLEAN DEFAULT FALSE,                  -- บริการยอดนิยม
    description TEXT NULL,                          -- รายละเอียดบริการ
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE NULL
);
```

#### 3. `nail_technicians` Table (ตารางช่างประจำร้าน)
```sql
CREATE TABLE nail_technicians (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,                     -- ชื่อช่าง
    role VARCHAR(255) DEFAULT 'ช่างประจำร้าน',      -- ตำแหน่ง / ความถนัด
    avatar_url TEXT NULL,                           -- รูปโปรไฟล์ช่าง (Supabase Storage)
    profile_img TEXT NULL,
    rating NUMERIC(3,2) DEFAULT 5.00,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE NULL
);
```

#### 4. `users` Table (ตารางข้อมูลลูกค้า)
```sql
CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    phone VARCHAR(50) NOT NULL UNIQUE,
    line_user_id VARCHAR(255) NULL UNIQUE,         -- LINE LIFF User ID
    role VARCHAR(20) DEFAULT 'customer',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE NULL
);
```

#### 5. `shop_settings` Table (ตารางตั้งค่าข้อมูลร้านและมัดจำ)
```sql
CREATE TABLE shop_settings (
    id BIGSERIAL PRIMARY KEY,
    shop_status VARCHAR(20) DEFAULT 'open',         -- 'open', 'closed'
    open_time VARCHAR(10) DEFAULT '10:00',
    close_time VARCHAR(10) DEFAULT '20:00',
    shop_phone VARCHAR(50) DEFAULT '0812345678',
    prompt_pay_number VARCHAR(50) NULL,             -- หมายเลข PromptPay สำหรับโอนมัดจำ
    account_name VARCHAR(255) NULL,                 -- ชื่อบัญชีรับโอน
    bank_name VARCHAR(255) NULL,                    -- ชื่อธนาคาร
    deposit_amount INT DEFAULT 100,                 -- จำนวนเงินมัดจำตั้งต้น (บาท)
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
```

#### 6. `admins` Table (ตารางผู้ดูแลระบบ/เจ้าของร้าน)
```sql
CREATE TABLE admins (
    id BIGSERIAL PRIMARY KEY,
    username VARCHAR(100) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    role VARCHAR(20) DEFAULT 'admin',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
```

---

## 3. Booking Status State Machine

```
[ pending ] 
     │  (Customer creates booking & uploads deposit slip)
     ▼
[ deposit_status: pending / status: pending ] 
     │ 
     ├─── (Owner verifies slip: PATCH /api/bookings/:id/verify-slip)
     │       ├── approve ──────► [ deposit_status: verified, status: confirmed ]
     │       └── reject  ──────► [ deposit_status: rejected, status: pending ]
     │
     ├─── (Service Started) ───► [ status: in_service ]
     ├─── (Service Finished) ──► [ status: completed ]
     └─── (Cancelled) ─────────► [ status: cancelled ] / [ status: no_show ]
```

### Status Reference Matrix

| Status Values | Value String | Description |
|---|---|---|
| **Booking Status** | `pending` | สร้างรายการจองแล้ว รอร้านตรวจสอบและยืนยัน |
| | `confirmed` | ร้านตรวจสอบสลิปมัดจำและยืนยันคิวแล้ว |
| | `in_service` | ลูกค้ากำลังรับบริการที่ร้าน |
| | `completed` | ให้บริการเสร็จสิ้น ชำระเงินครบถ้วน |
| | `cancelled` | ยกเลิกคิวจอง (โดยลูกค้าหรือร้านค้า) |
| | `no_show` | ลูกค้าไม่มาตามนัด |
| **Deposit Status** | `none` | ไม่มีการมัดจำ |
| | `pending` | อัปโหลดสลิปแล้ว รอร้านตรวจสอบ |
| | `verified` | ร้านตรวจสอบสลิปอนุมัติแล้ว |
| | `rejected` | สลิปถูกปฏิเสธ (ต้องแนบรูปใหม่หรือติดต่อร้าน) |
| **Payment Method** | `cash` | เงินสด |
| | `transfer` | โอนเงินผ่านธนาคาร / PromptPay |
| | `card` | บัตรเครดิต/เดบิต |

---

## 4. Complete API Endpoint Specification

### 4.1 Authentication APIs

#### 1. Admin Login
- **POST** `/api/auth/login`
- **Request Body**:
  ```json
  {
    "username": "admin",
    "password": "nailly2025"
  }
  ```
- **Response (200 OK)**:
  ```json
  {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "user": {
      "username": "admin",
      "name": "ผู้ดูแลระบบ",
      "role": "admin"
    }
  }
  ```

#### 2. LINE LIFF Customer Login
- **POST** `/api/auth/line`
- **Request Body**:
  ```json
  {
    "access_token": "eyJhbGciOiJSUzI1Ni..."
  }
  ```
- **Response (200 OK)**:
  ```json
  {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "customer": {
      "id": 1,
      "name": "ลูกค้า LINE",
      "phone": "0812345678",
      "lineUserId": "U37c4df1ef2b448acdf28d255a3b3161e"
    }
  }
  ```

#### 3. Check Current Admin Session
- **GET** `/api/auth/me`
- **Headers**: `Authorization: Bearer <AdminToken>`

---

### 4.2 Shop Settings & Public Configuration

#### 4. Get Shop Settings & Deposit Info (Public)
- **GET** `/api/settings`
- **Response (200 OK)**:
  ```json
  {
    "shopStatus": "open",
    "openTime": "10:00",
    "closeTime": "20:00",
    "shopPhone": "0812345678",
    "promptPayNumber": "0812345678",
    "accountName": "ร้าน Howtocute Studio",
    "bankName": "กสิกรไทย",
    "depositAmount": 100
  }
  ```

#### 5. Update Shop Settings (Admin)
- **PUT** `/api/settings`
- **Headers**: `Authorization: Bearer <AdminToken>`
- **Request Body**:
  ```json
  {
    "shopStatus": "open",
    "openTime": "10:00",
    "closeTime": "20:00",
    "shopPhone": "0812345678",
    "promptPayNumber": "0812345678",
    "accountName": "ร้าน Howtocute Studio",
    "bankName": "กสิกรไทย",
    "depositAmount": 100
  }
  ```

---

### 4.3 Services & Technicians Catalogue APIs

#### 6. Get All Services (Public)
- **GET** `/api/services`
- **Response (200 OK)**:
  ```json
  [
    {
      "id": 1,
      "serviceId": "SERV-001",
      "name": "ต่อขนตา ทรงธรรมชาติ",
      "price": 890,
      "duration": 60,
      "imageUrl": "https://<supabase-project>.supabase.co/storage/v1/object/public/profileImg/eyelash-1.jpg",
      "popular": true,
      "description": "ต่อขนตาเส้นต่อเส้น ทรงธรรมชาติ สบายตา"
    }
  ]
  ```

#### 7. Get All Technicians (Public / Admin)
- **GET** `/api/technicians`
- **Response (200 OK)**:
  ```json
  [
    {
      "id": 1,
      "name": "ช่างฝน",
      "role": "ช่างประจำร้าน",
      "avatarUrl": "https://<supabase-project>.supabase.co/storage/v1/object/public/profileImg/tech-1.jpg",
      "rating": 5.0,
      "isActive": true
    }
  ]
  ```

---

### 4.4 Booking & Busy Slot APIs

#### 8. Check Busy / Unavailable Slots for Date
- **GET** `/api/bookings/busy-slots?date=2026-07-28`
- **Response (200 OK)**:
  ```json
  {
    "date": "2026-07-28",
    "busySlots": ["11:00", "14:00"]
  }
  ```

#### 9. Create Booking (Public / Guest Customer)
- **POST** `/api/bookings`
- **Request Body**:
  ```json
  {
    "serviceId": 1,
    "customerName": "วิภาดา สุขใจ",
    "customerPhone": "0812345678",
    "startAt": "2026-07-28T14:00:00+07:00",
    "paymentMethod": "transfer",
    "note": "ต้องการสีธรรมชาติ"
  }
  ```
- **Response (201 Created)**:
  ```json
  {
    "bookingNo": "HTC-20260728-001",
    "id": 108,
    "customerName": "วิภาดา สุขใจ",
    "customerPhone": "0812345678",
    "serviceName": "ต่อขนตา ทรงธรรมชาติ",
    "price": 890,
    "durationMinutes": 60,
    "startAt": "2026-07-28T14:00:00+07:00",
    "endAt": "2026-07-28T15:00:00+07:00",
    "status": "pending",
    "depositAmount": 100,
    "depositStatus": "pending"
  }
  ```

#### 10. Upload Deposit Slip Image
- **POST** `/api/bookings/:id/upload-slip`
- **Request Body**:
  ```json
  {
    "slipUrl": "data:image/jpeg;base64,/9j/4AAQSkZJRg..."
  }
  ```
- **Response (200 OK)**:
  ```json
  {
    "message": "Upload slip successfully",
    "slipUrl": "https://<supabase-project>.supabase.co/storage/v1/object/public/FileUpload/slips/booking_108_1722080000.jpg"
  }
  ```

#### 11. Get Customer's Bookings
- **GET** `/api/bookings/customer?phone=0812345678`
- **GET** `/api/bookings/customer/me` *(Requires Customer JWT Token)*

---

### 4.5 Store Owner / Admin Management APIs

#### 12. List All Bookings (Admin)
- **GET** `/api/bookings`
- **Headers**: `Authorization: Bearer <AdminToken>`
- **Query Params**: `status`, `date`, `search`

#### 13. Verify / Reject Deposit Slip (Admin)
- **PATCH** `/api/bookings/:id/verify-slip`
- **Headers**: `Authorization: Bearer <AdminToken>`
- **Request Body**:
  ```json
  {
    "status": "verified",
    "rejectReason": ""
  }
  ```
  *(เมื่ออนุมัติ `verified` ระบบจะปรับสถานะ booking เป็น `confirmed` โดยอัตโนมัติ)*

#### 14. Assign / Reassign Technician to Booking (Admin)
- **PATCH** `/api/bookings/:id/assign-technician`
- **Headers**: `Authorization: Bearer <AdminToken>`
- **Request Body**:
  ```json
  {
    "technicianId": 1
  }
  ```

#### 15. Update Booking Status (Admin)
- **PATCH** `/api/bookings/:id/status`
- **Headers**: `Authorization: Bearer <AdminToken>`
- **Request Body**:
  ```json
  {
    "status": "in_service"
  }
  ```

#### 16. Dashboard & Reports (Admin)
- **GET** `/api/dashboard/stats`
- **GET** `/api/reports?period=week`
- **GET** `/api/reports?period=month`

---

## 5. System Keep-Alive & Operational Notes

1. **Keep-Alive Endpoint**:
   - `GET /api/keep-alive` หรือ `HEAD /api/keep-alive` ใช้สั่งรัน `SELECT 1` บน PostgreSQL Supabase เพื่อรักษาสถานะ Active ของฐานข้อมูล
2. **Supabase Storage Integration**:
   - Bucket `FileUpload`: สำหรับสลิปโอนเงินมัดจำ
   - Bucket `profileImg`: สำหรับรูปบริการและรูปช่างประจำร้าน
3. **CORS Configuration**:
   - กำหนดผ่าน Environment variable `ALLOW_ORIGIN` ใน `.env`
