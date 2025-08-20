# MemZ

![Build Status](https://github.com/skyrocket-qy/authz/actions/workflows/ci.yml/badge.svg)
![Coverage](https://img.shields.io/badge/coverage-0%25-brightgreen)
![Release](https://img.shields.io/github/v/release/user/repo.svg)
![Dependencies](https://img.shields.io/david/skyrocket-qy/authz.svg)
![Language](https://img.shields.io/github/languages/top/user/repo.svg)


<img src="manifest/icon/f2.png" alt="icon" width="150">

Inspired by [Zanzibar](https://research.google/pubs/zanzibar-googles-consistent-global-authorization-system/), **MemZ** is a distributed, memory-first authorization system optimized for read-heavy workloads.

---

## Goals

- **Read-heavy performance**
- **Distributed with high availability** and eventual consistency
- **High throughput + low latency**
  - 50k+ QPS
  - p95 latency < 10ms
- **In-memory evaluation** for maximum speed
- **Performance-focused** — not all edge cases are supported

---

## Supported Models

- Role-Based Access Control (**RBAC**)
- Hierarchical Relations
- Global admin roles
- Object-less permissions
- Router-style access control
- Filesystem-style permissions
- Static attribute-based access control (**static ABAC**)

---

## Limitations

- Not fully consistent
- No dynamic policy model
- Memory footprint: ~10 GB per 10⁸ tuples
- Single write database — not optimized for write-heavy workloads

---

## Supported Protocols

- gRPC
- ConnectRPC
