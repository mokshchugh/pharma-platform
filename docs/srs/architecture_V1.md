# Software Requirements Specification (SRS)

## Pharmaceutical Industrial Data Acquisition & Analytics Platform

**Version:** 0.1 (Architecture Draft)

**Status:** Design Phase

---

# 1. Introduction

## 1.1 Purpose

This document defines the software architecture, requirements, and design decisions for an Industrial Data Acquisition Platform intended for deployment in pharmaceutical manufacturing facilities.

The system aims to provide a secure, scalable, and production-grade solution for collecting telemetry from heterogeneous PLCs, storing high-frequency time-series data, aggregating production metrics, and exposing analytics through a unified API.

Unlike traditional SCADA systems, this platform is intended to serve as an independent analytics layer without modifying PLC logic or interfering with existing manufacturing systems.

---

# 2. Project Objectives

The platform shall:

* Acquire production data from multiple PLC vendors.
* Store raw telemetry efficiently.
* Generate production KPIs.
* Provide real-time dashboards.
* Support historical reporting.
* Remain vendor independent.
* Be deployable using Docker.
* Operate continuously (24×7).
* Support future AI/ML modules.
* Maintain production-grade security.

---

# 3. Technology Stack

## Backend

* Go

## Databases

* QuestDB
* PostgreSQL

## Frontend

* React (planned)

## Infrastructure

* Docker
* Docker Compose

## Operating System

* Arch Linux (development)
* Linux Server (deployment)

---

# 4. High-Level Architecture

```
                  Users
                     │
              HTTPS / TLS
                     │
             Reverse Proxy
                     │
               Go API Server
          ┌──────────┴──────────┐
          │                     │
          ▼                     ▼
      QuestDB             PostgreSQL
          ▲                     ▲
          │                     │
   Aggregation Service──────────┘
          ▲
          │
    PLC Collector
          ▲
          │
        PLC Network
```

---

# 5. Core Components

## PLC Collector

Responsibilities:

* Connect to PLCs.
* Read telemetry.
* Normalize data.
* Validate values.
* Write raw telemetry to QuestDB.

The collector SHALL NOT communicate directly with PostgreSQL.

---

## QuestDB

Purpose:

Storage of raw, high-frequency time-series telemetry.

Examples:

* Cycle count
* Machine state
* Temperature
* Pressure
* Production counters
* Alarm events

QuestDB is the authoritative source of all raw industrial data.

---

## Aggregation Service

Responsibilities:

* Read historical data from QuestDB.
* Calculate production metrics.
* Generate summaries.
* Upsert aggregated records into PostgreSQL.

Examples:

* Hourly OEE
* Daily production
* Shift reports
* MTBF
* MTTR
* Availability
* Performance
* Quality

---

## PostgreSQL

Purpose:

Persistent storage of business and configuration data.

Includes:

* Users
* Roles
* Machine registry
* PLC configuration
* Tag definitions
* Alarm definitions
* Batch information
* Product metadata
* Shift schedules
* Aggregated KPIs
* Audit logs

---

## API

The API acts as the single access point to the platform.

Responsibilities:

* Authentication
* Authorization
* Data aggregation
* Response composition
* Business logic

The frontend SHALL NEVER communicate directly with any database.

---

## Frontend

Responsibilities:

* Dashboard visualization
* Historical trends
* Reports
* Alarm monitoring
* Configuration interface

---

# 6. Data Flow

Raw Data Flow

```
PLC
    ↓
Collector
    ↓
QuestDB
```

Aggregated Data Flow

```
QuestDB
      ↓
Aggregation Service
      ↓
PostgreSQL
```

Query Flow

```
Frontend
      ↓
API
      ↓
QuestDB

or

Frontend
      ↓
API
      ↓
PostgreSQL

or

Frontend
      ↓
API
      ↓
QuestDB + PostgreSQL
```

---

# 7. Database Responsibilities

## QuestDB

Stores:

* Raw telemetry
* Time-series events
* Sensor values
* Machine status
* Production counters

Characteristics:

* High write throughput
* Time-series optimized
* Immutable historical data

---

## PostgreSQL

Stores:

Configuration

* Machines
* PLCs
* Tags

Business

* Products
* Recipes
* Batches

Analytics

* OEE
* Downtime
* Reports

Security

* Users
* Roles
* Permissions

Audit

* Configuration changes
* User actions

---

# 8. Security Architecture

## Principle

Every component shall only access resources required for its operation.

---

Collector

Permissions

* Write QuestDB only

No PostgreSQL access.

---

Aggregation Service

Permissions

* Read QuestDB
* Write PostgreSQL

---

API

Permissions

* Read QuestDB
* Read PostgreSQL

No direct writes to production telemetry.

---

Frontend

Permissions

* HTTPS access to API only

No database access.

---

# 9. Trust Boundaries

```
Internet
      │
Reverse Proxy
      │
API Network
      │
Backend Network
      │
Database Network
```

Database containers SHALL NOT be exposed publicly.

---

# 10. Repository Pattern

The backend shall implement:

```
HTTP Handler
      ↓
Service Layer
      ↓
Repository Layer
      ↓
Database
```

Repositories shall isolate SQL from business logic.

---

# 11. Deployment Model

Containers

* API
* Collector
* Aggregation Service
* QuestDB
* PostgreSQL
* Grafana

Persistent storage SHALL exist outside the Git repository.

Example

```
~/Projects/pharma-platform

~/Projects/pharma-platform-data
```

---

# 12. Non-Functional Requirements

Availability

* 24×7 operation

Performance

* Continuous telemetry ingestion

Reliability

* Automatic restart
* Health monitoring

Scalability

* Multiple PLCs
* Multiple production lines

Maintainability

* Modular services
* Docker deployment
* Version-controlled configuration

Security

* Least privilege
* Role-based access
* HTTPS
* Audit logging

---

# 13. Future Enhancements

* OPC UA integration
* Native PLC drivers
* MQTT support
* AI anomaly detection
* Predictive maintenance
* Machine learning analytics
* Multi-site deployment
* High availability
* Kubernetes deployment
* Enterprise authentication (LDAP/OIDC)

---

# 14. Current Design Decisions (ADR Summary)

1. Go selected as the primary implementation language.
2. Docker Compose selected for deployment.
3. QuestDB stores raw telemetry.
4. PostgreSQL stores configuration and aggregated business data.
5. API is the only public backend interface.
6. Collector writes exclusively to QuestDB.
7. Aggregation Service transfers summarized data from QuestDB to PostgreSQL.
8. Persistent storage resides outside the Git repository.
9. Security follows the Principle of Least Privilege.
10. Modular architecture preferred over a monolithic implementation.

---

# 15. Project Status

Current Phase:

Architecture Design

Completed

* Technology stack selection
* Development environment planning
* Docker strategy
* Database architecture
* Security model
* Component responsibilities
* High-level system architecture

Next Milestone

Infrastructure Deployment

* Docker Compose
* QuestDB
* PostgreSQL
* Grafana
* Internal networking
* Persistent volumes
* Health checks
