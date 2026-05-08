package main

import (
	"fmt"
	"log"

	"github.com/LeviyLokotb/fedget/internal/config"
	"github.com/LeviyLokotb/fedget/internal/core"
	osprovider "github.com/LeviyLokotb/fedget/internal/os_provider"
)

func main() {
	conf := config.Config{}
	osp := osprovider.LinuxOsProvider{}

	app := core.NewApp(conf, osp)

	devices, err := app.GetValueDevices()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(devices)
}
