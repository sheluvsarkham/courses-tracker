package main

import (
	"embed"
	"io/fs"
	"log"
	"os"
	"path/filepath"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var frontend embed.FS

func main() {
	dataPath, err := courseDataPath()
	if err != nil {
		log.Fatal(err)
	}

	app, err := NewApp(filepath.Dir(dataPath))
	if err != nil {
		log.Fatal(err)
	}
	assets, err := fs.Sub(frontend, "frontend/dist")
	if err != nil {
		log.Fatal(err)
	}

	if err := wails.Run(&options.App{
		Title:       "Трекер курсов",
		Width:       1180,
		Height:      780,
		MinWidth:    900,
		MinHeight:   620,
		AssetServer: &assetserver.Options{Assets: assets},
		BackgroundColour: &options.RGBA{
			R: 247,
			G: 248,
			B: 245,
			A: 1,
		},
		Bind: []interface{}{app},
	}); err != nil {
		log.Fatal(err)
	}
}

func courseDataPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	configPath := filepath.Join(configDir, "CourseTracker", "courses.json")
	if _, err := os.Stat(configPath); err == nil {
		return configPath, nil
	}

	legacyPaths := []string{filepath.Join("data", "courses.json")}
	if executable, err := os.Executable(); err == nil {
		legacyPaths = append(legacyPaths, filepath.Join(filepath.Dir(executable), "data", "courses.json"))
	}
	for _, legacyPath := range legacyPaths {
		data, err := os.ReadFile(legacyPath)
		if err == nil {
			if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
				return "", err
			}
			if err := os.WriteFile(configPath, data, 0o644); err != nil {
				return "", err
			}
			return configPath, nil
		}
		if !os.IsNotExist(err) {
			return "", err
		}
	}
	return configPath, nil
}
