### SimpleBank: A Comprehensive Banking Application

Welcome to **SimpleBank**, a robust banking application developed using Go. This application is designed to demonstrate various modern technologies and practices in software development, including HTTP and gRPC services, gRPC Gateway, Docker containerization, and token-based authentication using JWT and PASETO.

#### Key Features

1. **HTTP and gRPC Services**:
   - **HTTP API**: Provides RESTful endpoints for various banking operations, making it easy to interact with the application using standard HTTP requests.
   - **gRPC API**: Offers a high-performance, language-agnostic RPC protocol for more efficient communication between services. gRPC is ideal for scenarios where low latency and high throughput are critical.

2. **gRPC Gateway**:
   - Acts as a bridge between HTTP RESTful services and gRPC services. The gRPC Gateway allows clients to use standard HTTP methods to interact with gRPC services, simplifying the integration process with existing systems and front-end applications.

3. **Docker Containerization**:
   - The application is containerized using Docker, ensuring consistent environments across development, testing, and production. Docker simplifies deployment and scaling by packaging the application along with its dependencies into a portable container.

4. **Token-Based Authentication**:
   - **JWT (JSON Web Token)**: Provides a standardized method for securely transmitting information between parties as a JSON object. JWTs are used for stateless authentication, where tokens are passed between client and server for user verification.
   - **PASETO (Platform-Agnostic Security Tokens)**: An alternative to JWT, PASETO is designed to address some of the security shortcomings of JWT. PASETO tokens are simpler and more secure, offering a robust solution for authentication and data integrity.
