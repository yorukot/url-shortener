# URL Shortener with Expiration - README
This is a simple URL shortener service written in Go, with support for optional expiration times (in minutes). It uses MongoDB for storage and Gin as the HTTP framework.

## Features
- Create short URLs for any long URL.
- Optional expiration time (in minutes) for URLs.
- Redirect using the short URL, with expiration checks.
- Basic token-based authentication.

## Prerequisites
- Go 1.19+ installed on your machine.
- MongoDB instance running and accessible.
- API_TOKEN and MONGODB_URI set in environment variables.

## Setup
Clone the repository:

```bash
git clone https://github.com/your-repo/url-shortener.git
cd url-shortener
```

Install dependencies:
```bash
go mod tidy
```
Set up environment variables:
Create a .env file in the root directory:
```bash
echo "API_TOKEN=your_token_here" >> .env
echo "MONGODB_URI=mongodb://username:password@localhost:27017/urlshortener" >> .env
echo "GIN_MODE=release" >> .env
```

Run the application:
```bash
go run main.go
```

The server will start on port 8080.

## API Usage
Authentication
All requests must include an Authorization header with the value of the API_TOKEN set in the environment.

Endpoints
### Create/Shorten a URL
Request
`POST /shorten`

Headers:

```json
{
  "Authorization": "Bearer your_api_token"
}
```
Body:

```json
{
  "long_url": "https://example.com",
  "short_url": "customAlias",   // Optional
  "exp": 10                    // Optional, expiration in minutes
}
```
- long_url: The original URL (required).
- short_url: A custom alias for the short URL (optional; auto-generated if not provided).
- exp: Expiration time in minutes (optional; URL never expires if not provided).

Response:
```json
{
  "long_url": "https://example.com",
  "short_url": "abc123",
  "exp": 1709999999  // UNIX timestamp of expiration (if provided)
}
```
### Redirect to Long URL
Request
`GET /:shortURL`

#### Example:

```bash
curl -X GET http://localhost:8080/abc123
```

- Behavior:
If the URL is valid and not expired, it redirects to the original long_url.

If the URL has expired, responds with:
```json
{ "error": "URL has expired" }
```
If the URL does not exist, responds with:
```json
{ "error": "URL not found" }
```

### Use curl create a Short URL exmaple 
```bash
curl -X POST http://localhost:8080/shorten \
-H "Authorization: your_api_token" \
-H "Content-Type: application/json" \
-d '{
  "long_url": "https://example.com",
  "exp": 10
}'
```
#### Response:

```json
{
  "long_url": "https://example.com",
  "short_url": "abc123",
  "exp": 1709999999
}
```
#### Redirect Using Short URL
```bash
curl -X GET http://localhost:8080/abc123
```
- Redirects to https://example.com if not expired.
- Returns an error if expired or not found.

## Notes:

- Ensure MongoDB is running and accessible at the URI specified in MONGODB_URI.
- Expired URLs are not deleted automatically; you can create a background job for cleanup if needed.

Feel free to modify and extend this project as required! 🎉