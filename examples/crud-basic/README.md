## Setup
```bash
PORT=8888 go run main.go
```

## Usage

### Get List
```bash
curl -H "Accept: application/json" "http://localhost:8888/persons?page=1&size=5"
```

### Get One
```bash
curl -H "Accept: application/json" "http://localhost:8888/persons/1"
```

### Post

```bash
curl -X POST \
-H "Content-Type: application/json" \
-H "Accept: application/json" \
-d '{"name":"John Constantine", "email":"devilmaycry@yooou.com"}' "http://localhost:8888/persons"
```

### Put

```bash
curl -X PUT \
-H "Content-Type: application/json" \
-H "Accept: application/json" \
-d '{"name":"John Constantine Update", "email":"devilmaycry@yooou.com"}' "http://localhost:8888/persons/1"
```

### Delete

```bash
curl -X DELETE \
-H "Accept: application/json" "http://localhost:8888/persons/1"
```