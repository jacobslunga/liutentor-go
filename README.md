# LiUTentor Go API

Simple Go API for fetching exam data.

## Tech

- Go
- Echo v5
- Supabase

## Requirements

- Go (1.26+)
- Supabase project and service key

## Environment variables

Create a `.env` file for local development (or set env vars directly):

```env
APP_ENV=development
PORT=1323
SUPABASE_URL=https://your-project.supabase.co
SUPABASE_SERVICE_KEY=your-service-role-key
```

## Run locally

```bash
go run ./cmd/api
```

Server starts on `http://localhost:1323` by default.

## Run locally with Air (hot reload)

This project already includes an Air config in `.air.toml`.

Install Air (pick one):

```bash
go install github.com/air-verse/air@latest
```

or on macOS with Homebrew:

```bash
brew install air
```

Run with Air:

```bash
air
```

Air builds and runs `./cmd/api` using the config in `.air.toml` and restarts automatically on file changes.

## API routes

Base path: `/v1`

- `GET /v1/exams/:university/:courseCode`
- `GET /v1/exams/:examId`

### `GET /v1/exams/:university/:courseCode`

Example request:

```http
GET /v1/exams/LIU/TDDD27
```

Example success response (`200`):

```json
{
  "data": {
    "courseCode": "TDDD27",
    "courseName": "Programmering och problemlosning",
    "exams": [
      {
        "id": 101,
        "course_code": "TDDD27",
        "exam_date": "2024-08-22",
        "pdf_url": "https://example.com/exams/tddd27-2024-08-22.pdf",
        "exam_name": "Ordinarie tentamen",
        "has_solution": true,
        "statistics": {
          "registered": 120,
          "passed": 83
        },
        "pass_rate": 69.2
      }
    ]
  },
  "message": "Exams fetched successfully"
}
```

Example error responses:

```json
{
  "error": "Invalid university"
}
```

```json
{
  "error": "No exam documents found for this course"
}
```

### `GET /v1/exams/:examId`

Example request:

```http
GET /v1/exams/101
```

Example success response (`200`):

```json
{
  "data": {
    "exam": {
      "id": 101,
      "course_code": "TDDD27",
      "exam_date": "2024-08-22",
      "pdf_url": "https://example.com/exams/tddd27-2024-08-22.pdf"
    },
    "solution": {
      "id": 55,
      "exam_id": 101
    }
  },
  "message": "Exam fetched successfully"
}
```

Example error responses:

```json
{
  "error": "examId must be a positive integer"
}
```

```json
{
  "error": "Exam not found"
}
```

## Deploy (Railway)

Set these Railway variables:

- `APP_ENV=production`
- `SUPABASE_URL`
- `SUPABASE_SERVICE_KEY`

`PORT` is provided by Railway.
