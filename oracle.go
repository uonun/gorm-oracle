package oracle

import (
	_ "github.com/sijms/go-ora/v2"

	"gorm.io/gorm"
)

type Config struct {
	DriverName string
	DSN        string

	SkipInitializeWithVersion     bool
	DefaultStringSize             uint
	DontSupportRenameIndex        bool
	DontSupportRenameColumn       bool
	DontSupportNullAsDefaultValue bool

	serverVersion string
	connPool      gorm.ConnPool

	// DontSupportIdentity 为 true 时表明不支持 IDENTITY 关键字
	// See: https://docs.oracle.com/database/121/DRDAA/migr_tools_feat.htm#DRDAA109
	supportIdentity bool

	// supportOffsetFetch 为 true 时支持 OFFSET ... FETCH ... 子句
	// See:
	// - https://docs.oracle.com/database/121/SQLRF/statements_10002.htm#SQLRF55636
	// - https://support.oracle.com/knowledge/Oracle%20Database%20Products/1600130_1.html#GOAL
	supportOffsetFetch bool
}

func Open(dsn string) gorm.Dialector {
	return &Dialector{Config: (&Config{DSN: dsn}).applyEnv()}
}

func New(config Config) gorm.Dialector {
	return &Dialector{Config: (&config).applyEnv()}
}

func (c *Config) applyEnv() *Config {
	// TODO: apply exists env
	return c
}
