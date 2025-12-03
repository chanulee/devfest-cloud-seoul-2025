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
	"google.golang.org/adk/tool"
	"google.golang.org/adk/tool/functiontool"

	"google.golang.org/genai"
)

func main() {
	ctx := context.Background()

	model, err := gemini.NewModel(ctx,
		"gemini-2.5-flash",
		//"gemini-3-pro-preview",
		&genai.ClientConfig{
			APIKey: "AIzaSyCd8mFAuCpibAU11AI5WkjsmF1wl_c8u0Y",
		})
	if err != nil {
		log.Fatalf("Failed to create model: %v", err)
	}

	// 1 agent = 1 tool
	// 2 tools mean 2 agents

	weatherTool, _ := functiontool.New(functiontool.Config{ // main에서 이걸 한번 불러줘야 함
		Name: "get_weather", Description: "Get weather for a city"}, // 이 설명이 엄청 상세하게 적혀있어야 함
		getWeather,
	)

	sentimentTool, _ := functiontool.New(
		functiontool.Config{Name: "analyze_sentiment", Description: "Analyze text sentiment"},
		analyzeSentiment)

	myAgent, err := llmagent.New(llmagent.Config{
		Name:  "helper_agent",
		Model: model,
		Instruction: "You are a helper. If asked about weather, use get_weather.  " +
			"Then analyze the user's reaction using analyze_sentiment.",
		Tools: []tool.Tool{weatherTool, sentimentTool},
	})

	if err != nil {
		log.Fatalf("Failed to create agent: %v", err)
	}

	config := &launcher.Config{
		AgentLoader: agent.NewSingleLoader(myAgent),
	}

	l := full.NewLauncher()

	if err = l.Execute(ctx, config, os.Args[1:]); err != nil {
		log.Fatalf("Run failed: %v\n\n%s", err, l.CommandLineSyntax())
	}
}
