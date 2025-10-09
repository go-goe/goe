# GOE
GO Entity or just "GOE" is a type safe ORM for Go

[![test status](https://github.com/go-goe/goe/actions/workflows/tests.yml/badge.svg "test status")](https://github.com/go-goe/goe/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/go-goe/goe)](https://goreportcard.com/report/github.com/go-goe/goe)
[![Go.Dev reference](https://img.shields.io/badge/go.dev-reference-blue?logo=go&logoColor=white)](https://pkg.go.dev/github.com/go-goe/goe)
[![MIT license](https://img.shields.io/badge/license-MIT-brightgreen.svg)](https://opensource.org/licenses/MIT)

![](goe.png)
*GOE logo by [Luanexs](https://www.instagram.com/luanexs/)*

## Requirements
- go v1.24 or above;

## Features
Check out the [Benchmarks](#benchmarks) section for a overview on GOE performance compared to other packages like, ent, GORM, sqlc, and others.

- 🚫 Non-String Usage;
	- write queries with only Go code
- 🔖 Type Safety;
	- get errors on compile time
- 📦 Auto Migrations;
	- automatic generate tables from your structs
- 📄 SQL Like queries; 
	- use Go code to write queries with well known functions
- 🗂️ Iterator
	- range over function to iterate over the rows
- 📚 Pagination
	- paginate your large selects with just a function call
- ♻️ Wrappers
	- wrappers for simple queries and builders for complex ones


## Content
- [Install](#install)
- [Available Drivers](#available-drivers)
    - [PostgreSQL](#postgresql)
	- [SQLite](#sqlite)
- [Quick Start](#quick-start)
- [Database](#database)
	- [Supported Types](#supported-types)
	- [Struct Mapping](#struct-mapping)
	- [Setting primary key](#setting-primary-key)
	- [Setting type](#setting-type)
	- [Setting null](#setting-null)
	- [Setting default](#setting-default)
	- [Relationship](#relationship)
		- [One to One](#one-to-one)
		- [Many to One](#many-to-one)
		- [Many to Many](#many-to-many)
		- [Self Referential](#self-referential)
	- [Index](#index)
		- [Create Index](#create-index)
		- [Unique Index](#unique-index)
		- [Two Columns Index](#two-columns-index)
	- [Schemas](#schemas)
	- [Logging](#logging)
	- [Open](#open)
	- [Migrate](#migrate)
		- [Auto Migrate](#auto-migrate)
		- [Drop and Rename](#drop-and-rename)
		- [Migrate to a SQL file](#migrate-to-a-sql-file) 
- [Select](#select)
	- [Find](#find)
	- [List](#list)
	- [Select From](#select-from)
	- [Select Specific Fields](#select-specific-fields)
	- [Where](#where)
	- [Join](#join)
	- [OrderBy](#orderby)
	- [Pagination](#pagination)
	- [Aggregates](#aggregates)
	- [Functions](#functions)
- [Insert](#insert)
	- [Insert One](#insert-one)
	- [Insert Batch](#insert-batch)
- [Update](#update)
	- [Save](#save)
	- [Update Set](#update-set)
- [Delete](#delete)
	- [Remove](#remove)
	- [Delete Batch](#delete-batch)
- [Transaction](#transaction)
	- [Begin Transaction](#begin-transaction)
	- [Commit and Rollback](#commit-and-rollback)
	- [Isolation](#isolation)
- [Benchmarks](#benchmarks)

## Install
```
go get github.com/go-goe/goe
```
As any database/sql support in go, you have to get a specific driver for your database, check [Available Drivers](#available-drivers)

## Available Drivers
### PostgreSQL
```
go get github.com/go-goe/postgres
```
#### Usage
```go
type Animal struct {
	// animal fields
}

type Database struct {
	Animal         *Animal
	*goe.DB
}

dns := "user=postgres password=postgres host=localhost port=5432 database=postgres"
db, err := goe.Open[Database](postgres.Open(dns, postgres.Config{}))
```

### SQLite
```
go get github.com/go-goe/sqlite
```

#### Usage
```go
type Animal struct {
	// animal fields
}

type Database struct {
	Animal         *Animal
	*goe.DB
}

db, err := goe.Open[Database](sqlite.Open("goe.db", sqlite.Config{}))
```
## Quick Start
```go
package main

import (
	"fmt"

	"github.com/go-goe/goe"
	"github.com/go-goe/sqlite"
)

type Animal struct {
	ID    int
	Name  string
	Emoji string
}

type Database struct {
	Animal *Animal
	*goe.DB
}

func main() {
	db, err := goe.Open[Database](sqlite.Open("goe.db", sqlite.Config{}))
	if err != nil {
		panic(err)
	}
	defer goe.Close(db)

	err = goe.Migrate(db).AutoMigrate()
	if err != nil {
		panic(err)
	}

	err = goe.Delete(db.Animal).All()
	if err != nil {
		panic(err)
	}

	animals := []Animal{
		{Name: "Cat", Emoji: "🐈"},
		{Name: "Dog", Emoji: "🐕"},
		{Name: "Rat", Emoji: "🐀"},
		{Name: "Pig", Emoji: "🐖"},
		{Name: "Whale", Emoji: "🐋"},
		{Name: "Fish", Emoji: "🐟"},
		{Name: "Bird", Emoji: "🐦"},
	}

	err = goe.Insert(db.Animal).All(animals)
	if err != nil {
		panic(err)
	}

	animals, err = goe.List(db.Animal).AsSlice()
	if err != nil {
		panic(err)
	}
	fmt.Println(animals)
}
```
## Database
```go
type Database struct {
	User    	*User
	Role    	*Role
	UserLog 	*UserLog
	*goe.DB
}
```
In goe, it's necessary to define a Database struct,
this struct implements *goe.DB and a pointer to all
the structs that's it's to be mappend.

It's through the Database struct that you will
interact with your database.

### Supported Types

GOE supports any type that implements the [Scanner Interface](https://pkg.go.dev/database/sql#Scanner). Most common are sql.Null types from database/sql package.

```go
type Table struct {
	Price      decimal.Decimal     `goe:"type:decimal(10,4)"`
	NullId     sql.Null[uuid.UUID] `goe:"type:uuid"`
	NullString sql.NullString      `goe:"type:varchar(100)"`
}
```

[Back to Contents](#content)
### Struct mapping
```go
type User struct {
	Id        	uint //this is primary key
	Login     	string
	Password  	string
}
```
> [!NOTE] 
> By default the field "Id" is primary key and all ids of integers are auto increment.

[Back to Contents](#content)
### Setting primary key
```go
type User struct {
	Identifier	uint `goe:"pk"`
	Login     	string
	Password	string
}
```
In case you want to specify 
a primary key use the tag value "pk".

[Back to Contents](#content)
### Setting type
```go
type User struct {
	Id       	string `goe:"pk;type:uuid"`
	Login    	string `goe:"type:varchar(10)"`
	Name     	string `goe:"type:varchar(150)"`
	Password 	string `goe:"type:varchar(60)"`
}
```
You can specify a type using the tag value "type"

[Back to Contents](#content)

### Setting null
```go
type User struct {
	Id        int
	Name      string
	Email     *string // this will be a null column
	Phone     sql.NullString `goe:"type:varchar(20)"` // also null
}
```

> [!IMPORTANT] 
> A pointer is considered a null column in Database.

[Back to Contents](#content)

### Setting default

```go
type User struct {
	Id        int
	Name      string
	Email     *string
	CreatedAt  time.Time `goe:"default:current_timestamp"`
}
```

To ensure that a default value will be used, call the `IgnoreFields` on the `Insert` function, otherwise GOE will try to insert the value stored on the field.

```go
err = goe.Insert(db.User).IgnoreFields(&db.User.CreatedAt).One(&u)
```

[Back to Contents](#content)

### Relationship
In goe relational fields are created using the pattern TargetTable+TargetTableId, so if you want to have a foreign key to User, you will have to write a field like "UserId" or "IdUser".
#### One To One
```go
type User struct {
	Id       	uint
	Login    	string
	Name     	string
	Password 	string
}

type UserDetails struct {
	Id       	uint
	Email   	string
	Birthdate 	time.Time
	UserId   	uint // one to one with User
}
```

[Back to Contents](#content)
#### Many To One
**For simplifications all relational slices should be the last fields on struct.**
```go
type User struct {
	Id       	uint
	Name     	string
	Password 	string
	UserLogs 	[]UserLog // one User has many UserLogs
}

type UserLog struct {
	Id       	uint
	Action   	string
	DateTime 	time.Time
	UserId   	uint // if remove the slice from user, will became a one to one
}
```

The difference from one to one and many to one it's the add of a slice field on the "many" struct

[Back to Contents](#content)
#### Many to Many
**For simplifications all relational slices should be the last fields on struct.**
```go
type User struct {
	Id       	uint
	Name     	string
	Password 	string
	UserRoles 	[]UserRole
}

type UserRole struct {
	UserId  	uint `goe:"pk"`
	RoleId  	uint `goe:"pk"`
}

type Role struct {
	Id        	uint
	Name      	string
	UserRoles 	[]UserRole
}
```
Is used a combination of two many to one to generate a many to many. In this example, User has many UserRole and Role has many UserRole.

It's used the tags "pk" for ensure that the foreign keys will be both primary key.

[Back to Contents](#content)

#### Self-Referential

One to Many

```go
type Page struct {
	Id     int
	Number int
	PageId *int
	Pages  []Page
}
```

One to One

```go
type Page struct {
	Id     int
	Number int
	PageId *int
}
```

[Back to Contents](#content)
### Index
#### Unique Index
```go
type User struct {
	Id       	uint
	Name     	string
	Email    	string  `goe:"unique"`
}
```
To create a unique index you need the "unique" goe tag

[Back to Contents](#content)
#### Create Index
```go
type User struct {
	Id       uint
	Name     string
	Email 	 string `goe:"index"`
}
```
To create a common index you need the "index" goe tag

[Back to Contents](#content)
<!-- #### Function Index
```
type User struct {
	Id       uint
	Name     string
	Email    string `goe:"index(n:idx_email f:lower)"`
}
```
> To create a function index you need to pass the "f" parameter with the function name -->
#### Two Columns Index
```go
type User struct {
	Id       uint
	Name    string `goe:"index(n:idx_name_status)"`
	Email   string `goe:"index(n:idx_name_status);unique"`
}
```

Using the goe tag "index()", you can pass the index infos as a function call. "n:" is a parameter for name, to have a two column index just need two indexes with same name. You can use the semicolon ";" to create another single index for the field.

[Back to Contents](#content)

#### Two Columns Unique Index
```go
type User struct {
	Id       uint
	Name    string `goe:"index(unique n:idx_name_status)"`
	Email   string `goe:"index(unique n:idx_name_status);unique"`
}
```

Just as creating a [Two Column Index](#two-columns-index) but added the "unique" value inside the index function.

[Back to Contents](#content)

## Schemas

On GOE it's possible to create schemas by the database struct, all schemas should have the suffix `Schema`
or a tag `goe:"schema"`.

```go
type User struct {
	...
}

type UserRole struct {
	...
}

type Role struct {
	...
}
// schema with suffix Schema
type UserSchema struct {
	User     *User
	UserRole *UserRole
	Role     *Role
}
// schema with any name
type Authentication struct {
	User     *User
	UserRole *UserRole
	Role     *Role
}

type Database struct {
	Status  *Status // status will be on the default schema
	*UserSchema // all structs on UserSchema will be created inside user schema
	*Authentication `goe:"schema"` // will create Authentication schema
	*goe.DB
}
```

> [!TIP]
> On SQLite any schema will be a new attached db file.

[Back to Contents](#content)

## Logging

GOE supports any logger that implements the Logger interface

```go
type Logger interface {
	InfoContext(ctx context.Context, msg string, kv ...any)
	WarnContext(ctx context.Context, msg string, kv ...any)
	ErrorContext(ctx context.Context, msg string, kv ...any)
}
```

The logger is defined on database opening
```go
db, err := goe.Open[Database](sqlite.Open("goe.db", sqlite.Config{
		DatabaseConfig: goe.DatabaseConfig{
			Logger:           slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})),
			IncludeArguments: true,
			QueryThreshold:   time.Second},
	}))
```

> [!TIP]
> You can use slog as your standard logger or make a adapt over the Logger interface.

[Back to Contents](#content)

## Open
To open a database use `goe.Open` function, it's require a valid driver. Most of the drives will require a dns/path connection and a config setup. On `goe.Open` needs to specify the struct database.

If you don't need the database connection anymore, call `goe.Close` to ensure that all the database resources will be removed from memory.

```go
type Database struct {
	Animal         *Animal
	AnimalFood     *AnimalFood
	Food           *Food
	*goe.DB
}

dns := "user=postgres password=postgres host=localhost port=5432 database=postgres"

db, err := goe.Open[Database](postgres.Open(dns, postgres.Config{}))

if err != nil {
	// handler error
}
```

[Back to Contents](#content)
## Migrate

### Auto Migrate

To auto migrate the structs, use the `goe.Migrate(db).AutoMigrate()` passing the database returned by `goe.Open`.

```go
// migrate all database structs
err = goe.Migrate(db).AutoMigrate()
if err != nil {
	// handler error
}
```
[Back to Contents](#content)
### Drop and Rename

```go
type Select struct {
	ID   int
	Name string
}

type Database struct {
	Select         *Select
	*goe.DB
}

err = goe.Migrate(db).OnTable("Select").RenameColumn("Name", "NewName")
if err != nil {
	// handler error
}

err = goe.Migrate(db).OnTable("Select").DropColumn("NewName")
if err != nil {
	// handler error
}

err = goe.Migrate(db).OnTable("Select").RenameTable("NewSelect")
if err != nil {
	// handler error
}

err = goe.Migrate(db).OnTable("NewSelect").DropTable()
if err != nil {
	// handler error
}
```
[Back to Contents](#content)
### Migrate to a SQL file

GOE drivers supports a output migrate path to specify a directory to store the generated SQL. In this way, calling the "AutoMigrate" function goe WILL NOT auto apply the migrations and output the result as a sql file in the specified path.

```go
// open the database with the migrate path config setup
db, err := goe.Open[Database](sqlite.Open("goe.db", sqlite.Config{
	MigratePath: "migrate/",
}))
if err != nil {
	// handler error
}

// AutoMigrate will output the result as a sql file, and not auto apply the migration
err = goe.Migrate(db).AutoMigrate()
if err != nil {
	// handler error
}
```

In this example the file will be output in the "migrate/" path, as follow:

```
📂 migrate
|   ├── SQLite_1760042267.sql
go.mod
```

> [!TIP]
> Any other migration like "DropTable", "RenameColumn" and others... will have the same result as "AutoMigrate", and will generate the SQL file.

[Back to Contents](#content)
## Select
### Find
Find is used when you want to return a single result.
```go
// one primary key
animal, err = goe.Find(db.Animal).ById(Animal{Id: 2})

// two primary keys
animalFood, err = goe.Find(db.AnimalFood).ById(AnimalFood{IdAnimal: 3, IdFood: 2})

// find record by value, if have more than one it will returns the first
cat, err = goe.Find(db.Animal).ByValue(Animal{Name: "Cat"})
```

> [!TIP]
> Use **goe.FindContext** for specify a context.

> [!TIP]
> Use **OnErrNotFound** to replace ErrNotFound with a new error.

[Back to Contents](#content)
### List

List has support for [OrderBy](#orderby), [Pagination](#pagination) and [Joins](#select-join).

```go
// list all animals
animals, err = goe.List(db.Animal).AsSlice()

// list the animals with name "Cat", Id "3" and IdHabitat "4"
animals, err = goe.List(db.Animal).Filter(Animal{Name: "Cat", Id: 3, IdHabitat: 4}).AsSlice()

// when using % on filter, goe makes a like operation
animals, err = goe.List(db.Animal).Filter(Animal{Name: "%Cat%"}).AsSlice()
```

> [!TIP]
> Use **goe.ListContext** for specify a context.

[Back to Contents](#content)
### Select From

Return all animals as a slice
```go
// select * from animals
animals, err = goe.List(db.Animal).AsSlice()

if err != nil {
	// handler error
}
```

> [!TIP]
> Use **goe.SelectContext** for specify a context.

[Back to Contents](#content)

### Select Iterator

Iterate over the rows
```go
for row, err := range goe.List(db.Animal).Rows() {
	// iterator rows
 }
```

[Back to Contents](#content)

### Select Specific Fields
```go
var result []struct {
	User    string
	Role    *string
	EndTime *time.Time
}

// row is the generic struct
for row, err := range goe.Select[struct {
		User    string     // output row
		Role    *string    // output row
		EndTime *time.Time // output row
	}](&struct {
		User    *string     // table column
		Role    *string     // table column
		EndTime **time.Time // table column
}{
		User:    &db.User.Name,
		Role:    &db.Role.Name,
		EndTime: &db.UserRole.EndDate,
}).
	Joins(
		join.LeftJoin[int](&db.User.Id, &db.UserRole.UserId),
		join.LeftJoin[int](&db.UserRole.RoleId, &db.Role.Id),
	).
	OrderByAsc(&db.User.Id).Rows() {

	if err != nil {
		//handler error
	}
	//handler rows
	result = append(result, row)
}
```

For specific field is used a new struct, each new field guards the reference for the database attribute.

[Back to Contents](#content)

### Where
For where, goe uses a sub-package where, on where package you have all the goe available where operations.
```go
animals, err = goe.List(db.Animal).Where(where.Equals(&db.Animal.Id, 2)).AsSlice()

if err != nil {
	//handler error
}
```

It's possible to group a list of where operations inside Where()

```go
animals, err = goe.List(db.Animal).Where(
					where.And(
						where.LessEquals(&db.Animal.Id, 2), 
						where.In(&db.Animal.Name, []string{"Cat", "Dog"}),
					),
				).AsSlice()

if err != nil {
	//handler error
}
```

You can use a if to call a where operation only if it's match
```go
selectQuery := goe.List(db.Animal).Where(where.LessEquals(&db.Animal.Id, 30))

if filter.In {
	selectQuery = selectQuery.Where(
		where.And(
			where.LessEquals(&db.Animal.Id, 30), 
			where.In(&db.Animal.Name, []string{"Cat", "Dog"}),
		),
	)
}

animals, err = selectQuery.AsSlice()

if err != nil {
	//handler error
}
```

It's possible to use a query inside a `where.In`

```go
// use AsQuery() for get a result as a query
querySelect := goe.Select[any](&struct{ Name *string }{Name: &db.Animal.Name}).
					Joins(
						join.Join[int](&db.Animal.Id, &db.AnimalFood.IdAnimal),
						join.Join[uuid.UUID](&db.AnimalFood.IdFood, &db.Food.Id)).
					Where(
						where.In(&db.Food.Name, []string{foods[0].Name, foods[1].Name})).
					AsQuery()

// where in with another query
a, err := goe.List(db.Animal).Where(where.In(&db.Animal.Name, querySelect)).AsSlice()

if err != nil {
	//handler error
}
```
On where, GOE supports operations on two columns, all where operations that have `Arg` as suffix it's used for operation on columns.

In the example, the operator greater (>) on the columns Score and Minimum is used to return all exams that have a score greater than the minimum.

```go
err = goe.List(db.Exam).
	Where(where.GreaterArg[float32](&db.Exam.Score, &db.Exam.Minimum)).AsSlice()
```


[Back to Contents](#content)

### Join
On join, goe uses a sub-package join, on join package you have all the goe available join operations.

For the join operations, you need to specify the type, this make the joins operations more safe. So if you change a type from a field, the compiler will throw a error.
```go
animals, err = goe.List(db.Animal).
			   Joins(
					join.Join[int](&db.Animal.Id, &db.AnimalFood.IdAnimal),
					join.Join[uuid.UUID](&db.Food.Id, &db.AnimalFood.IdFood),
			   ).AsSlice()

if err != nil {
	//handler error
}
```

Same as where, you can use a if to only make a join if the condition match.

[Back to Contents](#content)
### OrderBy
For OrderBy you need to pass a reference to a mapped database field.

It's possible to OrderBy desc and asc. List and Select has support for OrderBy queries.
#### List
```go
animals, err = goe.List(db.Animal).OrderByDesc(&db.Animal.Id).AsSlice()

if err != nil {
	//handler error
}
```
#### Select
```go
animals, err = goe.List(db.Animal).OrderByAsc(&db.Animal.Id).AsSlice()

if err != nil {
	//handler error
}
```

[Back to Contents](#content)
### Pagination
For pagination, it's possible to run on Select and List functions

#### Select Pagination
```go
// page 1 of size 10
page, err = goe.List(db.Animal).AsPagination(1, 10)

if err != nil {
	//handler error
}
```

#### List Pagination
```go
// page 1 of size 10
page, err = goe.List(db.Animal).AsPagination(1, 10)

if err != nil {
	//handler error
}
```

> [!NOTE]
> AsPagination default values for page and size are 1 and 10 respectively.

[Back to Contents](#content)
### Aggregates
For aggregates goe uses a sub-package aggregate, on aggregate package you have all the goe available aggregates. 

On select fields, goe uses query sub-package for declaring a aggregate field on struct.

```go
result, err := goe.Select[struct {
					query.Count
				}](&struct{ 
					*query.Count 
				}{
					aggregate.Count(&db.Animal.Id)
				}).AsSlice()

if err != nil {
	// handler error
}

// count value as int64
result[0].Value
```

[Back to Contents](#content)
### Functions
For functions goe uses a sub-package function, on function package you have all the goe available functions. 

On select fields, goe uses query sub-package for declaring a function result field on struct.
```go
for row, err := range goe.Select[struct {
					UpperName query.Function[string]
				}](&struct {
					UpperName *query.Function[string]
				}{
					UpperName: function.ToUpper(&db.Animal.Name),
				}).Rows() {
					if err != nil {
						//handler error
					}
					//function result value
					row.UpperName.Value
				}
```

Functions can be used inside where.
```go
animals, err = goe.List(db.Animal).
Where(
	where.Like(function.ToUpper(&db.Animal.Name), "%CAT%")
).AsSlice()

if err != nil {
	//handler error
}
```

> [!NOTE] 
> where like expected a second argument always as string.

```go
animals, err = goe.List(db.Animal).
			   Where(
					where.Equals(function.ToUpper(&db.Animal.Name), function.Argument("CAT")),
			   ).AsSlice()

if err != nil {
	//handler error
}
```

> [!IMPORTANT]
> to by pass the compiler type warning, use function.Argument. This way the compiler will check the argument value.

[Back to Contents](#content)
## Insert
On Insert if the primary key value is auto-increment, the new Id will be stored on the object after the insert.

### Insert One
```go
a := Animal{Name: "Cat", Emoji: "🐘"}
err = goe.Insert(db.Animal).One(&a)

if err != nil {
	//handler error
}

// new generated id
a.Id
```

> [!TIP] 
> Use **goe.InsertContext** for specify a context.

[Back to Contents](#content)
### Insert Batch
```go
foods := []Food{
		{Name: "Meat", Emoji: "🥩"},
		{Name: "Hotdog", Emoji: "🌭"},
		{Name: "Cookie", Emoji: "🍪"},
	}
err = goe.Insert(db.Food).All(foods)

if err != nil {
	//handler error
}
```

> [!TIP] 
> Use **goe.InsertContext** for specify a context.

[Back to Contents](#content)
## Update
### Save
Save is the basic function for updates a single record; 
only updates the non-zero values.
```go
a := Animal{Id: 2}
a.Name = "Update Cat"

// update animal of id 2
err = goe.Save(db.Animal).ByValue(a)

if err != nil {
	//handler error
}
```

> [!TIP] 
> Use **goe.SaveContext** for specify a context.

[Back to Contents](#content)

### Update Set
Update with set uses update sub-package. This is used for more complex updates, like updating a field with zero/nil values or make a batch update.

```go
a := Animal{Id: 2}

// a.IdHabitat is nil, so is ignored by Save
err = goe.Update(db.Animal).
	  Sets(update.Set(&db.Animal.IdHabitat, a.IdHabitat)).
	  Where(where.Equals(&db.Animal.Id, a.Id))

if err != nil {
	//handler error
}
```

Check out the [Where](#where) section for more information about where operations.

> [!CAUTION]
> The where call ensures that only the matched rows will be updated.

> [!TIP] 
> Use **goe.UpdateContext** for specify a context.

[Back to Contents](#content)
## Delete
### Remove
Remove is used for remove only one record by primary key
```go
// remove animal of id 2
err = goe.Remove(db.Animal).ById(Animal{Id: 2})

if err != nil {
	//handler error
}
```

> [!TIP] 
> Use **goe.RemoveContext** for specify a context.

[Back to Contents](#content)

### Delete Batch
Delete all records from Animal
```go
err = goe.Delete(db.Animal).All()

if err != nil {
	//handler error
}
```

Delete all matched records
```go
err = goe.Delete(db.Animal).Where(where.Like(&db.Animal.Name, "%Cat%"))

if err != nil {
	//handler error
}
```

Check out the [Where](#where) section for more information about where operations.

> [!CAUTION]
> The where call ensures that only the matched rows will be deleted.

> [!TIP]
> Use **goe.DeleteContext** for specify a context.

[Back to Contents](#content)

## Transaction

### Begin Transaction
Setup the transaction with the database function `db.NewTransaction()`
```go
tx, err = db.NewTransaction()
if err != nil {
	// handler error
}
defer tx.Rollback()
```

You can use the `OnTransaction()` function to setup a transaction for [Select](#select), [Insert](#insert), [Update](#update) and [Delete](#delete).

> [!TIP]
> Ensure to call `defer tx.Rollback()`; this will make the Rollback happens if something go wrong

> [!TIP]
> Use **goe.NewTransactionContext** for specify a context

[Back to Contents](#content)

### Commit and Rollback

To Commit a Transaction just call `tx.Commit()`
```go
err = tx.Commit()

if err != nil {
	// handler the error
}
```

To Rollback a Transaction just call `tx.Rollback()`
```go
err = tx.Rollback()

if err != nil {
	// handler the error
}
```

[Back to Contents](#content)

### Isolation

The isolation is used for control the flow and security of  multiple transactions. On goe you can use the [sql.IsolationLevel](https://pkg.go.dev/database/sql#IsolationLevel).

By default if you call `db.NewTransaction()` it's use the Serializable isolation.

[Back to Contents](#content)

## Benchmarks

Source code of benchmarks can be find on [lauro-santana/go-orm-benchmarks](https://github.com/lauro-santana/go-orm-benchmarks). The benchmarks will be update if any new feature is released for any of these packages. You are welcome to notice me if are anything new or wrong with the benchmarks on the repository.

## Select

### Select One
Package | Avg n/op | Avg ns/op | Avg B/op | Avg allocs/op
| --- | --- | --- | --- | --- |
**pgx** | 20655 | 57493 | 995 | 18
**sqlc** | 21656 | 57810 | 995 | 18
**database/sql** | 19466 | 61851 | 1720 | 47
**goe** | 17270 | 70941 | 4552 | 70
**ent** | 15342 | 76627 | 5152 | 135
**bun** | 12975 | 92397 | 5657 | 23
**gorm** | 10000 | 110914 | 5674 | 110

### Select Cursor Pagination
| Package | Avg n/op | Avg ns/op | Avg B/op | Avg allocs/op |
| --- | --- | --- | --- | --- |
| **pgx** | 1900.0 | 638138.0 | 52602.5 | 730.0 |
| **database/sql** | 1754.0 | 681157.5 | 57922.8 | 1360.0 |
| **sqlc** | 1666.0 | 706612.8 | 79800.8 | 770.0 |
| **goe** | 1586.0 | 777239.3 | 73694.5 | 1130.0 |
| **ent** | 1370.0 | 884408.0 | 111683.0 | 3010.0 |
| **bun** | 1053.8 | 1177303.5 | 90663.5 | 1170.0 |
| **gorm** | 957.0 | 1219613.0 | 100829.3 | 2580.0 |

[Back to Contents](#content)

## Insert

### Insert One Record
| Package | n/op | ns/op | B/op | allocs/op |
| --- | --- | --- | --- | --- |
| **sqlc** | 2806.6 | 435619.4 | 288.0 | 8.0 |
| **pgx** | 2649.4 | 437431.0 | 304.0 | 9.0 |
| **database/sql** | 2546.6 | 437955.8 | 576.0 | 11.0 |
| **goe** | 2673.0 | 450239.6 | 2734.2 | 40.0 |
| **ent** | 2561.2 | 470270.8 | 3943.6 | 90.0 |
| **bun** | 2400.0 | 484351.0 | 5090.4 | 15.0 |
| **gorm** | 2464.4 | 487071.6 | 6110.2 | 91.0 |

### Insert 2000 Bulk Records
| Package | Avg n/op | Avg ns/op | Avg B/op | Avg allocs/op |
| --- | --- | --- | --- | --- |
| **pgx** | 127.5 | 9032688.5 | 335594.5 | 2047.0 |
| **sqlc** | 116.5 | 9784080.0 | 675911.0 | 12048.5 |
| **goe** | 73.0 | 15777915.5 | 6059695.5 | 39879.0 |
| **database/sql** | 70.0 | 16383613.5 | 6054050.5 | 39833.5 |
| **ent** | 56.0 | 18343069.5 | 9805318.5 | 80389.0 |
| **gorm** | 42.0 | 28939012.0 | 5829835.5 | 95750.0 |
| **bun** | 39.5 | 28704378.5 | 1755007.5 | 8036.5 |

database/sql has to append the strings and values by a external function, so you could make a optimized database/sql usage for insert bulk to get a result equivalent to pgx or sqlc.

[Back to Contents](#content)

## Update

### Update One
| Package | Avg n/op | Avg ns/op | Avg B/op | Avg allocs/op |
| --- | --- | --- | --- | --- |
| **pgx** | 2666.5 | 453922.5 | 329.25 | 9.0 |
| **database/sql** | 2533.0 | 455223.0 | 632.0 | 11.0 |
| **sqlc** | 2707.75 | 466297.0 | 313.5 | 8.0 |
| **goe** | 2421.25 | 488427.0 | 3108.25 | 37.0 |
| **bun** | 2263.25 | 506721.75 | 4789.0 | 6.0 |
| **gorm** | 2491.5 | 541869.25 | 7489.0 | 99.0 |
| **ent** | 1782.75 | 649685.0 | 7462.0 | 183.0 |

[Back to Contents](#content)

## Delete

### Delete One

| Package | n/op | ns/op | B/op | allocs/op |
| --- | --- | --- | --- | --- |
| **sqlc** | 2537.8 | 447178.6 | 112.0 | 3.2 |
| **pgx** | 2640.0 | 456962.2 | 127.0 | 4.0 |
| **database/sql** | 2548.2 | 464805.0 | 206.8 | 6.0 |
| **goe** | 2522.6 | 469782.4 | 1531.0 | 20.0 |
| **ent** | 2621.6 | 470532.0 | 1808.4 | 40.0 |
| **bun** | 2488.0 | 486804.6 | 4668.6 | 5.0 |
| **gorm** | 2359.4 | 494844.2 | 3075.8 | 44.0 |

[Back to Contents](#content)
