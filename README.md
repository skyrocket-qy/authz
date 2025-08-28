# MemZ

<p align="center">
  <img src="manifest/icon/f2.png" alt="icon" width="150">
</p>

<p align="center">
  <strong>A distributed, memory-first authorization system inspired by <a href="https://research.google/pubs/zanzibar-googles-consistent-global-authorization-system/">Zanzibar</a>.</strong>
</p>

<p align="center">
  <a href="https://github.com/skyrocket-qy/authz/actions/workflows/ci.yml"><img src="https://github.com/skyrocket-qy/authz/actions/workflows/ci.yml/badge.svg" alt="Build Status"></a>
  <a href="#"><img src="https://img.shields.io/badge/coverage-0%25-brightgreen" alt="Coverage"></a>
  <a href="https://github.com/skyrocket-qy/authz/releases"><img src="https://img.shields.io/github/v/release/skyrocket-qy/authz.svg" alt="Release"></a>
</p>

MemZ is a high-performance, distributed authorization system designed for read-heavy workloads. It is inspired by Google's Zanzibar, a global authorization system known for its scalability and consistency. MemZ prioritizes in-memory evaluation to achieve low latency and high throughput, making it suitable for applications that require fast and reliable access control checks.

---

## ✨ Features

MemZ is packed with features designed for performance, scalability, and flexibility. Here are some of the key features:

- **High Performance:** Optimized for read-heavy workloads with in-memory evaluation, achieving p95 latency of less than 10ms.
- **High Availability:** Distributed architecture with eventual consistency ensures that the system is always available.
- **High Throughput:** Capable of handling over 50,000 requests per second.
- **Scalable:** Designed to scale horizontally to handle a growing number of requests.
- **Flexible Authorization Models:** Supports a wide range of authorization models, including:
  - **Role-Based Access Control (RBAC):** Control access based on user roles.
  - **Hierarchical Relations:** Define relationships between roles and resources.
  - **Global Admin Roles:** Easily create superuser roles with global permissions.
  - **Object-less Permissions:** Grant permissions that are not tied to a specific object.
  - **Router-style Access Control:** Implement access control similar to how routes are handled in web frameworks.
  - **Filesystem-style Permissions:** Manage permissions in a hierarchical way, similar to a file system.
  - **Static Attribute-Based Access Control (static ABAC):** Control access based on static attributes of users and resources.

## 🏛️ Architecture

MemZ's architecture is designed for high availability and low latency. It consists of a central database (source of truth), a Kafka message queue, and a cluster of authorization replicas.

The update model works as follows:

1.  **Permission Update:** Any change in permissions is written to the central database.
2.  **Event Publication:** An event with a new version number is published to a Kafka topic.
3.  **Replica Update:** Each authorization replica listens to the Kafka topic, consumes the event, and applies the changes to its in-memory data store.
4.  **Resilience:** If a replica fails to apply an update, it will retry. If the retries fail, it will fall back to rebuilding its in-memory store from the central database.
5.  **Distributed Lock:** A distributed lock mechanism is used to ensure that not all replicas are updated at the same time, which prevents downtime.

Here is a simplified diagram of the architecture:

```text
+----------+      +-----------+      +-----------------+
|          |----->|           |----->|                 |
|  Client  |      |   AuthZ   |      |  In-Memory      |
|          |<-----|  Replica  |<-----|  Data Store     |
+----------+      +-----------+      +-----------------+
                      |
                      | Applies Delta
                      |
                  +-------+
                  |       |
                  | Kafka |
                  |       |
                  +-------+
                      ^
                      | Publishes Event
                      |
                +------------+
                |            |
                |  Database  |
                | (Source of |
                |   Truth)   |
                +------------+
```

## 🚀 Performance

The following benchmarks were run on a MacBook Pro with an i7-9750H CPU @ 2.60GHz but limit in docker container with 4 cpus and 8g memory. The tests were run for 10 seconds with a total of 110,000 tuples.

| Virtual Users (VUs) | Requests per Second (RPS) | Average Latency | p95 Latency |
| ------------------- | ------------------------- | --------------- | ----------- |
| 1                   | 1,001                     | 927µs           | 1.48ms      |
| 50                  | 4,100                     | 12.08ms         | 31.29ms     |
| 100                 | 4,198                     | 23.67ms         | 49.2ms      |
| 200                 | 4,086                     | 48.68ms         | 86.54ms     |
| 500                 | 4,116                     | 119.86ms        | 204.34ms    |

## 🏁 Getting Started

To get started with MemZ, you'll need to have the following prerequisites installed:

- **Go:** Version 1.25 or higher.
- **Docker:** To run the required services.

### 1. Clone the Repository

```bash
git clone https://github.com/skyrocket-qy/authz.git
cd authz
```

### 2. Start Required Services

MemZ requires a PostgreSQL database and a Redis instance to be running. You can easily start these services using the provided `Makefile` and Docker.

```bash
make pg
make redis
```

### 3. Run the Application

Once the database and Redis are running, you can start the MemZ server.

```bash
go run .
```

The server will start on port `8080`.

## 💖 Contributing

We welcome contributions from the community! If you'd like to contribute to MemZ, please follow these steps:

1.  **Fork the repository:** Click the "Fork" button at the top right of this page.
2.  **Clone your fork:** `git clone https://github.com/YOUR_USERNAME/authz.git`
3.  **Create a new branch:** `git checkout -b my-new-feature`
4.  **Make your changes:** Add your new feature or fix a bug.
5.  **Commit your changes:** `git commit -am 'Add some feature'`
6.  **Push to the branch:** `git push origin my-new-feature`
7.  **Submit a pull request:** Open a pull request from your fork to the main MemZ repository.

We appreciate your help in making MemZ better!

## 🤝 Supported Protocols
MemZ supports the following protocols for communication:

- **gRPC:** A high-performance, open-source universal RPC framework.
- **ConnectRPC:** A simple, reliable, and interoperable RPC framework for Go.

## ❗ Limitations
MemZ is designed with a focus on performance and read-heavy workloads. As a result, it has the following limitations:

- **Eventual Consistency:** MemZ is not fully consistent and relies on an eventual consistency model.
- **Static Policy Model:** It does not support dynamic policy models.
- **Memory Footprint:** The memory footprint is approximately 10 GB per 100 million tuples.
- **Write Performance:** It uses a single write database and is not optimized for write-heavy workloads.
