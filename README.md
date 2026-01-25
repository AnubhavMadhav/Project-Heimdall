
# Project Heimdall

> *"Heimdall sees all, hears all, and guards the Bifrost. No enemy shall pass through to Asgard."*

<!-- <img src="assets/heimdall.png" alt="heimdall" width="717"> -->

**Project Heimdall** is a high-security middleware that acts as a **Model Context Protocol (MCP) Gatekeeper** for PostgreSQL databases. It allows Large Language Models (like Claude or Cursor) to query your database to answer questions, while strictly blocking any attempt to modify, delete, or destroy data.

<!-- ![Heimdall Architecture](/heimdall-architecture.png) -->
<img src="heimdall-architecture.png" alt="heimdall-architecture" width="520">

---

## What is Project-Heimdall?

### The Layman's Explanation (The Asgard Analogy)
Imagine your Production Database is **Asgard**—a realm of infinite value that must be protected.
The LLM (Claude/AI) is a traveler trying to cross the **Bifrost** (the bridge) to see what is inside.
**Heimdall** stands at the gate.
* He allows travelers to **look** around ("Show me the users," "List the tables").
* But if a traveler tries to **burn** a village (`DROP TABLE`) or **steal** gold (`DELETE FROM`), Heimdall summons his sword and strictly blocks the path.

### The Technical Explanation
Heimdall is an **MCP Server** written in **Golang**. It sits between the AI client and the PostgreSQL database.
Instead of passing SQL queries directly to the driver, it passes them through a custom **Abstract Syntax Tree (AST) Parser**. This parser deconstructs the SQL command into its grammatical components to ensure it is a read-only `SELECT` statement. If any mutation keywords (`INSERT`, `UPDATE`, `DELETE`, `DROP`, `ALTER`) are detected in the syntax tree, the query is rejected before it ever touches a database connection.

> **LLM (Client)** → **MCP Server** → **AST Security Layer** → **PostgreSQL**


### Key Features
* **AST-Based Security:** Uses `sqlparser` to tokenize and analyze queries, strictly rejecting any non-SELECT statements (superior to regex matching).
* **Model Context Protocol (MCP):** Native integration with the Anthropic ecosystem, exposing `list_tables`, `get_schema`, and `safe_query` as standardized tools.
* **High-Performance Driver:** Built on `pgx/v5` for efficient connection pooling and type handling.

![Heimdall Build & Test](https://github.com/AnubhavMadhav/Project-Heimdall/actions/workflows/go.yml/badge.svg)

---

## Tech Stack & Rationale

We didn't just pick "popular" tools; we picked the *safest* tools for high-concurrency infrastructure.

| Component | Choice | Why we chose it (vs. Alternatives) |
| :--- | :--- | :--- |
| **Language** | **Golang** | **Pros:** Strict typing prevents runtime surprises. Goroutines handle concurrent MCP requests efficiently. <br>**Cons:** More verbose than Python, but safer for infrastructure. |
| **Protocol** | **mcp-go** | **Pros:** Native implementation of the Model Context Protocol. Handles the JSON-RPC handshake so we focus on logic. |
| **Driver** | **pgx/v5** | **Pros:** Uses the PostgreSQL binary protocol (faster than text). Better connection pooling than standard `database/sql`. |
| **Security** | **xwb1989/sqlparser** | **Pros:** Parses SQL into an **Abstract Syntax Tree (AST)**. <br>**Why not Regex?** Regex is easily tricked (e.g., `SeLeCt` vs `SELECT`). An AST understands the *intent* of the code, making it nearly impossible to bypass. |
| **Architecture** | **Hexagonal** | **Pros:** Decouples the "Security Layer" from the "Database Layer." Allows us to swap the database or protocol later without rewriting the core logic. |

---

## How to Setup Locally

Follow these steps to deploy Heimdall on your local machine (Mac, Linux, or Windows WSL).

### Prerequisites
* [Go 1.23+](https://go.dev/dl/)
* [Docker](https://www.docker.com/) (for the local database)
* [Claude Desktop](https://claude.ai/download) (or any MCP-compliant client)

### 1. Clone & Build
```bash
git clone [https://github.com/AnubhavMadhav/Project-Heimdall.git](https://github.com/AnubhavMadhav/Project-Heimdall.git)
cd Project-Heimdall
cd heimdall
go mod tidy
go build -o heimdall cmd/heimdall/main.go
```

Note: This creates a binary executable named heimdall in your folder.


### 2. Start a Local Database (Docker)
If you don't have a database, spin up a safe sandbox:
```
# Run Postgres on port 5432
- docker run --name heimdall-db -e POSTGRES_PASSWORD=password -p 5432:5432 -d postgres

# Create some dummy data to test
- docker exec -it heimdall-db psql -U postgres -c "CREATE TABLE heroes (id SERIAL, name TEXT, role TEXT);"
- docker exec -it heimdall-db psql -U postgres -c "INSERT INTO heroes (name, role) VALUES ('Thor', 'God of Thunder'), ('Loki', 'Trickster');"

# (Crucial) Export the DB URL for manual testing
- export DATABASE_URL="postgres://postgres:password@localhost:5432/postgres?sslmode=disable"
```


### 3. Configure Claude Desktop
Tell Claude where to find your Gatekeeper.

- Mac/Linux: Edit `~/Library/Application Support/Claude/claude_desktop_config.json`
- Windows: Edit `%APPDATA%\Claude\claude_desktop_config.json`

Add this configuration (Update the path to your actual project folder!):
```
{
  "mcpServers": {
    "heimdall": {
      "command": "`/YOUR/FULL/PATH/TO/Project-Heimdall/heimdall/heimdall",
      "args": [],
      "env": {
        "DATABASE_URL": "postgres://postgres:password@localhost:5432/postgres?sslmode=disable"
      }
    }
  }
}
```

### 4. Restart & Chat
Restart Claude Desktop. You should see the 🔌 (plug/connector) connection icon. Ask: "Check the heimdall database. Who are the heroes?"

---

## Demo: Heimdall in Action

See Heimdall protecting the database in real-time interactions with Claude Desktop.

### 1. Discovery & Safe Access
Heimdall allows the LLM to explore the schema and read data safely.

| **Scenario 1: Schema Discovery** | **Scenario 2: Safe Data Retrieval** |
| :---: | :---: |
| ![Listing Tables](assets/demo-discovery.png) | ![Reading Data](assets/demo-select.png) |
| *Claude asks to see available tables.* | *Claude executes a standard `SELECT *` query.* |

### 2. The Gatekeeper (Security Enforcement)
This is the core feature. Notice how Heimdall intercepts the AST before it reaches the database.

| **Scenario 3: Blocking a DELETE** | **Scenario 4: Preventing SQL Injection** |
| :---: | :---: |
| ![Blocked Delete](assets/demo-block-delete.png) | ![Blocked Injection](assets/demo-injection.png) |
| *Claude attempts a `DELETE` command. Heimdall detects the `*sqlparser.Delete` node and rejects it.* | *A stacked query (`SELECT; DROP`) is caught by the parser.* |


---

## Security Deep Dive: AST vs. Regex
Heimdall doesn't use Regular Expressions (Regex) for security because:

The Regex Problem: A naive filter might block strings containing `DELETE`.
- Attack: "Select * from deleted_users" -> Blocked (False Positive)
- Attack: "D E L E T E from users" -> Allowed (False Negative)

The Heimdall (AST) Solution: Heimdall converts the SQL string into a structured tree object. It checks the Node Type.
- Is `stmt` of type `*sqlparser.Select?` Pass.
- Is `stmt` of type `*sqlparser.Delete?` Fail.

This ensures that `SELECT * FROM deleted_users` is allowed (because it's a SELECT node), but `DELETE FROM users` is blocked, no matter how you format the whitespace.

---

## Project Structure
We follow the Standard Go Layout with Hexagonal Architecture:

```
heimdall/
├── cmd/main.go           # Entry point (Wiring & Config)
├── internal/
│   ├── core/             # Pure Business Logic
│   │   ├── ports/        # Interfaces (Gatekeeper)
│   │   └── services/     # Implementation (The Guard)
│   └── adapters/         # Infrastructure
│       ├── mcp/          # The Interface (Claude)
│       ├── postgres/     # The Storage (Asgard)
│       └── security/     # The Firewall (AST Parser)
└── pkg/logger/           # Structured Logging
```

*Built with ❤️ and Golang by Anubhav Madhav.*