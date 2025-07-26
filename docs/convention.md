# 🧭 Convention Guide

## 🔤 Abbreviations
>
> Any abbreviation must be explained as globally and early as possible in documentation or comments.

| Full Term   | Abbreviation |
|-------------|--------------|
| controller  | `ctrl`       |
| usecase     | `ucase`      |
| service     | `svc`        |
| client      | `cli`        |

---

## 📁 Folder Naming Convention

### 1. Use Plural for "Collections" of Things

Use plural when the folder contains **multiple implementations** of a type.

**Examples:**

- `/controllers` → `user_controller.go`, `auth_controller.go`
- `/services` → `user_svc.go`, `payment_svc.go`
- `/repositories` → `user_repo.go`, `order_repo.go`
- `/models` → `user.go`, `product.go`

**Why?**  
Plural naming indicates the folder is a **collection** of similar entities.

---

### 2. Use Singular for Domain-Centric or Utility Folders

Use singular when the folder represents a **single concept** or a **single-purpose utility**.

**Examples:**

- `/domain/user.go`
- `/pkg/validator`

**Why?**  
Singular emphasizes atomicity — each folder or file focuses on **one concept**.

---

### 3. Common Exceptions

- `/cmd` – Always singular (Go convention for executables)
- `/internal` – Always singular (Go standard)

---

## 📦 Package Naming Convention

- Always use **singular** for package names.
- Use lowercase with no underscores.

---

## 🧠 Function Naming Convention

- Use `verb + noun` as much as possible.
- Use function **comments to explain "why" the function exists**.
- Avoid overly generic verbs like `handle`, `process`, or `do`.

**Bad:**

```go
func HandleUser()
```

**Good:**

```go
func CreateUser()
```