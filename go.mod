module github.com/uonun/gorm-oracle

go 1.15

require (
	github.com/google/uuid v1.3.0
	github.com/pkg/errors v0.9.1
	github.com/sijms/go-ora/v2 v2.5.3
	gorm.io/gorm v1.23.8
)

// replace gorm.io/gorm => ../gorm
// replace github.com/sijms/go-ora/v2 => ../go-ora/v2
