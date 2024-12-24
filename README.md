# DB to Excel exporter

- First you need to install `task` tool [taskfile.dev](https://taskfile.dev/)

- Run `task` command to list all available tasks

## Usage

### Build the binary

```bash
task build
```

### Example

```bash
./exporter --query="select * from products" \
  --db-url="user=postgres dbname=postgres sslmode=disable password=postgres_password port=5412 host=localhost" \
  --page-size=100000 \
  --output=products.xlsx
```
