/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"relai/internal/api"
	"relai/internal/config"

	"github.com/alejoacosta74/go-logger"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "relai",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: runRelaiAPI,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

const (
	DefaultConfigFile = "./internal/config/config.yaml"
)

var (
	cfgFile string
)

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", DefaultConfigFile, "config file (default is ./config.yaml)")
}

func runRelaiAPI(cmd *cobra.Command, args []string) {
	cfg, err := config.LoadConfig(cfgFile)
	if err != nil {
		logger.Warnf("Config error: %v", err)
		logger.Info("Continuing with default values")
	}

	logger.SetLevel(cfg.LogLevel)

	// Create a new server instance.
	server := api.NewServer(cfg)

	// Run the server in a goroutine.
	go func() {
		if err := server.Start(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("Server error: %v", err)
		}
	}()

	// Listen for interrupt signals to gracefully shutdown.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit // Block until a signal is received.
	logger.Warn("Server shutdown signal received")
	// Create a deadline for the shutdown.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Attempt graceful shutdown.
	if err := server.Shutdown(ctx); err != nil {
		logger.Fatalf("Server forced to shutdown: %v", err)
	}

	logger.Info("Server shutdown successfully")
}
