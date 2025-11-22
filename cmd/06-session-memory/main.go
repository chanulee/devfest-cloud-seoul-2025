package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"google.golang.org/adk/agent"
	"google.golang.org/adk/agent/llmagent"
	"google.golang.org/adk/memory"
	"google.golang.org/adk/runner"
	"google.golang.org/adk/tool"
	"google.golang.org/adk/tool/functiontool"

	"google.golang.org/adk/model/gemini"
	"google.golang.org/adk/session"
	"google.golang.org/genai"
)

// --- Tool 정의 ---
type Args struct {
	Query string `json:"query" jsonschema:"The query to search for in the memory."`
}

type Result struct {
	Results []string `json:"results"`
}

// memorySearchToolFunc: 단순 텍스트 검색을 수행하지만, 검색어를 띄어쓰기 단위로 쪼개서 유연하게 찾도록 개선
func memorySearchToolFunc(tctx tool.Context, args Args) (Result, error) {
	fmt.Printf("\n[Tool] 검색어: '%s'", args.Query)

	// 1. 기본 검색 (라이브러리 제공 기능)
	searchResults, err := tctx.SearchMemory(context.Background(), args.Query)
	if err != nil {
		log.Printf("Error searching memory: %v", err)
		return Result{}, fmt.Errorf("failed memory search")
	}

	var results []string
	seen := make(map[string]bool) // 중복 제거용

	for _, res := range searchResults.Memories {
		if res.Content != nil {
			text := strings.Join(textParts(res.Content), " ")
			// 중복된 내용은 제외
			if !seen[text] {
				results = append(results, text)
				seen[text] = true
			}
		}
	}

	if len(results) == 0 {
		fmt.Println(" -> 결과 없음")
		return Result{Results: []string{"No relevant memories found."}}, nil
	}

	fmt.Printf(" -> %d개 찾음\n", len(results))
	return Result{Results: results}, nil
}

var memorySearchTool = must(functiontool.New(
	functiontool.Config{
		Name: "search_past_conversations",
		// 설명(Description)에 한국어 검색을 강조합니다.
		Description: "Searches past conversations. If the user speaks Korean, YOU MUST SEARCH IN KOREAN keywords (e.g., '이름', '좋아하는 것').",
	},
	memorySearchToolFunc,
))

func main() {
	ctx := context.Background()

	// 1. 모델 초기화 (gemini-1.5-flash 사용)
	model, err := gemini.NewModel(ctx, "gemini-2.5-flash", &genai.ClientConfig{})
	if err != nil {
		log.Fatalf("Failed to create model: %v", err)
	}

	// 2. 서비스 초기화 (기본 InMemoryService 사용 - 컴파일 에러 방지)
	sessionService := session.InMemoryService()
	memoryService := memory.InMemoryService()

	// 3. 에이전트 설정 (프롬프트로 언어 문제 해결)
	rootAgent, err := llmagent.New(llmagent.Config{
		Name:  "root_agent",
		Model: model,
		// ★핵심★: 에이전트에게 한국어 검색을 강제하는 Instruction
		Instruction: `You are a helpful assistant with a good memory.
		
		RULES FOR MEMORY:
		1. Always use 'search_past_conversations' tool when the user asks about personal info (name, past topics).
		2. CRITICAL: If the conversation is in Korean, generate the search query IN KOREAN.
		   - Bad Query: "user name"
		   - Good Query: "내 이름", "사용자 이름", "이름은"
		3. If the tool returns the information, answer naturally in Korean.`,
		Tools: []tool.Tool{memorySearchTool},
	})
	if err != nil {
		log.Fatalf("Failed to create agent: %v", err)
	}

	// 4. Runner 생성
	r, err := runner.New(runner.Config{
		AppName:        "MemoryApp",
		Agent:          rootAgent,
		SessionService: sessionService,
		MemoryService:  memoryService,
	})
	if err != nil {
		log.Fatalf("Failed to create runner: %v", err)
	}

	sessionID := "session1"
	userID := "user1"

	must(sessionService.Create(ctx, &session.CreateRequest{
		UserID:    userID,
		AppName:   "MemoryApp",
		SessionID: sessionID,
	}))

	fmt.Println(">>> 봇이 준비되었습니다. (종료: exit)")
	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("\nUser: ")
		if !scanner.Scan() {
			break
		}
		input := scanner.Text()
		if strings.ToLower(input) == "exit" {
			break
		}

		userContent := genai.NewContentFromText(input, genai.RoleUser)
		events := r.Run(ctx, userID, sessionID, userContent, agent.RunConfig{})

		for event, err := range events {
			if err != nil {
				log.Printf("Error: %v", err)
				continue
			}
			if event.Content != nil {
				for _, part := range event.Content.Parts {
					fmt.Print(part.Text)
				}
			}
		}
		fmt.Println()

		// 5. 기억 저장 (업데이트된 세션 가져오기)
		latestSession, err := sessionService.Get(ctx, &session.GetRequest{
			AppName:   "MemoryApp",
			UserID:    userID,
			SessionID: sessionID,
		})
		if err != nil {
			log.Printf("세션 조회 실패: %v", err)
			continue
		}

		if err := memoryService.AddSession(ctx, latestSession.Session); err != nil {
			log.Printf("메모리 저장 실패: %v", err)
		} else {
			fmt.Println("--- [시스템] 기억 저장 완료 ---")
		}
	}
}

func must[T any](obj T, err error) T {
	if err != nil {
		panic(err)
	}
	return obj
}

func textParts(content *genai.Content) []string {
	var texts []string
	if content == nil {
		return texts
	}
	for _, part := range content.Parts {
		if part.Text != "" {
			texts = append(texts, part.Text)
		}
	}
	return texts
}
