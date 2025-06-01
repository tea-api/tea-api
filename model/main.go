package model

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"tea-api/common"
	"tea-api/constant"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var groupCol string
var keyCol string

func initCol() {
	if common.UsingPostgreSQL {
		groupCol = `"group"`
		keyCol = `"key"`

	} else {
		groupCol = "`group`"
		keyCol = "`key`"
	}
}

var DB *gorm.DB

var LOG_DB *gorm.DB

func createRootAccountIfNeed() error {
	var user User
	//if user.Status != common.UserStatusEnabled {
	if err := DB.First(&user).Error; err != nil {
		common.SysLog("no user exists, create a root user for you: username is root, password is 123456")
		hashedPassword, err := common.Password2Hash("123456")
		if err != nil {
			return err
		}
		rootUser := User{
			Username:    "root",
			Password:    hashedPassword,
			Role:        common.RoleRootUser,
			Status:      common.UserStatusEnabled,
			DisplayName: "Root User",
			AccessToken: nil,
			Quota:       100000000,
		}
		DB.Create(&rootUser)
	}
	return nil
}

func CheckSetup() {
	setup := GetSetup()
	if setup == nil {
		// No setup record exists, check if we have a root user
		if RootUserExists() {
			common.SysLog("system is not initialized, but root user exists")
			// Create setup record
			newSetup := Setup{
				Version:       common.Version,
				InitializedAt: time.Now().Unix(),
			}
			err := DB.Create(&newSetup).Error
			if err != nil {
				common.SysLog("failed to create setup record: " + err.Error())
				constant.Setup = false
			} else {
				common.SysLog("setup record created successfully")
				constant.Setup = true
			}
		} else {
			common.SysLog("system is not initialized and no root user exists")
			constant.Setup = false
		}
	} else {
		// Setup record exists, system is initialized
		common.SysLog("system is already initialized at: " + time.Unix(setup.InitializedAt, 0).String())
		common.SysLog("setting constant.Setup to true")
		constant.Setup = true
	}

	// 添加调试日志确认状态
	common.SysLog(fmt.Sprintf("CheckSetup completed: constant.Setup = %v", constant.Setup))
}

func chooseDB(envName string) (*gorm.DB, error) {
	defer func() {
		initCol()
	}()
	dsn := os.Getenv(envName)
	if dsn != "" {
		if strings.HasPrefix(dsn, "postgres://") || strings.HasPrefix(dsn, "postgresql://") {
			// Use PostgreSQL
			common.SysLog("using PostgreSQL as database")
			common.UsingPostgreSQL = true
			return gorm.Open(postgres.New(postgres.Config{
				DSN:                  dsn,
				PreferSimpleProtocol: true, // disables implicit prepared statement usage
			}), &gorm.Config{
				PrepareStmt: true, // precompile SQL
			})
		}
		if strings.HasPrefix(dsn, "local") {
			common.SysLog("SQL_DSN not set, using SQLite as database")
			common.UsingSQLite = true
			return gorm.Open(sqlite.Open(common.SQLitePath), &gorm.Config{
				PrepareStmt: true, // precompile SQL
			})
		}
		// Use MySQL
		common.SysLog("using MySQL as database")
		// check parseTime
		if !strings.Contains(dsn, "parseTime") {
			if strings.Contains(dsn, "?") {
				dsn += "&parseTime=true"
			} else {
				dsn += "?parseTime=true"
			}
		}

		// 添加连接超时和等待超时参数
		if !strings.Contains(dsn, "timeout") {
			dsn += "&timeout=30s"
		}
		if !strings.Contains(dsn, "readTimeout") {
			dsn += "&readTimeout=30s"
		}
		if !strings.Contains(dsn, "writeTimeout") {
			dsn += "&writeTimeout=30s"
		}
		// 增加等待超时设置，防止连接被意外关闭
		if !strings.Contains(dsn, "wait_timeout") {
			dsn += "&wait_timeout=86400" // 24小时
		}
		if !strings.Contains(dsn, "interactive_timeout") {
			dsn += "&interactive_timeout=86400" // 24小时
		}

		common.UsingMySQL = true
		return gorm.Open(mysql.Open(dsn), &gorm.Config{
			PrepareStmt: true, // precompile SQL
		})
	}
	// Use SQLite
	common.SysLog("SQL_DSN not set, using SQLite as database")
	common.UsingSQLite = true
	return gorm.Open(sqlite.Open(common.SQLitePath), &gorm.Config{
		PrepareStmt: true, // precompile SQL
	})
}

func InitDB() (err error) {
	db, err := chooseDB("SQL_DSN")
	if err == nil {
		if common.DebugEnabled {
			db = db.Debug()
		}
		DB = db
		sqlDB, err := DB.DB()
		if err != nil {
			return err
		}

		// 调整连接池参数，确保长连接保持活跃
		sqlDB.SetMaxIdleConns(common.GetEnvOrDefault("SQL_MAX_IDLE_CONNS", 20))
		sqlDB.SetMaxOpenConns(common.GetEnvOrDefault("SQL_MAX_OPEN_CONNS", 100))

		// 设置连接最大生存时间，确保连接及时刷新
		sqlDB.SetConnMaxLifetime(time.Second * time.Duration(common.GetEnvOrDefault("SQL_MAX_LIFETIME", 3600)))
		// 设置空闲连接最大存活时间
		sqlDB.SetConnMaxIdleTime(time.Second * time.Duration(common.GetEnvOrDefault("SQL_MAX_IDLE_TIME", 1800)))

		// 尝试执行ping操作以测试连接
		if err = sqlDB.Ping(); err != nil {
			common.SysError("MySQL connection ping failed: " + err.Error())
			return err
		}

		// 启动MySQL连接保持
		go KeepMySQLAlive()

		if !common.IsMasterNode {
			return nil
		}
		if common.UsingMySQL {
			_, _ = sqlDB.Exec("ALTER TABLE channels MODIFY model_mapping TEXT;") // TODO: delete this line when most users have upgraded

			// 设置MySQL会话变量，增加超时时间
			_, _ = sqlDB.Exec("SET SESSION wait_timeout=86400")
			_, _ = sqlDB.Exec("SET SESSION interactive_timeout=86400")
		}
		common.SysLog("database migration started")
		err = migrateDB()
		return err
	} else {
		common.FatalLog(err)
	}
	return err
}

func InitLogDB() (err error) {
	if os.Getenv("LOG_SQL_DSN") == "" {
		LOG_DB = DB
		return
	}
	db, err := chooseDB("LOG_SQL_DSN")
	if err == nil {
		if common.DebugEnabled {
			db = db.Debug()
		}
		LOG_DB = db
		sqlDB, err := LOG_DB.DB()
		if err != nil {
			return err
		}

		// 调整连接池参数，确保长连接保持活跃
		sqlDB.SetMaxIdleConns(common.GetEnvOrDefault("SQL_MAX_IDLE_CONNS", 20))
		sqlDB.SetMaxOpenConns(common.GetEnvOrDefault("SQL_MAX_OPEN_CONNS", 100))

		// 设置连接最大生存时间，确保连接及时刷新
		sqlDB.SetConnMaxLifetime(time.Second * time.Duration(common.GetEnvOrDefault("SQL_MAX_LIFETIME", 3600)))
		// 设置空闲连接最大存活时间
		sqlDB.SetConnMaxIdleTime(time.Second * time.Duration(common.GetEnvOrDefault("SQL_MAX_IDLE_TIME", 1800)))

		// 尝试执行ping操作以测试连接
		if err = sqlDB.Ping(); err != nil {
			common.SysError("MySQL LOG connection ping failed: " + err.Error())
			return err
		}

		if !common.IsMasterNode {
			return nil
		}

		if common.UsingMySQL {
			// 设置MySQL会话变量，增加超时时间
			_, _ = sqlDB.Exec("SET SESSION wait_timeout=86400")
			_, _ = sqlDB.Exec("SET SESSION interactive_timeout=86400")
		}

		common.SysLog("database migration started")
		err = migrateLOGDB()
		return err
	} else {
		common.FatalLog(err)
	}
	return err
}

func migrateDB() error {
	err := DB.AutoMigrate(&Channel{})
	if err != nil {
		return err
	}
	err = DB.AutoMigrate(&Token{})
	if err != nil {
		return err
	}
	err = DB.AutoMigrate(&User{})
	if err != nil {
		return err
	}
	err = DB.AutoMigrate(&Option{})
	if err != nil {
		return err
	}
	err = DB.AutoMigrate(&Redemption{})
	if err != nil {
		return err
	}
	err = DB.AutoMigrate(&Ability{})
	if err != nil {
		return err
	}
	err = DB.AutoMigrate(&Log{})
	if err != nil {
		return err
	}
	err = DB.AutoMigrate(&Midjourney{})
	if err != nil {
		return err
	}
	err = DB.AutoMigrate(&TopUp{})
	if err != nil {
		return err
	}
	err = DB.AutoMigrate(&QuotaData{})
	if err != nil {
		return err
	}
	err = DB.AutoMigrate(&Task{})
	if err != nil {
		return err
	}
	err = DB.AutoMigrate(&CheckinRecord{})
	if err != nil {
		return err
	}
	err = DB.AutoMigrate(&ChannelStat{})
	if err != nil {
		return err
	}
	err = DB.AutoMigrate(&Setup{})
	common.SysLog("database migrated")
	//err = createRootAccountIfNeed()
	return err
}

func migrateLOGDB() error {
	var err error
	if err = LOG_DB.AutoMigrate(&Log{}); err != nil {
		return err
	}
	return nil
}

func closeDB(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	err = sqlDB.Close()
	return err
}

func CloseDB() error {
	if LOG_DB != DB {
		err := closeDB(LOG_DB)
		if err != nil {
			return err
		}
	}
	return closeDB(DB)
}

var (
	lastPingTime time.Time
	pingMutex    sync.Mutex
)

func PingDB() error {
	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}

	err = sqlDB.Ping()
	if err != nil {
		return err
	}

	// 如果LOG_DB不同于DB，也ping它
	if os.Getenv("LOG_SQL_DSN") != "" && LOG_DB != DB {
		logSqlDB, err := LOG_DB.DB()
		if err != nil {
			return err
		}

		err = logSqlDB.Ping()
		if err != nil {
			return err
		}
	}

	return nil
}

// 定期检查和保持MySQL连接
func KeepMySQLAlive() {
	if !common.UsingMySQL {
		return
	}

	// 每30秒执行一次Ping操作
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		if err := PingDB(); err != nil {
			common.SysError("MySQL连接Ping失败，尝试重新连接: " + err.Error())
			// 尝试重新连接
			sqlDB, err := DB.DB()
			if err != nil {
				common.SysError("获取SQL DB实例失败: " + err.Error())
				continue
			}

			if err = sqlDB.Ping(); err != nil {
				common.SysError("MySQL重新连接失败: " + err.Error())
			} else {
				common.SysLog("MySQL重新连接成功")
			}

			// 如果LOG_DB不同于DB，也检查它
			if os.Getenv("LOG_SQL_DSN") != "" && LOG_DB != DB {
				logSqlDB, err := LOG_DB.DB()
				if err != nil {
					common.SysError("获取LOG SQL DB实例失败: " + err.Error())
					continue
				}

				if err = logSqlDB.Ping(); err != nil {
					common.SysError("MySQL LOG DB重新连接失败: " + err.Error())
				} else {
					common.SysLog("MySQL LOG DB重新连接成功")
				}
			}
		}
	}
}
