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
	// 기초가 되는 빈 맥락을 생성함
	ctx := context.Background()

	// 빈 모델 생성
	model, err := gemini.NewModel(ctx,
		"gemini-2.5-flash",
		//"gemini-3-pro-preview",
		&genai.ClientConfig{
			APIKey: os.Getenv("GOOGLE_API_KEY"),
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
		AgentLoader: agent.NewSingleLoader(rootAgent), // launcher라는 개념이 존재함 , 무조건 root agent 하나가 있어야 함
	}

	l := full.NewLauncher() // full은 webui, launcher 다 한번에 실행하게 해 주겠다

	if err = l.Execute(ctx, config, os.Args[1:]); err != nil {
		log.Fatalf("Run failed: %v\n\n%s", err, l.CommandLineSyntax())
	}

}

// go run main.go web api webui
