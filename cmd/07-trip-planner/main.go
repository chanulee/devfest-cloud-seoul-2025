package main

import (
	"context"
	"log"
	"os"

	"google.golang.org/adk/agent"
	"google.golang.org/adk/agent/llmagent"
	"google.golang.org/adk/agent/workflowagents/parallelagent"
	"google.golang.org/adk/agent/workflowagents/sequentialagent"
	"google.golang.org/adk/cmd/launcher"
	"google.golang.org/adk/cmd/launcher/full"
	"google.golang.org/adk/model/gemini"
	"google.golang.org/adk/session"
	"google.golang.org/adk/tool"
	"google.golang.org/adk/tool/geminitool"
	"google.golang.org/genai"
)

func main() {
	ctx := context.Background()
	// 1. Initialize Model (Use 2.0-flash if 2.5 is not available)
	model, err := gemini.NewModel(ctx, "gemini-3-pro-preview", &genai.ClientConfig{})
	if err != nil {
		log.Fatalf("Failed to create model: %v", err)
	}

	// 2. Define Scouts (Instructions Updated)
	restaurantScout, _ := llmagent.New(llmagent.Config{
		Name:  "RestaurantScout",
		Model: model,
		// 변경: 입력에서 도시를 추출하도록 명시
		Instruction: `You are a Restaurant Scout. 
        The user's request will contain a destination city (e.g., "Plan a trip to Tokyo").
        1. Extract the city name from the request.
        2. IMMEDIATELY use Google Search to find the top 3 restaurants in that city.
        3. Output ONLY a brief list of the restaurants found. Do not ask for clarification.`,
		Tools:     []tool.Tool{geminitool.GoogleSearch{}},
		OutputKey: "restaurant_list",
	})

	activityScout, _ := llmagent.New(llmagent.Config{
		Name:  "ActivityScout",
		Model: model,
		// 변경: 입력에서 도시를 추출하도록 명시
		Instruction: `You are an Activity Scout.
        The user's request will contain a destination city (e.g., "Plan a trip to Tokyo").
        1. Extract the city name from the request.
        2. IMMEDIATELY use Google Search to find the top 3 tourist activities in that city.
        3. Output ONLY a brief list of the activities found. Do not ask for clarification.`,
		Tools:     []tool.Tool{geminitool.GoogleSearch{}},
		OutputKey: "activity_list",
	})

	// 3. Parallel Runner
	scouts, _ := parallelagent.New(parallelagent.Config{
		AgentConfig: agent.Config{
			Name:        "CityScouts",
			Description: "Scouts for restaurants and activities in parallel.",
			SubAgents:   []agent.Agent{restaurantScout, activityScout},
		},
	})

	// 4. Itinerary Planner
	planner, _ := llmagent.New(llmagent.Config{
		Name:  "ItineraryPlanner",
		Model: model,
		Instruction: `You are a travel planner. 
    Create a one-day itinerary based on the following research:
    
    Restaurants: {restaurant_list}
    Activities: {activity_list}
    
    Combine them into a logical schedule.`,
	})

	// 5. Sequential Pipeline
	tripPlanner, _ := sequentialagent.New(sequentialagent.Config{
		AgentConfig: agent.Config{
			Name:        "TripPlannerPipeline",
			Description: "Executes scouting and then planning.",
			SubAgents:   []agent.Agent{scouts, planner},
		},
	})

	// 6. Run with Explicit Runner
	sessionService := session.InMemoryService()
	//r, err := runner.New(runner.Config{
	//	AppName:        "TripPlannerApp",
	//	Agent:          tripPlanner,
	//	SessionService: sessionService,
	//})
	//if err != nil {
	//	log.Fatalf("Failed to create runner: %v", err)
	//}

	// Create Session
	//sess, _ := sessionService.Create(ctx, &session.CreateRequest{UserID: "user1", AppName: "TripPlannerApp"})

	config := &launcher.Config{
		AgentLoader:    agent.NewSingleLoader(tripPlanner),
		SessionService: sessionService,
	}

	l := full.NewLauncher()

	if err = l.Execute(ctx, config, os.Args[1:]); err != nil {
		log.Fatalf("Run failed: %v\n\n%s", err, l.CommandLineSyntax())
	}

	//// Handle Input (Args or Default)
	//prompt := "Plan a trip to Tokyo"
	//if len(os.Args) > 1 {
	//	prompt = strings.Join(os.Args[1:], " ")
	//}
	//
	//fmt.Printf(">>> Running Pipeline with Prompt: %q\n", prompt)
	//
	//// Run and Print Events
	//events := r.Run(ctx, "user1", sess.Session.ID(), genai.NewContentFromText(prompt, genai.RoleUser), agent.RunConfig{})
	//
	//for event, err := range events {
	//	if err != nil {
	//		log.Printf("Error: %v", err)
	//		continue
	//	}
	//
	//	// Print intermediate outputs clearly
	//	if event.Author == "RestaurantScout" || event.Author == "ActivityScout" {
	//		fmt.Printf("\n[Scout Result - %s]:\n", event.Author)
	//		for _, part := range event.Content.Parts {
	//			fmt.Print(part.Text)
	//		}
	//	} else if event.Author == "ItineraryPlanner" {
	//		fmt.Printf("\n\n[Final Plan]:\n")
	//		for _, part := range event.Content.Parts {
	//			fmt.Print(part.Text)
	//		}
	//	}
	//}
	//fmt.Println("\n\n<<< Done")
}
