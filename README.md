# Train Ticket Service

## Overview

This project is an implementation of a train ticket service using gRPC in Go, where users can purchase tickets, view receipts, and perform administrative actions. Here's an overview of the key components and functionalities:


## Getting Started

Make sure you have the following installed on your system:

* Go: [Installation Guide](https://golang.org/doc/install)
* Protocol Buffers and gRPC: [Installation Guide](https://grpc.io/docs/protoc-installation/)
* MakeFile: [Installation Guide](https://gnuwin32.sourceforge.net/packages/make.htm)

## How to use

Clone the repository:
```shell
git clone https://github.com/playbody/train-ticket-service.git
cd train-ticket-service
```

Build the project:

```shell
make build
```

Unit test the project:

```shell
make test
```

## Requirement

**Staff Software Engineer – Team Lead**

* Code must be published in GitHub with a link we can access (use public repo).
* Code must compile with some effort on unit tests, does not have to be 100%, but it shouldn’t be 0%.
* Please code this with Golang and gRPC.
* No persistence layer is required, just store the data in the current session/in memory.
* The results can be in the console output from your grpc-server and grpc-client.
* Depending on the level of authentication, take different actions.

Background: 
 > All API referenced are gRPC APIs, not REST ones.\
 > I want to board a train from London to France. The train ticket will cost $20, regardless of section or seat. 
* Authenticated APIs should be able to parse a JWT, formatted as if from an OAuth2 server, from the metadata to authenticate a request. No signature validation is required.
* Create API where you can submit a purchase for a ticket. 
  * (From, To, User, Price Paid) User should include first and last name, email address. 
  * The user is allocated a seat in the train as a result of the purchase. Assume the train has only 2 sections, section A and section B and each section has 10 seats.
* Auth API that shows the details of the receipt for the user
* Auth API that lets an admin view the users and seat they are allocated by the requested section
* Auth API to allow an admin or the user remove a user from the train
* Auth API to allow an admin or the user to modify a user’s seat

## Solution

### Configuration

The configuration is loaded from a YAML file. Modify the `config.yaml` file to customize the train sections, seat count, and routes.

```yaml
train:
 seat_count: 100
 sections:
  - A
  - B
 routes:
  - from: London
    to: France
    price: 20
  - from: Paris
    to: Berlin
    price: 30
  - from: New York
    to: Los Angeles
    price: 40
auth:
 secret_key: VhkgDGkS-k0J9A2KTZJm31kZnvQon7viSD2OtkB4V_c=
 expire: 3600
roles:
 - email: admin@test.com
   caps: admin, read, write
 - email: read@test.com
   caps: read
 - email: write@test.com
   caps: write
```

### Project Structure
1. server/server.go:

Defines the `TrainServer` struct, which implements the gRPC service.
Implements methods like `PurchaseTicket`, `GetReceipt`, `GetUsersBySection`, `RemoveUser`, `ModifySeat`, etc.
Handles the logic for ticket purchase, seat allocation, receipt retrieval, user removal, and seat modification.

2. server/config.go:

Defines the configuration structures (`Config`, `AuthConfig`, `RoleUser`, `TrainConfig`) used for server configuration and initialization.
Provides a method (`InitConfig`) to read configuration from a YAML file.

3. server/auth.go:

Implements JWT authentication-related functionalities.
Defines a custom `JwtClaims` struct representing JWT claims.
Provides functions for JWT generation (`GenerateJWT`), authentication check (`AuthCheck`), and JWT parsing middleware (`ParseJWTMiddleware`).

4. server/util.go:

Contains utility functions such as `isValidUser` for validating user data.

5. main.go:

The entry point of the server application.
Initializes the gRPC server, loads configuration, and registers the `TrainServer`.
Uses JWT parsing middleware to handle authentication for incoming requests.

### API Functionality
User or Admin can auth using this API
```go
func (s *TrainServer) AuthUser(_ context.Context, req *proto.AuthRequest) (*proto.AuthResponse, error)
```

Create API where you can submit a purchase for a ticket (Public API)
```go
func (s *TrainServer) PurchaseTicket(_ context.Context, req *proto.PurchaseRequest) (*proto.PurchaseResponse, error) 
```

An API that shows the details of the receipt for the user (Authenticated API)\
auth check logic: user or (admin | read) capability 
```go
func (s *TrainServer) GetReceipt(_ context.Context, req *proto.ReceiptRequest) (*proto.ReceiptResponse, error)
```

An API that lets you view the users and seat they are allocated by the requested section (Authenticated API)\
auth check logic: (admin | read) capability
```go
func (s *TrainServer) GetUsersBySection(_ context.Context, req *proto.SectionRequest) (*proto.SectionResponse, error)
```
An API to remove a user from the train (Authenticated API)\
auth check logic: user or (admin | write) capability
```go
func (s *TrainServer) RemoveUser(_ context.Context, req *proto.RemoveUserRequest) (*proto.RemoveUserResponse, error)
```
An API to modify a user’s seat (Authenticated API)\
auth check logic: user or (admin | write) capability
```go
func (s *TrainServer) ModifySeat(_ context.Context, req *proto.ModifySeatRequest) (*proto.ModifySeatResponse, error)
```
