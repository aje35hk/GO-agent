# Agent-Controller System

This project implements a remote execution system where a central **Controller** manages and sends instructions to a connected **Agent**.

## Architecture

The system consists of two main components running as Docker containers:

1.  **Controller**:
    *   Acts as the central server.
    *   Exposes a WebSocket endpoint (`/ws`) for the Agent to connect.
    *   Exposes an HTTP API (`/execute`) and Web UI to receive user commands.
    *   Bridges HTTP requests to the active WebSocket connection.

2.  **Agent**:
    *   Connects to the Controller via WebSocket.
    *   Listens for instructions.
    *   Executes Bash commands locally.
    *   Returns the output to the Controller.

## Communication Protocol

The communication happens over a persistent WebSocket connection.

### 1. Instruction (Controller -> Agent)

When you send a command via the UI, the Controller sends a JSON message to the Agent:

```json
{
  "id": "uuid-string",
  "type": "bash",
  "payload": "echo hello world"
}
```

*   `id`: Unique identifier for request-response correlation.
*   `type`: Type of instruction (currently supports `bash`).
*   `payload`: The actual command to execute.

### 2. Response (Agent -> Controller)

After execution, the Agent sends back a JSON response:

```json
{
  "id": "uuid-string",
  "status": "success",
  "message": "Command executed",
  "output": "hello world\n"
}
```

*   `id`: Matches the instruction ID.
*   `status`: `success` or `error`.
*   `output`: Standard output and standard error combined.

## Usage

### Prerequisites
*   Docker and Docker Compose

### Running the System

1.  Start the containers:
    ```bash
    docker-compose up --build
    ```

2.  Access the Web UI:
    *   Open [http://localhost:8080](http://localhost:8080) in your browser.

3.  Execute Commands:
    *   Enter a bash command (e.g., `ls -la`, `uname -a`) in the text area.
    *   Click **Run Command**.
    *   View the output from the Agent.

### API Usage

You can also execute commands via `curl`:

```bash
curl -X POST http://localhost:8080/execute \
     -H "Content-Type: application/json" \
     -d '{"type": "bash", "payload": "echo hello via curl"}'
```

## Directory Structure

*   `agent/`: Go source code for the agent.
*   `controller/`: Go source code for the controller and web UI.
*   `docker-compose.yml`: Orchestration for running both services.
