package main

import (
	"context"
	"log"
	"os"

	"google.golang.org/adk/agent"
	"google.golang.org/adk/agent/llmagent"
	"google.golang.org/adk/cmd/launcher"
	"google.golang.org/adk/cmd/launcher/full"
	"google.golang.org/adk/model/gemini"
	"google.golang.org/genai"
)

func main() {
	ctx := context.Background()

	model, err := gemini.NewModel(ctx,
		"gemini-3-pro-preview",
		//"gemini-3-pro-preview",
		&genai.ClientConfig{

			APIKey: "AQ.Ab8RN6I2SdICj3Dn_eIk0g1HdXlypmi3JDP9Cl7y_9ju0ofFTw",
		})
	if err != nil {
		log.Fatalf("Failed to create model: %v", err)
	}

	rootAgent, err := llmagent.New(llmagent.Config{
		Name:        "root_agent",
		Model:       model,
		Description: "A helpful agent.",
		Instruction: "You are a helpful assistant. Answer the user's questions.",
	})

	if err != nil {
		log.Fatalf("Failed to create agent: %v", err)
	}

	config := &launcher.Config{
		AgentLoader: agent.NewSingleLoader(rootAgent),
	}

	l := full.NewLauncher()

	if err = l.Execute(ctx, config, os.Args[1:]); err != nil {
		log.Fatalf("Run failed: %v\n\n%s", err, l.CommandLineSyntax())
	}

}
