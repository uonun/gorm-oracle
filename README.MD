# 用于 GORM 的 Oracle 驱动

## 使用指引

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
