# Avalanche Consensus Service

This project implements the Avalanche consensus protocol using dynamic graph algorithms for directed acyclic graphs (DAGs). The implementation follows the Model-View-Controller (MVC) architecture pattern for better organization and maintainability.

## Architecture

The project is structured using the MVC architecture pattern:

### Model Layer
- **`models/dag`**: Core data structures for directed acyclic graphs
- **`models/consensus`**: Implementation of the Avalanche consensus algorithm
- **`models/vertex`**: Business logic for vertex operations

### View Layer
- **`views`**: Response formatters and templates for API responses

### Controller Layer
- **`controllers`**: Request handlers that coordinate between models and views

### Service Layer
- **`services`**: Business logic that integrates multiple models and provides services to controllers

### Additional Components
- **`routes`**: Routing definitions that map URLs to controller methods
- **`middleware`**: HTTP middleware for cross-cutting concerns like logging
- **`config`**: Configuration management
- **`cmd`**: Application entry points

## API Endpoints

The service exposes the following RESTful API endpoints:

### Vertex Operations
- `POST /api/v1/vertex` - Submit a new vertex to the network
- `GET /api/v1/vertex/{id}` - Get details about a specific vertex
- `GET /api/v1/vertices` - List all vertices
- `GET /api/v1/vertices/finalized` - List all finalized vertices

### Peer Operations
- `GET /api/v1/connect?nodeID={id}` - Connect to this node
- `GET /api/v1/peers` - List all connected peers
- `POST /api/v1/peers/connect` - Connect to a list of peers

### Consensus Operations
- `POST /api/v1/consensus/start` - Start the consensus algorithm
- `POST /api/v1/consensus/stop` - Stop the consensus algorithm
- `GET /api/v1/consensus/status` - Get consensus status

### Health Check
- `GET /health` - Check if the service is running

## Running the Service

### Configuration

Create a configuration file (`config.json`):

```json
{
  "server_port": 8080,
  "node_id": "node-1",
  "peer_addresses": [
    "http://peer1:8080",
    "http://peer2:8080"
  ],
  "consensus_params": {
    "k": 10,
    "alpha": 8,
    "beta_virtuous": 20,
    "beta_rogue": 30
  }
}
```

### Starting the Service

```bash
# Run with a custom config file
go run src/cmd/main.go --config=my-config.json

# Run in simulation mode
go run src/cmd/main.go --simulation
```

## Development

### Project Structure

```
src/
├── models/          # Data structures and business logic
│   ├── dag/         # DAG implementation
│   ├── consensus/   # Consensus algorithms
│   └── vertex/      # Vertex models and validation
├── views/           # Response formatters
├── controllers/     # Request handlers
├── services/        # Business logic layer
├── routes/          # Route definitions
├── middleware/      # HTTP middleware
├── config/          # Configuration management
└── cmd/             # Application entry points
    └── main.go      # Main application
```

### Extending the Project

To add new features:

1. Add new models to the `models/` directory
2. Implement business logic in the `services/` directory
3. Create controllers in the `controllers/` directory
4. Update routes in the `routes/` directory
5. Use middleware in the `middleware/` directory for cross-cutting concerns

## How Avalanche Consensus Works

The Avalanche consensus protocol works by repeatedly sampling the network to determine which transactions (vertices in the DAG) should be accepted. The protocol has the following key parameters:

- `K`: The sample size (number of validators to query)
- `Alpha`: The threshold for decision making (number of positive responses needed)
- `BetaVirtuous`: The confidence threshold for virtuous vertices
- `BetaRogue`: The confidence threshold for conflicting vertices

The protocol operates as follows:

1. A node proposes a new vertex (transaction) to the network.
2. The vertex is added to the DAG and linked to its parent vertices.
3. The consensus algorithm repeatedly queries a random subset of the network to determine the preference for each vertex.
4. When a vertex receives enough consecutive positive responses, it is finalized.

## Future Improvements

- Add authentication and authorization for API endpoints
- Implement persistent storage for the DAG
- Add metrics and monitoring support
- Implement a client library for easier integration
- Add Swagger/OpenAPI documentation for the API
