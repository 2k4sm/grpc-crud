
# GRPC-CRUD

A simple user management service built using Golang, ScyllaDB, gRPC and served through HTTP using gRPC-gateway which provides a RESTful API interface that proxies HTTP requests to compatible gRPC actions.


## Tech Stack

- **Go**: Primary programming language
- **gRPC**: High-performance RPC framework
- **gRPC-Gateway**: Proxies HTTP APIs to gRPC
- **Protocol Buffers**: Interface definition language
- **ScyllaDB**: Performant NoSQL database
- **docker**/**docker-compose**: Containerization and environment setup
- **GitHub Actions**: CI/CD pipeline for auto deployment
- **Buf**: Protocol buffer compiler and dependency management


## Project Structure

```
.
├── buf.gen.yaml
├── buf.lock
├── buf.yaml
├── compose.yaml
├── Dockerfile
├── go.mod
├── go.sum
├── main.go
├── proto
│   └── users
│       ├── users_grpc.pb.go
│       ├── users.pb.go
│       ├── users.pb.gw.go
│       └── users.proto
├── README.md
└── src
    ├── db
    │   └── db.go
    ├── models
    │   └── User.go
    └── services
       └── userService.go
```


## Prerequisites

- Go 1.23.0 or later
- Docker and Docker Compose
- Buf CLI
- Git


## Setup and Installation

### 1. Clone the Repository

```bash
git clone https://github.com/2k4sm/grpc-crud.git
cd grpc-crud
```


### 2. Start ScyllaDB with Docker Compose

#### Set env variables into `.env`.
```bash
touch .env
echo "SDB_URI=localhost" >> .env
```
Then run
```bash
docker-compose up -d
```


This will start a ScyllaDB instance as defined in the `compose.yaml` file.

### 3. Install Dependencies

```bash
go mod download
```


### 4. Install Protobuf deps and Generate Protocol Buffer Code using buf

```bash
buf dep update
```
and
```bash
buf generate
```

### 5. Build and Run the Application

```bash
go build -o grpc-crud .
./grpc-crud
```

Alternatively, you can use:

```bash
go run main.go
```

### Local Ports
- grpc-gateway(Http) -> 6969
- grpc(tcp) -> 8080

### Deployed to EC2

This is configured for automatic deployment to EC2 using GitHub Actions. When changes are pushed to the main branch, the CI/CD pipeline will:

1. Build the application
2. Deploy to the EC2 instance

### Deployed URL: https://grpc-crud.2k4sm.tech

### Postman Collection
[grpc-crud.postman_collection.json](https://github.com/2k4sm/grpc-crud/blob/main/grpc-crud.postman_collection.json)

## HTTP Endpoints
- GET /users?email={email}&ph_number={ph_number}: Get a user by email or phone number

  ```bash
  curl -X GET "http://localhost:6969/users?email=john.doe@example.com"
  ```

  or

  ```bash
  curl -X GET "http://localhost:6969/users?ph_number=1234567890"
  ```
- POST /users: Create a new user

    ```bash
    curl -X POST http://localhost:6969/users \
      -H "Content-Type: application/json" \
      -d '{
            "first_name": "John",
            "last_name": "Doe",
            "gender": "MALE",
            "dob": "1990-01-01",
            "ph_number": "1234567890",
            "email": "john.doe@example.com",
            "access": "UNBLOCKED"
          }'
    ```
- PUT /users/{email}: Update a user by email

    ```bash
    curl -X PUT http://localhost:6969/users/john.doe@example.com \
      -H "Content-Type: application/json" \
      -d '{
            "first_name": "John",
            "last_name": "Doe",
            "gender": "MALE",
            "dob": "1990-01-01",
            "ph_number": "0987654321",
          }'
    ```
- PATCH /users/{curr_email}: Update a user's email or phone number

  ```bash
  curl -X PATCH http://localhost:6969/users/john.doe@example.com \
    -H "Content-Type: application/json" \
    -d '{
          "new_email": "john.new@example.com",
          "new_ph_number": "1122334455"
        }'
  ```
- POST /users/{email}/block: Block a user

  ```bash
  curl -X POST http://localhost:6969/users/jane.doe@example.com/block
  ```
- POST /users/{email}/unblock: Unblock a user

  ```bash
  curl -X POST http://localhost:6969/users/jane.doe@example.com/unblock
  ```

# Thank You for trying out grpc-crud.
