# 用于 GORM 的 Oracle 驱动

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
  // prepare connection options
  urlOptions := map[string]string{
    "CONNECTION TIMEOUT": "3",
  }
  // build DSN
  // CONN=(DESCRIPTION=(ADDRESS=(PROTOCOL=TCP)(HOST=***)(PORT=***))(CONNECT_DATA=(SERVER=DEDICATED)(SERVICE_NAME=***)))
  // USER= ""
  // PASSWORD = ""
  // oracle://{USER}:{PASSWORD}@:0/?CONNECTION TIMEOUT=3&connStr=CONN_TNS
  dsn = go_ora.BuildJDBC(os.Getenv("USER"), 
    os.Getenv("PASSWORD"), 
    os.Getenv("CONN"), urlOptions)

  // create orm and open connection
  dialector := oracle.Open(dsn)
  db, err := gorm.Open(dialector, &gorm.Config{})

  // use `db` same as gorm
  // ...
}
```

## 已支持的内容

- Connection
  - gorm.Open
- Query
  - db.Raw("").Scan(&model)
  - db.Find(&model)
- Insert
  - db.Exec("INSERT INTO ...", ...)
  - db.Create(&model)
  - db.Create(&multi_models)
- Update
  - // SUPPORT, TO BE TESTED
- Delete
  - db.Delete(&model)
- Transaction
  - db.Begin(), db.Rollback(), db.Commit()
  - db.SavePoint(""), db.RollbackTo("")

## 「已知的」不支持内容

- First、不带条件的 Take 等会生成 LIMIT 子句的方法
- 查询命令的参数传递时，只能按顺序匿名传入，无法按名称传入

## TODO

- [ ] 配置 db.BatchInsertSize > 0 的情况（finisher_api.go:L18）
- [ ] 支持根据 Model 按需 Create，从而无需 Model 包含所有数据库字段
- [ ] 