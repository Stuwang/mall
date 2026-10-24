package main

import (
	"flag"

	"mall/im/common/config"
	"mall/im/internal/app"
)

func main() {
	confPath := flag.String("conf", "configs/config.yaml", "config path")
	flag.Parse()

	cfg, err := config.Load(*confPath)
	if err != nil {
		panic(err)
	}
	app.Run(cfg)
}
