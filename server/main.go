package main

import (
    "log"
    "mask-443/config"
    "mask-443/logger"
    "mask-443/network"
    "os"
    "os/signal"
    "syscall"
)

func main() {
    // Initialize logger
    logger.Init(log.LstdFlags|log.Lshortfile, os.Stdout)
    logger.Info.Println("Starting mask-443 server...")

    // Load configuration
    cfg, err := config.LoadConfig("config.yml") // Assuming YAML for now
    if err != nil {
        logger.Error.Fatalf("Failed to load configuration: %v", err)
    }
    logger.Info.Printf("Configuration loaded: %+v", cfg.Server)

    // Create and start the network server
    server, err := network.NewServer(cfg)
    if err != nil {
        logger.Error.Fatalf("Failed to create server: %v", err)
    }

    go func() {
        if err := server.Start(); err != nil {
            logger.Error.Fatalf("Failed to start server: %v", err)
        }
    }()

    logger.Info.Println("mask-443 server started on port", cfg.Server.ListenPort)

    // Graceful shutdown
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    logger.Info.Println("Shutting down mask-443 server...")
    // Add server shutdown logic if needed server.Stop()
    logger.Info.Println("Server gracefully stopped.")
}
