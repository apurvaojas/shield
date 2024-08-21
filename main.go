package main

import (
	"log"
	"org-forms-config-management/config"
	"org-forms-config-management/infra/database"
	"org-forms-config-management/infra/logger"
	"org-forms-config-management/routers"
	"time"
	"github.com/spf13/viper"

	"org-forms-config-management/docs"
)

func main() {
	docs.SwaggerInfo.Title = "Org Forms Config Management"
	docs.SwaggerInfo.Description = "This is swagger api doc for organic forms config management"
	docs.SwaggerInfo.Version = "1.0"

	// docs.init()
	//set timezone
	viper.SetDefault("SERVER_TIMEZONE", "Asia/kolkata")
	loc, _ := time.LoadLocation(viper.GetString("SERVER_TIMEZONE"))
	time.Local = loc

	if err := config.SetupConfig(); err != nil {
		logger.Fatalf("config SetupConfig() error: %s", err)
	}
	masterDSN, replicaDSN := config.DbConfiguration()

	// Retry connection if error
	maxRetries := 5
	retryInterval := time.Second

	for i := 0; i < maxRetries; i++ {
		if err := database.DBConnection(masterDSN, replicaDSN); err != nil {
			log.Println("database DbConnection error: %s. Retrying in %v...", err, retryInterval)
			time.Sleep(retryInterval)
		} else {
			break
		}
	}

	router := routers.Routes()

	logger.Fatalf("%v", router.Run(config.ServerConfig()))

}
