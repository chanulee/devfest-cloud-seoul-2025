package main

import (
	"fmt"

	"google.golang.org/adk/tool"
)

// Sentiment Tool
type analyzeSentimentArgs struct {
	Text string `json:"text" jsonschema:"The text to analyze."`
}

func analyzeSentiment(ctx tool.Context, args analyzeSentimentArgs) (string, error) {
	fmt.Printf("[Tool] Analyzing sentiment for: %s\n", args.Text)
	// Simple mock logic
	return "Positive Sentiment", nil
}
