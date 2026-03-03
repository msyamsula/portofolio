package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/msyamsula/portofolio/backend-app/domain/agent/dto"
	"github.com/msyamsula/portofolio/backend-app/domain/agent/handler"
	"github.com/msyamsula/portofolio/backend-app/domain/agent/repository"
	"github.com/msyamsula/portofolio/backend-app/domain/agent/service"
	llmservice "github.com/msyamsula/portofolio/backend-app/infrastructure/llm/service"
	infraLogger "github.com/msyamsula/portofolio/backend-app/infrastructure/telemetry/logger"
)

// @title PostgreSQL CLI Agent
// @version 1.0
// @description Interactive REPL agent for PostgreSQL with natural language to SQL conversion
// @contact.name API Support
// @contact.url https://github.com/msyamsula/portofolio
// @license.name MIT
// @license.url https://opensource.org/licenses/MIT
func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Initialize logger
	initLogger()

	// Print banner
	printBanner()

	// Prompt for credentials
	connCfg, err := promptForCredentials()
	if err != nil {
		infraLogger.Error("failed to get credentials", err, nil)
		os.Exit(1)
	}

	// Prompt for API key (optional)
	apiKey := promptForAPIKey()

	// Initialize components
	agentService := initializeService(apiKey)

	// Connect to database
	db, err := agentService.Connect(ctx, connCfg)
	if err != nil {
		infraLogger.Error("failed to connect to database", err, nil)
		fmt.Printf("\n%s %v\n\n", errorColor("Connection failed:"), err)
		os.Exit(1)
	}
	defer db.Close()

	// Get connection info
	connInfo, err := agentService.GetConnectionInfo(ctx)
	if err != nil {
		infraLogger.Error("failed to get connection info", err, nil)
	} else {
		infraLogger.Info("connected to database", map[string]any{
			"database": connInfo.Database,
			"user":     connInfo.User,
		})
	}

	// Create handler
	agentHandler := handler.New(agentService)
	agentHandler.SetDatabase(connInfo.Database)
	agentHandler.SetUser(connInfo.User)

	// Start REPL in a goroutine
	errChan := make(chan error, 1)
	go func() {
		errChan <- agentHandler.Start(ctx)
	}()

	// Wait for signal or error
	select {
	case <-sigChan:
		fmt.Println("\n\nInterrupted. Exiting...")
		cancel()
	case err := <-errChan:
		if err != nil {
			infraLogger.Error("REPL error", err, nil)
		}
	}
}

// initLogger initializes the logger
func initLogger() {
	infraLogger.Init(context.Background(), infraLogger.Config{
		ServiceName:       "pg-agent",
		CollectorEndpoint: "",
		LogsEnabled:       true,
	})
}

// printBanner prints the application banner
func printBanner() {
	banner := `
╔════════════════════════════════════════════════════════════╗
║                                                              ║
║         PostgreSQL CLI Agent - Interactive REPL                   ║
║                                                              ║
║     Natural Language to SQL Conversion for PostgreSQL           ║
║                                                              ║
╚════════════════════════════════════════════════════════════╝
`
	fmt.Println(banner)
}

// promptForCredentials prompts for database credentials
func promptForCredentials() (dto.ConnectionConfig, error) {
	agentHandler := handler.New(service.New(nil))
	return agentHandler.PromptForCredentials()
}

// promptForAPIKey prompts for OpenAI API key
func promptForAPIKey() string {
	agentHandler := handler.New(service.New(nil))
	return agentHandler.PromptForAPIKey()
}

// initializeService initializes the agent service
func initializeService(apiKey string) service.Service {
	// Create repository (will be set after connection)
	repo := repository.NewPostgresRepository(nil)

	// Create agent service
	agentService := service.New(repo)

	// Create and set LLM service
	llmSvc, err := llmservice.New(apiKey)
	if err != nil {
		infraLogger.WarnError("failed to initialize LLM service", err, nil)
	} else {
		agentService.SetLLMService(llmSvc)
		if apiKey != "" {
			infraLogger.Info("LLM service initialized", nil)
		} else {
			infraLogger.Info("LLM service not configured (no API key provided)", nil)
		}
	}

	return agentService
}

// errorColor returns red-colored text for errors
func errorColor(text string) string {
	return fmt.Sprintf("\033[31m%s\033[0m", text)
}
