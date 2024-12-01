docker run -d \
  --name blogdb \
  -p 5433:5432 \
  -e POSTGRES_USER=bloguser \
  -e POSTGRES_DB=blogdb \
  -e POSTGRES_PASSWORD=yourpassword \
  postgres:16