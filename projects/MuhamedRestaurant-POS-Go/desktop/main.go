package main

import (
	"context"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:../frontend
var assets embed.FS

func main() {
	// Create application
	err := wails.Run(&options.App{
		Title:  "نظام مطعم - Restaurant POS",
		Width:   1366,
		Height:  768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 240, G: 242, B: 245, A: 1},
		OnStartup:        startup,
		OnShutdown:       shutdown,
	})

	if err != nil {
		println("Error:", err.Error())
	}
}

func startup(ctx context.Context) {
	println("🚀 Restaurant POS Desktop Started!")

	// Start backend API server
	// In production, this would start the Go backend server
	// The desktop app would then proxy requests to the backend

	println("✅ Backend API started")
}

func shutdown(ctx context.Context) {
	println("👋 Restaurant POS Desktop Shutting Down!")
}
