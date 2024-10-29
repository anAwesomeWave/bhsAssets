
# bhsAssets

## 📖 About
This repository contains the backend code for the BHS marketplace, written in Go. It provides a web application to manage and trade assets on the marketplace.

**!Note - jwt secret is public, should be changed in production!**
## ⚙️ Prerequisites
Ensure the following prerequisites are installed on your system:
1. **Docker**  
   Verify the installation:
   ```bash
   docker -v
   ```

2. **GNU Make**  
   Verify the installation:
   ```bash
   make -v
   ```

## 📦 Installation
1. **Clone the Repository**  
   Clone this repository from GitHub:
   ```bash
   git clone git@github.com:anAwesomeWave/bhsAssets.git
   ```

2. **Navigate to the Project Directory**
   ```bash
   cd bhsAssets
   ```

## 💡 Quickstart
You have several options to start the application (you may need to run some commands with `sudo`):

- **Launch the Entire Infrastructure**  
  This will start the database, apply migrations, build a project image, and run the binary file:
  ```bash
  make up
  ```

- **Launch Database and Apply Migrations**  
  Note: Use external ports to connect to the database on your host system.
  - Use port `5432` for the PostgreSQL database running in Docker.
  - Use port `54321` for a database running on the host system.
  ```bash
  make dev
  go build ./cmd/app
  ./app -v
  ```

- **Test the System**
  ```bash
  make test
  ```

- **Stop the System**
  ```bash
  make down
  ```
  *(Optional) Clear the Docker build cache (see Issues & Peculiarities).*

> **Tip:** Check other options in the Makefile.

## 🚧 Issues & Peculiarities
1. **Docker Build Cache Issues**  
   The app may  work poorly with the Docker build cache. You may want to check the cache size:
   ```bash
   docker system df
   ```
   And clean the cache if necessary:
   ```bash
   docker buildx prune -f
   ```

2. **Cross-Platform Compatibility**  
   While efforts have been made to maintain cross-platform support, the application works best on Linux.

## 🌐 Routes and Endpoints

### 🖥️ API Routes
The API is accessible via different base URLs depending on the environment:
- **Docker socket:** `http://localhost:8080`
- **Host socket:** `http://localhost:8082`

#### Asset Management
- `GET /api/assets?name={name}&min_price={price}&max_price={price}`  
  Retrieves a list of all assets that match query parameters.


- `POST /api/assets/{id}`  
  Purchases an asset by ID.
  **Response Status Codes:**
    - `201 Crreated` - Transaction successful.
    - `403 Forbidden` - You cannot buy this asset. You are the author.
    - `500 Internal Server Error` - Server error.


- `POST /api/assets/create`  
  Creates a new asset.

**Request Body:**
```json
{
    "name": "string (required)",
    "description": "string (optional)",
     "balance": "number (optional)"
}
```

- `GET /api/assets/{id}`  
  Retrieves details of a specific asset.


#### Authentication
- `POST /api/auth/login`  
  Logs in a user.  
  **Request Body:**  
  ```json
  {
    "login": "string (required)",
    "password": "string (required)"
  }
  ```
  **Response Status Codes:**  
  - `200 OK` - Successful login.
  - `400 Bad Request` - Invalid input.
  - `404 Not Found` - User not found.
  - `500 Internal Server Error` - Server error.

  Set and return new jwt token. It automatically will be placed to cookies 'jwt' (useful for postman and browsers).
  You can provide it with Header: "Authorization: BEARER {token}"


- `POST /api/auth/register`  
  Registers a new user.  
  **Request Body:**  
  ```json
  {
    "login": "string (required)",
    "password": "string (required)"
  }
  ```
  **Response Status Codes:**  
  - `201 Created` - Registration successful.
  - `400 Bad Request` - Invalid input.
  - `500 Internal Server Error` - Server error.

#### User Management
- `PATCH /api/users/balance/`  
  Set new current user's balance.
  **Request Body:**
  ```json
  {
    "balance": "number (required)"
  }
  ```
- `GET /api/users/me`  
  Retrieves information about the current user.
  **Response Status Codes:**
    - `200 OK` - returns json data.
    - `401 unauthorized` - jwt token was not found.
    - `500 Internal Server Error` - Server error.
**Response Body:**
  ```json
  {
    "Id": "number",
    "login": "string",
    "balance": "number"
  }
  ```
### 🌍 Website Routes
- `/`  
  Redirect to /assets

- `/assets`  
  Displays all available assets.

- `/assets/{id}`  
  Buy a specific asset.

- `/assets/create`  
  Ccreate a new asset.

- `/assets/{id}`  
  Details page for a specific asset.

- `/auth/login`  
  Login page.

- `/auth/logout`  
  Logout action.

- `/auth/register`  
  Registration page.

- `/static/`  
  Serves static files.

- `/users/balance`  
  Displays the user's balance.

- `/users/me`  
  Displays the user's profile information.

## 🛣️ Roadmap
1. **API Support Improvements** ✔️  
   Improve support for the API and ensure it functions correctly.

2. **Add user's page**  
   Add a user page with information about the user and all the resources created by him.

3. **Jwt token expiration**  
   Add user's token expiration after generating new one for him.

4. **Testing**  
   Cover the code with unit and functional tests.

5. **Documentation**  
   Generate Swagger/OpenAPI documentation. Currently, the API documentation is difficult to maintain as it is comment-based instead of code-based.

## 🤝 Contributing

Contributions are welcome! Please open an issue or submit a pull request.

