# Cinemesh-Core

**Enterprise-Grade REST API Gateway & Metadata Hub for Distributed Movie Platform**

Cinemesh-Core serves as the authoritative Identity Provider and Movie Bank, managing user authentication, movie metadata, and orchestrating requests across distributed sub-systems.

---

## Architecture Overview

```
┌─────────────────────────────────────────────────────────┐
│                    Cinemesh-Core API                     │
│              (Backend-for-Frontend Gateway)              │
├─────────────────────────────────────────────────────────┤
│  • JWT Authentication & User Management                  │
│  • Movie Metadata Hub (Cast, Crew, Synopsis)            │
│  • Search & Filtering Engine                             │
│  • Sub-System Orchestration Layer                        │
└─────────────────────────────────────────────────────────┘
                           ▲
                           │
                  ┌────────┴────────┐
                  │   Frontend UI    │
                  └─────────────────┘
```

---

## 🚀 Quick Start

### Prerequisites
- Go 1.21+
- Docker & Docker Compose
- Make (optional)

### 1 Clone & Setup
```bash
# Clone the repository
git clone https://github.com/Ponloe/cinemesh-core
cd cinemesh-core

# Install dependencies
go mod download
```

### 2 Configure Environment
```bash
# Copy environment template
cp .env.example .env

# Edit .env and set your JWT_SECRET
nano .env
```
### 3 Run the server
```bash

go run ./cmd/server

```