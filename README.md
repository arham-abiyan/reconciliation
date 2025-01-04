# Transaction Reconciliation Service

This repository contains a Go-based application for reconciling transactions between internal system records and bank statements.

## Cloning the Repository

To get started, clone this repository and navigate into it:

```bash
git clone https://github.com/arham-abiyan/reconciliation.git
cd reconciliation
```

## Running the Application

### Command-Line Execution

To execute the reconciliation service via the command line, use the following command:

```bash
go run cmd/cmd/main.go -system system-trx.csv -bank bank-a.csv -bank bank-b.csv -start 2024-12-01 -end 2024-12-31
```

**Parameters:**
- `-system`: Path to the system transactions CSV file (e.g., `system-trx.csv`).
- `-bank`: Path to one or more bank statement CSV files (e.g., `bank-a.csv`, `bank-b.csv`).
- `-start`: Start date for the reconciliation timeframe (e.g., `2024-12-01`).
- `-end`: End date for the reconciliation timeframe (e.g., `2024-12-31`).

### Web Server Execution

To execute the reconciliation service as a web server, use the following command:

```bash
go run cmd/server/main.go
```

This will start a web server listening on port `8080`.

#### Making a Request

To perform a reconciliation via the web server, make a `POST` request to the `/api/reconcile` endpoint with the required parameters:

```bash
curl -X POST \
  http://localhost:8080/api/reconcile \
  -H "Content-Type: multipart/form-data" \
  -F "system_file=@system-trx.csv" \
  -F "bank_files=@bank-a.csv" \
  -F "bank_files=@bank-b.csv" \
  -F "start_date=2024-01-01" \
  -F "end_date=2024-12-31"
```

**Form Data Fields:**
- `system_file`: The system transactions CSV file (e.g., `system-trx.csv`).
- `bank_files`: One or more bank statement CSV files (e.g., `bank-a.csv`, `bank-b.csv`).
- `start_date`: Start date for the reconciliation timeframe (e.g., `2024-01-01`).
- `end_date`: End date for the reconciliation timeframe (e.g., `2024-12-31`).

### Notes
- Ensure all required CSV files exist in the appropriate directory.
- Use valid date formats (e.g., `YYYY-MM-DD`) for the `start_date` and `end_date` fields.

