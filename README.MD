# GORM-ORACLE

用于 GORM 的 Oracle 驱动。

[![Go Report Card](https://goreportcard.com/badge/github.com/uonun/gorm-oracle)](https://goreportcard.com/report/github.com/uonun/gorm-oracle)
[![Go](https://github.com/uonun/gorm-oracle/actions/workflows/go.yml/badge.svg)](https://github.com/uonun/gorm-oracle/actions/workflows/go.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/uonun/gorm-oracle.svg)](https://pkg.go.dev/github.com/uonun/gorm-oracle)
[![MIT license](https://img.shields.io/badge/license-MIT-brightgreen.svg)](https://opensource.org/licenses/MIT)



## 使用指引

```bash
go get github.com/uonun/gorm-oracle
```

```golang
import (
	"..."
	oracle "github.com/uonun/gorm-oracle"
	"..."
)

func main(){
  // build DSN, see: https://github.com/sijms/go-ora
  //  CONN=(DESCRIPTION=(ADDRESS=(PROTOCOL=TCP)(HOST=***)(PORT=***))(CONNECT_DATA=(SERVER=DEDICATED)(SERVICE_NAME=***)))
  //  USER= ""
  //  PASSWORD = ""
  
  // prepare connection options
  urlOptions := map[string]string{
    "CONNECTION TIMEOUT": "3",
  }
  
  // oracle://{USER}:{PASSWORD}@:0/?CONNECTION TIMEOUT=3&connStr=CONN_TNS
  dsn = go_ora.BuildJDBC(os.Getenv("USER"), 
    os.Getenv("PASSWORD"), 
    os.Getenv("CONN"), urlOptions)

  // create with default configuration
  // or create use customer configuration
  // both will be overwritten by Environment variables
  dialector := oracle.Open(dsn)
  // dialector = oracle.New(oracle.Config{
  //   DSN: dsn,
  //   SkipInitializeWithVersion: true,
  //   // DefaultStringSize: ,
  // })

  db, err := gorm.Open(dialector, &gorm.Config{})

  // use `db` same as gorm
  // ...
}
```

## 已支持并测试的内容

### Connection
  - gorm.Open


### Query
  - db.Raw("").Scan(&model)
  - db.Find(&model)

### Insert
  - db.Exec("INSERT INTO ...", ...)
  - db.Create(&model)
  - db.Create(&multi_models)
  
#### Returning

Option 1: tag `autoIncrement` for sequence column to return. For example:

```golang
// Customer table comment
type CustomerReturning struct {
  CustomerID   int64     `gorm:"column:CUSTOMER_ID;sequence:CUSTOMERS_S;autoIncrement" json:"customer_id"`
  // ... other fields
}
```

see: [TestInsertReturningModels](./test/insert_test.go), [CustomerReturning](./test/models.go)

Option 2: use `clause.Returning`.

```golang
db := getDb(t)
cs := make([]Customer, 10) // init 10 instances.
tx := db.Clauses(clause.Returning{ // returning `CUSTOMER_ID`
    Columns: []clause.Column{
      {Name: "CUSTOMER_ID"},
    },
  }).Create(&cs)) // create
```

see: [TestInsertModelsWithReturningClause](./test/insert_test.go)

### Update
  - db.Exec("UPDATE ... SET ...", ...)
  - db.Updates(&model) // single update
  - db.Where("id in ?", ids).Updates(model{}) // batch updates

### Delete
  - db.Delete(&model)

### Transaction
  - db.Begin(), db.Rollback(), db.Commit()
  - db.SavePoint(""), db.RollbackTo("")

## 暂未支持的内容

- 未支持 LIMIT 子句：如 First、不带条件的 Take。
- 未支持命名参数：查询命令的参数传递时，只能按顺序匿名传入，无法按名称传入。
- 有限支持 RowsAffected：包含 RETURNING 行为时，不支持通过 RowsAffected 返回实际的影响行数。
