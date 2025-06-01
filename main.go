package main

import (
	"embed"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"tea-api/common"
	"tea-api/constant"
	"tea-api/controller"
	"tea-api/middleware"
	"tea-api/model"
	"tea-api/router"
	"tea-api/service"
	"tea-api/setting/operation_setting"

	"github.com/bytedance/gopkg/util/gopool"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"

	_ "net/http/pprof"
)

func migrateOldSQLiteDB() {
	if os.Getenv("SQL_DSN") != "" {
		return
	}
	newPath := strings.SplitN(common.SQLitePath, "?", 2)[0]
	if _, err := os.Stat(newPath); err == nil {
		return
	}
	oldNames := []string{"veloera.db", "one-api.db", "new-api.db"}
	for _, old := range oldNames {
		if _, err := os.Stat(old); err == nil {
			common.SysLog(fmt.Sprintf("detected %s, migrating to %s", old, newPath))
			err = copyFile(old, newPath)
			if err != nil {
				common.SysError("failed to copy old db: " + err.Error())
				return
			}
			db, err := gorm.Open(sqlite.Open(common.SQLitePath), &gorm.Config{})
			if err != nil {
				common.SysError("failed to open migrated db: " + err.Error())
				_ = os.Remove(newPath)
				return
			}
			sqlDB, err := db.DB()
			if err == nil {
				err = sqlDB.Ping()
				sqlDB.Close()
			}
			if err != nil {
				common.SysError("migrated db validation failed: " + err.Error())
				_ = os.Remove(newPath)
				return
			}
			err = os.Rename(old, old+".bak")
			if err != nil {
				common.SysError("failed to rename old db: " + err.Error())
			} else {
				common.SysLog(fmt.Sprintf("renamed %s to %s.bak", old, old))
			}
			return
		}
	}
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := out.Close(); err == nil {
			err = cerr
		}
	}()
	if _, err = io.Copy(out, in); err != nil {
		return err
	}
	return out.Sync()
}

//go:embed web/dist
var buildFS embed.FS

//go:embed web/dist/index.html
var indexPage []byte

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		common.SysLog("Support for .env file is disabled: " + err.Error())
	}

	common.LoadEnv()

	common.SetupLogger()
	migrateOldSQLiteDB()
	common.SysLog("Tea API " + common.Version + " started")
	if os.Getenv("GIN_MODE") != "debug" {
		gin.SetMode(gin.ReleaseMode)
	}
	if common.DebugEnabled {
		common.SysLog("running in debug mode")
	}
	// Initialize SQL Database
	err = model.InitDB()
	if err != nil {
		common.FatalLog("failed to initialize database: " + err.Error())
	}

	model.CheckSetup()

	// Initialize SQL Database
	err = model.InitLogDB()
	if err != nil {
		common.FatalLog("failed to initialize database: " + err.Error())
	}
	defer func() {
		err := model.CloseDB()
		if err != nil {
			common.FatalLog("failed to close database: " + err.Error())
		}
	}()

	// Initialize Redis
	err = common.InitRedisClient()
	if err != nil {
		common.FatalLog("failed to initialize Redis: " + err.Error())
	}

	// Initialize model settings
	operation_setting.InitRatioSettings()
	// Initialize constants
	constant.InitEnv()
	// Initialize options
	model.InitOptionMap()

	service.InitTokenEncoders()

	if common.RedisEnabled {
		// for compatibility with old versions
		common.MemoryCacheEnabled = true
	}
	if common.MemoryCacheEnabled {
		common.SysLog("memory cache enabled")
		common.SysError(fmt.Sprintf("sync frequency: %d seconds", common.SyncFrequency))

		// Add panic recovery and retry for InitChannelCache
		func() {
			defer func() {
				if r := recover(); r != nil {
					common.SysError(fmt.Sprintf("InitChannelCache panic: %v, retrying once", r))
					// Retry once
					_, fixErr := model.FixAbility()
					if fixErr != nil {
						common.SysError(fmt.Sprintf("InitChannelCache failed: %s", fixErr.Error()))
					}
				}
			}()
			model.InitChannelCache()
		}()

		go model.SyncOptions(common.SyncFrequency)
		go model.SyncChannelCache(common.SyncFrequency)
	}

	// 数据看板
	go model.UpdateQuotaData()

	if os.Getenv("CHANNEL_UPDATE_FREQUENCY") != "" {
		frequency, err := strconv.Atoi(os.Getenv("CHANNEL_UPDATE_FREQUENCY"))
		if err != nil {
			common.FatalLog("failed to parse CHANNEL_UPDATE_FREQUENCY: " + err.Error())
		}
		go controller.AutomaticallyUpdateChannels(frequency)
	}
	if os.Getenv("CHANNEL_TEST_FREQUENCY") != "" {
		frequency, err := strconv.Atoi(os.Getenv("CHANNEL_TEST_FREQUENCY"))
		if err != nil {
			common.FatalLog("failed to parse CHANNEL_TEST_FREQUENCY: " + err.Error())
		}
		go controller.AutomaticallyTestChannels(frequency)
	}
	if common.IsMasterNode && constant.UpdateTask {
		gopool.Go(func() {
			controller.UpdateMidjourneyTaskBulk()
		})
		gopool.Go(func() {
			controller.UpdateTaskBulk()
		})
	}
	if os.Getenv("BATCH_UPDATE_ENABLED") == "true" {
		common.BatchUpdateEnabled = true
		common.SysLog("batch update enabled with interval " + strconv.Itoa(common.BatchUpdateInterval) + "s")
		model.InitBatchUpdater()
	}

	if os.Getenv("ENABLE_PPROF") == "true" {
		gopool.Go(func() {
			log.Println(http.ListenAndServe("0.0.0.0:8005", nil))
		})
		go common.Monitor()
		common.SysLog("pprof enabled")
	}

	// Initialize HTTP server
	server := gin.New()
	server.Use(gin.CustomRecovery(func(c *gin.Context, err any) {
		common.SysError(fmt.Sprintf("panic detected: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"message": fmt.Sprintf("Panic detected, error: %v. Please submit a issue here: https://github.com/tea-api/tea-api", err),
				"type":    "new_api_panic",
			},
		})
	}))
	// This will cause SSE not to work!!!
	//server.Use(gzip.Gzip(gzip.DefaultCompression))
	server.Use(middleware.RequestId())
	server.Use(middleware.AbnormalDetection())
	middleware.SetUpLogger(server)
	// Initialize session store
	store := cookie.NewStore([]byte(common.SessionSecret))
	store.Options(sessions.Options{
		Path:     "/",
		MaxAge:   2592000, // 30 days
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})
	server.Use(sessions.Sessions("session", store))

	router.SetRouter(server, buildFS, indexPage)
	var port = os.Getenv("PORT")
	if port == "" {
		port = strconv.Itoa(*common.Port)
	}
	err = server.Run(":" + port)
	if err != nil {
		common.FatalLog("failed to start HTTP server: " + err.Error())
	}
}
