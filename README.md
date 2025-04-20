# goth-todo 📝

A minimal fullstack todo app built using the GOTH stack:  
**Go + HTMX + Templ + PostgreSQL**, deployed with **Fly.io**.

<br/>

## What I Learned

This project helped me dive into fullstack web development with a modern Go workflow. Here’s a breakdown of what I picked up:

---

### Go Backend
- Built a clean HTTP server with the Go standard library
- Handled routing with `http.HandleFunc` and context-aware database queries
- Used the `pgx` driver and `pgxpool` for efficient PostgreSQL connections

---

### Templ (HTML in Go)
- Used [Templ](https://templ.guide) for type-safe HTML templating in Go
- Created reusable UI components (`layout.templ`, `todo.templ`)
- Rendered templates server-side for each request

---

### HTMX Frontend
- Enhanced user interaction with HTMX for inline updates and async behavior
- Enabled no-JS form submission and real-time deletion via `hx-post` and `hx-delete`

---

### PostgreSQL
- Connected to a Neon-hosted Postgres database via `DATABASE_URL`
- Designed a simple `todos` table with an auto-incrementing primary key
- Managed queries for listing, creating, and deleting todos

---

### Fly.io Deployment
- Dockerized the Go app for deployment
- Created and configured a Fly.io app from scratch using their dashboard
- Used custom domain `todo.tonylenguyen.com` with DNS + SSL support
- Managed environment secrets (`DATABASE_URL`) securely

---

## Stack
- **Backend**: Go (`net/http` + pgx)
- **Frontend**: HTMX + Templ
- **Database**: PostgreSQL via Neon
- **Hosting**: Fly.io
- **Domain**: Managed with Fastmail

---

## Screenshots
![image](https://github.com/user-attachments/assets/cdab743b-c550-4cec-8ae2-2d70a831b887)
