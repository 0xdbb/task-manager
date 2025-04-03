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
## Installation

### Prerequisites

Ensure you have the following installed:

- [Go](https://go.dev/dl/)
- [Docker](https://www.docker.com/)
- [Docker Compose](https://docs.docker.com/compose/)
- [kubectl](https://kubernetes.io/docs/tasks/tools/) - Optional
- [Minikube](https://minikube.sigs.k8s.io/docs/) - Optional

### Setup

1. Clone the repository:

   ```sh
   git clone git@github.com:0xdbb/task-manager.git
   cd task-manager



1. Start the application.
   ```sh
    make up

	
2. Run database migrations

   ```sh
    make goose-up

3. afsdfasd
   ```sh
    make up


4. adfasdf
   ```sh
    make up


5. asdfasd
   ```sh
    make up

6. adfsdsd


   ```sh
    make up
