# DB to Excel exporter

- First you need to install `task` tool [taskfile.dev](https://taskfile.dev/)

- Run `task` command to list all available tasks

## Usage

### Clone the repository

```bash
git clone https://github.com/savage-demon/db-excel-exporter
cd db-excel-exporter

```

### Build the binary

```bash
task build
```

### Usage example

```bash
./exporter --query="select * from products" \
  --db-url="user=postgres dbname=postgres sslmode=disable password=postgres_password port=5412 host=localhost" \
  --page-size=100000 \
  --output=products.xlsx
```
