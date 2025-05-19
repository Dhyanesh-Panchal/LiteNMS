# LiteNMS Backend

This is the backend service for LiteNMS (Lite Network Monitoring System). It provides RESTful APIs for managing network devices, credential profiles, discovery profiles, and monitoring data.

## API Endpoints

### 1. Device Management

#### Get Provisioned Devices
- **GET** `/api/devices`
- **Description**: Returns all provisioned devices
- **Response**:
  ```json
  {
    "devices": [
      {
        "ip": 2130706433,  // IPv4 as uint32 (e.g., 127.0.0.1)
        "credential_id": 1,
        "is_provisioned": true
      }
    ]
  }
  ```

### 2. Credential Profile Management

#### Get Credential Profiles
- **GET** `/api/credential-profiles`
- **Description**: Returns all credential profiles
- **Response**:
  ```json
  {
    "profiles": [
      {
        "id": 1,
        "hostname": "example.com",
        "password": "secret",
        "port": 22
      }
    ]
  }
  ```

#### Create Credential Profile
- **POST** `/api/credential-profiles`
- **Description**: Creates a new credential profile
- **Request Body**:
  ```json
  {
    "hostname": "example.com",
    "password": "secret",
    "port": 22
  }
  ```
- **Response**:
  ```json
  {
    "message": "Credential profile created successfully",
    "id": 1
  }
  ```

#### Update Credential Profile
- **PUT** `/api/credential-profiles/:id`
- **Description**: Updates an existing credential profile
- **URL Parameters**:
  - `id`: The ID of the credential profile to update
- **Request Body**:
  ```json
  {
    "hostname": "new.example.com",
    "password": "newsecret",
    "port": 23
  }
  ```
- **Response**:
  ```json
  {
    "message": "Credential profile updated successfully"
  }
  ```

### 3. Discovery Profile Management

#### Get Discovery Profiles
- **GET** `/api/discovery-profiles`
- **Description**: Returns all discovery profiles
- **Response**:
  ```json
  {
    "profiles": [
      {
        "id": 1,
        "device_ips": [2130706433, 3232235777],  // IPv4s as uint32
        "credential_profile_ids": [1, 2]
      }
    ]
  }
  ```

#### Create Discovery Profile
- **POST** `/api/discovery-profiles`
- **Description**: Creates a new discovery profile
- **Request Body**:
  ```json
  {
    "device_ips": [2130706433, 3232235777],  // IPv4s as uint32
    "credential_profile_ids": [1, 2]
  }
  ```
- **Response**:
  ```json
  {
    "message": "Discovery profile created successfully",
    "id": 1
  }
  ```

#### Update Discovery Profile
- **PUT** `/api/discovery-profiles/:id`
- **Description**: Updates an existing discovery profile
- **URL Parameters**:
  - `id`: The ID of the discovery profile to update
- **Request Body**:
  ```json
  {
    "device_ips": [2130706433, 3232235777],  // IPv4s as uint32
    "credential_profile_ids": [1, 2]
  }
  ```
- **Response**:
  ```json
  {
    "message": "Discovery profile updated successfully"
  }
  ```

### 4. Monitoring Data

#### Get Histogram Data
- **POST** `/api/histogram`
- **Description**: Returns histogram data for specified devices and counter
- **Request Body**:
  ```json
  {
    "from": 1744610677,
    "to": 1744620677,
    "counterID": 1,
    "objectIDs": [169093227, 2130706433]
  }
  ```
- **Response**: Returns histogram data points for each object ID

#### Get Histogram Data (Deprecated)
- **GET** `/api/histogram`
- **Description**: Deprecated endpoint that returns an example request
- **Response**: Example request body for the POST endpoint

## Error Responses

All endpoints may return the following error responses:

- **400 Bad Request**: Invalid request body or parameters
- **404 Not Found**: Requested resource not found
- **500 Internal Server Error**: Server-side error

Error responses follow this format:
```json
{
  "error": "Error message describing what went wrong"
}
```

## Notes

1. IPv4 addresses are stored as uint32 integers. For example:
   - 127.0.0.1 = 2130706433
   - 192.168.1.1 = 3232235777

2. The deprecated GET `/api/histogram` endpoint should not be used in new code. Use the POST endpoint instead.

3. All timestamps are Unix timestamps (seconds since epoch).

## Development

### Prerequisites
- Go 1.24+
- ZeroMQ libraries installed on your system

### Installation
1. Clone the repository
2. Run `go mod download` to download dependencies
3. Start the server with `go run main.go`

### Architecture
The backend uses:
- Gin for HTTP routing
- ZeroMQ for communication with the reporting database
- Clean architecture with controllers, models, and interfaces

### Components
- `reportdb`: Client for communicating with the reporting database over ZeroMQ
- `db`: Database interfaces and adapters
- `controllers`: HTTP request handlers
- `models`: Data structures used throughout the application
- `routes`: API route definitions 