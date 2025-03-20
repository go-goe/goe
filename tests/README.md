```
go get -u -t
```

```
docker compose up -d
```

```
go test . -race -count=1
```

```
docker compose down
```

```
go test -bench Select -benchmem -run ^$

go test -bench Join -benchmem -run ^$
```