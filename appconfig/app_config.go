package appconfig

import "github.com/SaiNageswarS/go-api-boot/config"

type AppConfig struct {
	config.BootConfig `ini:",extends"`
	MongoURI          string `ini:"mongo_uri"`
	ProfileBucket     string `ini:"profile_bucket"`
}
