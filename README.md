column2struct
====

Generate [gorm.io](https://gorm.io) stucts from MySQL(or MariaDB)'s ```information_schema.columns```.

# Install

```shell
go get github.com/eyasuyuki/column2struct
```

# Usage

```shell
go run github.com/eyasuyuki/column2struct <DSN> [<filename>]
```
DSN 
: data set name

filename
: output file name

## Example

```
go run github.com/eyasuyuki/column2struct 'user:password@tcp(127.0.0.1:3306)/dbname?charset=utf8&parseTime=True&loc=UTC'  ./database/database.go 
```

## Environment variable

|environment variable|default|description|
---|---|---
|PACKAGE_NAME|```"database"```|generated code's package name|
