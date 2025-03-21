# URL Shortener - Docker Setup & API Documentation

## Running the Docker Container

To start the URL shortener service, run the following command:

```sh
docker run -d -p 80:8080 slincnik/urlshortener
```

This will launch the container in detached mode and map port `8080` inside the container to port `80` on the host machine.

## API Endpoints

### 1. Create a Shortened URL
- **Endpoint:** `/shorten`
- **Method:** `POST`
- **Request Body:**
  ```json
  {
   "url": "https://example.com" 
  }
  ```
- **Response:**
  ```json
  {
   "short_url": "abc123" 
  }
  ```
- **Description:**
  Sends a URL in the request body and receives a shortened code.

### 2. Redirect to Original URL
- **Endpoint:** `/`
- **Method:** `GET`
- **Example:**
  ```sh
  curl -X GET "http://localhost/abc123"
  ```
- **Description:**
  Takes a short URL code as a parameter and redirects to the original URL.

## Example Workflow

1. Start the container using the command:
   ```sh
   docker run -d -p 80:8080 slincnik/urlshortener
   ```
2. Create a shortened URL:
   ```sh
   curl -X POST -H "Content-Type: application/json" -d '{ "url": "https://example.com" }' http://localhost/shorten
   ```
3. Use the returned `short_url` to access the original link:
   ```sh
   curl -X GET "http://localhost/abc123"
   ```