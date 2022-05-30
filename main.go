package main

import (
	"fmt"
	"github.com/cbroglie/mustache"
	mysql "github.com/go-sql-driver/mysql"
	"github.com/iancoleman/strcase"
	gmysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
	"os"
)

const (
	SQL = `
SELECT
  table_name,
  column_name,
  data_type,
  is_nullable,
  column_key
FROM
  information_schema.columns
WHERE
  table_schema NOT IN ('information_schema','sys','performance_schema','mysql')
  AND table_schema = ?
ORDER BY
  table_name,
  ordinal_position
;
`
	OUTPUT_TEMPATE = `// Code generated by github.com/eyasuyuki/column2struct, DO NOT EDIT.

package {{PackageName}}

import (
	"encoding/base64"
	"strconv"
{{#UseTime}}
	"time"
{{/UseTime}}
)

const TIMESTAMP_PATTERN = "2006-01-02 15:04:05.999999999"
	
{{#Tables}}
// {{StructName}}

type {{StructName}} struct {
	{{#Columns}}
	{{FieldName}} {{GoType}} {{{Tags}}}
	{{/Columns}}
}

func ({{StructName}}) TableName() string {
	return "{{TableName}}"
}

func ({{StructName}}) IdFromBase64(id string) (int64, error) {
	out, err := base64.StdEncoding.DecodeString(id)
	if err != nil {
		panic(any(err))
	}
	strId := out[len("{{StructName}}:"):len(out)]
	return strconv.ParseInt(string(strId), 10, 64)
}

func (m {{StructName}}) Base64Id() string {
	strId := "{{StructName}}:"+strconv.FormatInt(m.Id, 10)
	return base64.StdEncoding.EncodeToString([]byte(strId))
}

{{/Tables}}
`

	PACKAGE_NAME_KEY     = "PACKAGE_NAME"
	DEFAULT_PACKAGE_NAME = "database"
)

var DATA_MAP = map[string]string{
	"bigint":     "int64",
	"blob":       "[]byte",
	"char":       "string",
	"date":       "time.Time",
	"datetime":   "time.Time",
	"decimal":    "float64",
	"double":     "float64",
	"enum":       "string",
	"float":      "float",
	"int":        "int",
	"longblob":   "[]byte",
	"longtext":   "string",
	"mediumtext": "string",
	"set":        "string",
	"smallint":   "int",
	"text":       "string",
	"time":       "time.Time",
	"timestamp":  "time.Time",
	"tinyint":    "int",
	"varbinary":  "[]byte",
	"varchar":    "string",
}

type Column struct {
	TableName  string
	ColumnName string
	DataType   string
	IsNullable string
	ColumnKey  string
}

type Table struct {
	TableName string
	Columns   []Column
}

type Tables struct {
	Tables      []Table
	PackageName string
	UseTime     bool
}

func (t Table) StructName() string {
	return strcase.ToCamel(t.TableName)
}

func (c Column) FieldName() string {
	return strcase.ToCamel(c.ColumnName)
}

func (c Column) Tags() string {
	if c.ColumnKey == "PRI" {
		return "`gorm:\"primaryKey;autoIncrement:false\"`"
	} else {
		return ""
	}
}

func (c Column) GoType() string {
	if c.isNullable() {
		return "*" + DATA_MAP[c.DataType]
	} else {
		return DATA_MAP[c.DataType]
	}
}

func (c Column) UseTime() bool {
	if c.GoType() == "time.Time" {
		return true
	} else {
		return false
	}
}

// utility method
func (c *Column) isNullable() bool {
	return c.IsNullable == "YES"
}

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("%s <DSN> [<filepath>]\n", os.Args[0])
		os.Exit(1)
	}

	// read from information_schema.column
	cfg, err := mysql.ParseDSN(os.Args[1])
	if err != nil {
		log.Fatalf(err.Error())
	}

	db, err := gorm.Open(gmysql.Open(cfg.FormatDSN()), &gorm.Config{})
	if err != nil {
		log.Fatalf(err.Error())
	}
	defer func() {
		if db != nil {
			sqlDb, err := db.DB()
			if err != nil {
				panic(any(err))
			}
			sqlDb.Close()
		}
	}()

	// query
	var columns []Column
	if err = db.Raw(SQL, cfg.DBName).Scan(&columns).Error; err != nil {
		log.Fatalf(err.Error())
	}

	// get package name
	packageName := os.Getenv(PACKAGE_NAME_KEY)
	if packageName == "" {
		packageName = DEFAULT_PACKAGE_NAME
	}

	// set tables
	tables := &Tables{PackageName: packageName}
	var table *Table
	var prevTableName string
	for _, column := range columns {
		if column.TableName == "flyway_schema_history" {
			continue
		}
		if column.TableName != prevTableName {
			if table != nil {
				tables.Tables = append(tables.Tables, *table)
			}
			table = &Table{TableName: column.TableName}
			prevTableName = column.TableName
		}
		table.Columns = append(table.Columns, column)
		if column.UseTime() {
			tables.UseTime = true
		}
	}
	tables.Tables = append(tables.Tables, *table)

	//debug
	//json, _ := json.Marshal(tables)
	//log.Printf("%v \n", string(json))

	output, _ := mustache.Render(OUTPUT_TEMPATE, tables)

	// write
	writer := os.Stdout
	if len(os.Args) > 2 {
		file, err := os.Create(os.Args[2])
		if err != nil {
			log.Fatalf(err.Error())
		}
		writer = file
	}
	writer.Write([]byte(output))
	defer writer.Close()
}
