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

// 구조화된 아웃풋 = 스키마
// 설명을 잘 하는 것이 중요
// json으로 구조를 정의를 해줘야

func main() {
	ctx := context.Background()

	model, err := gemini.NewModel(ctx,
		"gemini-3-pro-preview",
		//"gemini-3-pro-preview",
		&genai.ClientConfig{
			APIKey: os.Getenv("GOOGLE_API_KEY"),
		})
	if err != nil {
		log.Fatalf("Failed to create model: %v", err)
	}

	outputSchema := &genai.Schema{
		Type: genai.TypeObject,
		Properties: map[string]*genai.Schema{
			"summary":      {Type: genai.TypeString},
			"action_items": {Type: genai.TypeArray, Items: &genai.Schema{Type: genai.TypeString}},
		},
	}

	routerAgent, err := llmagent.New(llmagent.Config{
		Name:         "root_agent",
		Model:        model,
		Description:  "A helpful agent. Uses a router to route the user questions.",
		Instruction:  "You are a helpful assistant. Answer the user's questions.",
		OutputSchema: outputSchema,
	})

	if err != nil {
		log.Fatalf("Failed to create agent: %v", err)
	}

	config := &launcher.Config{
		AgentLoader: agent.NewSingleLoader(routerAgent),
	}

	l := full.NewLauncher()

	if err = l.Execute(ctx, config, os.Args[1:]); err != nil {
		log.Fatalf("Run failed: %v\n\n%s", err, l.CommandLineSyntax())
	}
}

// tool + output structure --> 1 agent = 1 schema
// 툴 + 스키마 하면 안됨
// 툴로 검색을 하고 다른 에이전트한테 넘기고 걔가 구조화된 아웃풋을 주면 됨
// system eng. = basically layered prompts and configurations
