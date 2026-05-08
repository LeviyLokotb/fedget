package main

import (
	"log"
	"os"

	"github.com/LeviyLokotb/fedget/internal/config"
	"github.com/LeviyLokotb/fedget/internal/core"
	osprovider "github.com/LeviyLokotb/fedget/internal/os_provider"
	"github.com/LeviyLokotb/fedget/internal/ui"
	"github.com/gotk3/gotk3/gtk"
)

func main() {
	// Инициализация GTK
	log.Println("Started")
	gtk.Init(&os.Args)

	// Конфигурация
	cfg := config.Config{}

	// Создание провайдера ОС
	osProv := osprovider.LinuxOsProvider{}

	// Создание ядра приложения
	core := core.NewApp(cfg, osProv)

	// Создание GUI
	log.Println("Create GUI...")
	gui, err := ui.NewGUI(core)
	if err != nil {
		log.Fatal(err)
	}
	gui.Show()

	log.Println("Created")

	// Запуск основного цикла GTK
	log.Println("Starting App")
	gtk.Main()
}
