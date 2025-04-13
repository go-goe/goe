```
go get -u -t
```

```
docker compose up -d
```

```
GOE_DRIVER=SQLite go test . -v -race -count=1
```

```
docker compose down
```

```
GOE_DRIVER=SQLite go test -bench Select -benchmem -run ^$

GOE_DRIVER=SQLite go test -bench Join -benchmem -run ^$
```