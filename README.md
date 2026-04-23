# Bank Service (NPCI - Bank Adapter)

## 📌 Overview

The **Bank Service** acts as a **middleware adapter between the NPCI switch and a specific bank's internal ledger**. 

It is responsible for:
* Translating asynchronous Kafka instructions into synchronous HTTP calls to the bank's core banking API.
* Maintaining strict idempotency to ensure no operation is executed twice.
* Handling fallback status inquiries from the switch.
* Providing proxy endpoints for direct account resolution.

This service models the **gateway infrastructure hosted by individual banks** to connect to the unified NPCI network.

##  🔑  Responsibilities

### 1. Protocol Translation
* Consumes Kafka messages (Debit, Credit, Refund instructions).
* Fires synchronous HTTP requests to the `bank-management` API.
* Produces the HTTP response outcome back to Kafka.

### 2. Idempotency Management
* Prevents duplicate deductions or credits if the core switch retries an instruction.
* Caches the outcome of bank operations locally.

### 3. Status Enquiry Handling
* Consumes inquiry events from the core switch.
* Checks local cache or queries the core banking system to return the status of an ambiguous transaction.

### 4. Integration with High-Speed Ledgers
* Interfaces with a core banking ledger designed for high-throughput. It aligns with the upstream ledger's optimized reconciliation approach, performing reconciliation right before get-balance checks rather than performing costly per-transaction validations.

---

## 📨 Kafka Topics

| Topic Name            | Purpose                                     |
| --------------------- | ------------------------------------------- |
| `bank.instruction.v1` | Incoming commands from the core switch      |
| `bank.enquiry.v1`     | Incoming status check requests              |
| `bank.response.v1`    | Outgoing operation results to the switch    |
| `bank.response.failed`| Dead Letter Queue routing for poison pills  |

The Bank Service:
* **consumes** → `bank.instruction.v1`, `bank.enquiry.v1`
* **produces** → `bank.response.v1`, `bank.response.failed`

---

## 🧩 Design Decisions

### Edge Idempotency
* Idempotency is enforced *before* the request hits the internal bank management system. This shields the core banking database from redundant load.

### Graceful Degradation
* Errors returned by the HTTP client are analyzed. Transient errors (like `503 Service Unavailable` or timeouts) are rejected without committing the Kafka offset, allowing natural retry backoff.

### Shared Group IDs
* Kafka consumer groups (`bank-grp`) allow multiple instances of the Bank Service to run in parallel, automatically load-balancing instructions across partitions.

---

## 🧠 Redis Usage

Redis is used as the **Idempotency Store**.

### Purpose:
* Cache the success/failure outcome of processed requests.
* TTL is set to 24 hours to handle delayed network retries.

### Example:
```text
Key: idempotency:<txn_id><operation_type>
Value: {
  "status": "SUCCESS",
  "bank_ref": "REF98765"
}
```

---

## Tech Stack

* **Language:** Go
* **Framework:** Gin (for supplementary proxy routes)
* **Messaging:** Kafka
* **Cache:** Redis
* **Serialization:** Protocol Buffers

---

## Project Structure

```text
bank-service/
├── cmd/
├── internals/
│   ├── handlers/       # REST endpoints for account verification
│   ├── http_client/    # client for core banking system communication
│   ├── kafka/
│   ├── middlewares/
│   ├── pb/
│   ├── repository/     # redis store operations
│   ├── routes/
│   ├── services/       # adapter logic & idempotency checks
│   └── workers/        # instruction & status consumer loops
└── go.mod
```