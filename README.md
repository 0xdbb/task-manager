# Task Management System

## Overview

The **Task Management System** is a backend service designed to handle task processing efficiently, supporting user authentication and authorization, distributed task execution, real-time updates, and high availability. 

Built with:
-
- Golang
- PostgreSQL
- RabbitMQ
- Docker
- Kubernetes


The core functionality includes:
- JWT-based authentication with role-based access control (RBAC) for admins, standard users, and task workers.
- Task creation, claiming, and processing with a message queue and task locking.
- Fault-tolerant task processing with automatic retries for failed tasks.
- A scalable architecture deployable with Docker.


## Architecture

### Components
1. **API Service** (Golang):
   - Built with the `gin` framework for lightweight, fast routing.
   - Endpoints for authentication (`/auth/*`), user management (`/user/`) and task management (`/task/*`).
   -Publishes task creation events with priority to RabbitMQ. 

2. **Database** (PostgreSQL):
   - Stores users (with roles) and tasks. See database docs: [Database Docs](https://dbdocs.io/dennisboachie9/task-management-system)
   - Indexes set on frequently accessed fields for efficient retrieval.

3. **Message Queue** (RabbitMQ):
   - `task-queue`: Distributes tasks to workers in a round-robin fashion

4. **Worker Service** (Golang):
   - Consumes tasks from RabbitMQ, parses payload, fetches weather data , and updates status.
   - Implements round-robin task claiming among multiple workers.

5. **Deployment** (Docker/Kubernetes):
   - Containerized services: `api`, `worker`, `postgres-db`, `rabbitmq`.
   - Kubernetes manifests provided for production-ready scaling.

### Diagram
```mermaid
graph TD
    A[User] -->|HTTP Requests| B[API:8000]
    B -->|Publishes Tasks| C[RabbitMQ:task-queue]
    C -->|Consumes| D[Worker]
    C -->|Retries Failed Tasks| E[RabbitMQ:task-retry]
    D -->|Updates Status| F[PostgreSQL]
    E -->|Requeues| D
    B -->|Reads/Writes| F
    A -->|Swagger UI| G[API Docs]
    F -->|Polls Status| B
    B -->|Returns Updates| A

    subgraph "Real-Time (Polling)"
        F --> A
    end
``````

### Prerequisites

Ensure you have the following installed:

- [Go](https://go.dev/dl/)
- [make](https://www.gnu.org/s/make/manual/make.html)
- [Docker](https://www.docker.com/)
- [Docker Compose](https://docs.docker.com/compose/)
- [kubectl](https://kubernetes.io/docs/tasks/tools/) - Optional
- [Minikube](https://minikube.sigs.k8s.io/docs/) - Optional

### Setup

```text
task-manager/
├── cmd/                    # Entry points for the application
│   ├── api/                # Main API service executable (Golang)
│   └── worker/             # Worker service executable (Golang)
├── docs/                   # Documentation resources
│   ├── db/                 # Database schema and migration files
│   └── swagger/            # Swagger API specification files
├── internal/               # Private application code (Golang internal packages)
│   ├── config/             # Configuration loading and environment variable handling
│   ├── database/           # Database connection and query logic (PostgreSQL)
│   ├── queue/              # Message queue integration (RabbitMQ)
│   ├── server/             # HTTP server setup and routing (gin framework)
│   ├── token/              # JWT authentication and token management
│   ├── weather/            # Weather-related logic (placeholder or custom task processing)
│   └── worker/             # Worker-specific business logic for task processing
├── k8s/                    # Kubernetes manifests for deployment
│   ├── api-deployment.yaml # API service deployment configuration
│   ├── config-secrets.yaml # Secrets for environment variables (e.g., JWT_SECRET)
│   ├── hpa.yaml            # Horizontal Pod Autoscaler for dynamic scaling
│   ├── postgres.yaml       # PostgreSQL deployment and service config
│   ├── rabbitmq.yaml       # RabbitMQ deployment and service config
│   └── worker-deployment.yaml # Worker service deployment configuration
└── util/                   # Shared utility functions and data structures

```


1. Clone the repository:

   ```sh
   git clone git@github.com:0xdbb/task-manager.git
   cd task-manager
   ````

1. Start the application. This should build and/or pull required images.
   ```sh
    make up
    ````
    Check out `Makefile` for more commands

    - Here 4 containers will be started
        - worker
        - api
        - postgres-db
        - rabbitmq

3. visit [localhost](http://localhost:8000/api/v1) in your browser to view docs

4.  You can visit the rabbitmq management UI at [ui](http://localhost:15672/) in your browser to view docs

### System test flow
<!--
 !   ╭───────────────────────────────────────────────────────╮
 !   │ To test api with swagger docs:                        │
 !   ╰───────────────────────────────────────────────────────╯
-->

1. **Register**:
    - register user as ADMIN first using example data  at `/auth/register` .

2. **Login**:
 - login at `/auth/login` and copy the `access_token`.


3. **Set Authorization Header**:
    - In Swagger, click "Authorize" and enter Bearer `<access_token>`.  

4. **Create Task**:
    - Let's skip over to `tasks` section. Create task at POST `/task`. 
Copy the ID of the logged-in admin from the `/auth/login` response and use it in the example request body.

5. **Poll Task Status**:
    - Once task is created, you can head over to `/task/{id}/status` to poll database for change in task status.
Workers will process the task (status transitions: `pending` → `in-progress` → `completed`/`failed`).

6. **Role-Based Behavior**:
   - Admin: Full access (manage users/tasks).
    - Standard: Create tasks, update/get own tasks.
    - Worker: Claim and process tasks (internal role, not user-facing).

## Deliverables
- API Docs: Swagger UI at [localhost](http://localhost:8000/api/v1).
- Tests: Unit tests (`make test`) cover authentication and task logic; integration tests in tests/.
- Deployment: Dockerized services; live demo at [api](https://task-manager-6x3jxg.fly.dev/api/v1)
