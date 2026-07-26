# Nail Booking API

Go REST API learning project using Gin, GORM, and PostgreSQL.

## Run Locally

```bash
go run ./cmd/api
```

## Admin Authentication

Set `JWT_SECRET`, `ADMIN_USERNAME`, and `ADMIN_PASSWORD` in `.env` before deployment. Local development defaults to `admin` / `nailly2025`.

```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"nailly2025"}'
```

Protected endpoints require `Authorization: Bearer <token>`. `GET /api/auth/me` can be used to validate the current session.

## Shop Settings

`GET /api/settings` is public so the customer booking flow can read PromptPay/deposit details.
`PUT /api/settings` requires an admin token.

```text
GET /api/settings
PUT /api/settings
```

The update body contains `shopStatus` (`open` or `closed`), `openTime`, `closeTime`, `shopPhone`, `promptPayNumber`, `accountName`, `bankName`, and `depositAmount`. Times use the `HH:MM` 24-hour format.

## Slip Uploads

Set these environment variables to upload base64 slip images to Supabase Storage:

```bash
SUPABASE_URL=https://your-project-ref.supabase.co
SUPABASE_SERVICE_ROLE_KEY=your-service-role-key
SUPABASE_STORAGE_BUCKET=FileUpload
```

`POST /api/bookings/:id/upload-slip` accepts `{ "slipUrl": "data:image/png;base64,..." }`, uploads the image to the configured bucket, and stores the resulting public URL on the booking.
The `FileUpload` bucket should be public, or configured with a read policy that lets the admin frontend display stored slip images.

## Dashboard

`GET /api/dashboard/stats` requires an admin token and returns today's appointment/revenue totals, customer and service counts, today's appointment list, and popular services. Popular-service `rate` is `0` until review data is added.

## CORS

Set `ALLOW_ORIGIN` to the frontend origin that can call this API. Use `*` for local development, or a real frontend URL in production.

## Keep Supabase Active

Use this endpoint from a cron job or uptime monitor to create a lightweight database activity:

```bash
curl https://your-api-domain.com/api/keep-alive
```

The endpoint runs `SELECT 1` against PostgreSQL and returns `200 OK` when the database is reachable.

## Test

```bash
go test ./...
```

## Reports

```bash
curl "http://localhost:8080/api/reports?period=week"
curl "http://localhost:8080/api/reports?period=month"
```

Report rules:

- Revenue and payment breakdown include only completed bookings.
- Appointment totals exclude cancelled and no-show bookings.
- The daily revenue target is currently `5000` baht.
- Try completing `TestReportServiceMonthPractice` in `internal/service/report_service_test.go` as a Go exercise.

## Docker

```bash
docker compose up --build
```

## Structure

```txt
cmd/api        application entrypoint
internal       app-specific packages
pkg            reusable helper packages
docs           project docs
```
