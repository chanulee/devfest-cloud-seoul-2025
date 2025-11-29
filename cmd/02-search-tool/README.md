# Go ADK 핸즈온 Session 2: 도구(Tools)를 활용한 검색 에이전트 만들기

환영합니다! 👋 두 번째 세션에 오신 것을 환영합니다.

지난 세션에서 우리는 기본적인 대화형 에이전트를 만들었습니다. 하지만 그 에이전트는 학습된 시점 이후의 정보나 실시간 뉴스는 알지 못한다는 한계가 있었죠.

이번 세션에서는 ADK의 강력한 기능인 **Tools(도구)** 시스템을 사용하여, 에이전트에게 **Google Search** 능력을 부여해 보겠습니다. 이제 여러분의 에이전트는 최신 정보를 검색하여 답변할 수 있게 됩니다.

### 🎯 학습 목표
*   **ADK Tool Interface** 이해하기
*   `geminitool` 패키지를 사용하여 Google Search 기능 연동하기
*   실시간 정보가 필요한 질문에 답변하는 에이전트 구현하기

### 💻 코드 상세 분석

이번 코드는 세션 1과 구조가 비슷하지만, **Tools** 설정 부분이 추가되었습니다. 변경된 부분을 중점적으로 살펴보겠습니다.

### 1. 도구(Tool) 관련 패키지 추가
```go
import (
    // ... 기존 import 생략 ...
	"google.golang.org/adk/tool"
	"google.golang.org/adk/tool/geminitool"
    // ...
)
```
*   `google.golang.org/adk/tool`: ADK에서 도구를 정의하고 관리하는 인터페이스입니다.
*   `google.golang.org/adk/tool/geminitool`: Gemini 모델이 사용할 수 있는 사전 정의된 도구 모음(예: 구글 검색, 코드 실행 등)입니다.

### 2. 모델 초기화 (동일)
```go
	model, err := gemini.NewModel(ctx,
		"gemini-3-pro-preview",
		&genai.ClientConfig{
			APIKey: os.Getenv("GOOGLE_API_KEY"),
		})
    // ...
```
*   이전과 동일하게 Gemini 모델을 초기화합니다. Google Search 기능은 `gemini-2.5` 이상의 모델들에서 매우 효과적으로 작동합니다.

### 3. 에이전트에 도구 장착하기 (핵심 변경 사항) ⭐
가장 중요한 부분입니다. 에이전트를 생성할 때 `Tools` 옵션을 추가합니다.

```go
	// 변수명은 timeAgent지만, 실제 역할은 검색 에이전트입니다.
	timeAgent, err := llmagent.New(llmagent.Config{
		Name:        "search_agent", // 에이전트 이름 변경
		Model:       model,
		Description: "A helpful agent that searches the web.", // 설명 업데이트
		Instruction: "You are a helpful assistant. Use Google Search to answer the user's questions.", // 검색을 활용하라고 지시
		
        // [핵심] 도구 목록 정의
		Tools: []tool.Tool{
			geminitool.GoogleSearch{}, // Google 검색 도구 추가
		},
	})
```

*   **`Tools: []tool.Tool{...}`**: 에이전트가 사용할 수 있는 도구들의 목록입니다.
*   **`geminitool.GoogleSearch{}`**: 별도의 복잡한 구현 없이 이 한 줄만으로 에이전트는 Google 검색 엔진(Grounding with Google Search)을 사용할 수 있는 능력을 갖게 됩니다.
*   **`Instruction`**: 프롬프트에 "Google Search를 사용해라"라고 명시해 주면, 모델이 언제 도구를 사용해야 할지 더 잘 판단합니다.

### 4. 런처 실행 (동일)
```go
	config := &launcher.Config{
		AgentLoader: agent.NewSingleLoader(timeAgent),
	}
    // ... 실행 로직 동일
```

---

### 🚀 실행 및 테스트 (Let's Run!)

코드를 저장하고 터미널에서 실행해 봅시다. 이번에는 실시간 정보가 필요한 질문을 던져보는 것이 중요합니다.

### 1. 대화 모드 실행
```bash
go run main.go chat
```

### 2. 질문 예시 (비교 체험)

**Q1. (과거 지식) "미국의 수도는 어디야?"**
*   모델이 이미 알고 있는 지식이므로 검색 없이 바로 대답할 수도 있습니다.

**Q2. (실시간 정보) "어제 손흥민 경기 결과 알려줘" 또는 "오늘 서울 날씨 어때?"**
*   **Session 1의 에이전트**: "죄송합니다. 저는 실시간 정보에 접근할 수 없습니다."라고 답했을 것입니다.
*   **Session 2의 에이전트**: 내부적으로 Google 검색을 수행(Grounding)하고, 최신 정보를 바탕으로 답변을 생성합니다.

---

### 🔍 무엇이 일어난 건가요? (Under the Hood)

1.  사용자가 **"오늘 서울 날씨 어때?"**라고 묻습니다.
2.  에이전트(LLM)는 자신이 가진 지식으로는 이 답을 알 수 없다고 판단합니다.
3.  하지만 `geminitool.GoogleSearch`라는 도구가 있다는 것을 알고 있습니다.
4.  에이전트는 스스로 **"서울 날씨"**라는 검색 쿼리를 생성하여 도구를 호출합니다.
5.  Google 검색 결과가 에이전트에게 전달됩니다.
6.  에이전트는 검색 결과를 요약하여 사용자에게 자연스러운 답변으로 전달합니다.

이 모든 과정이 ADK와 Gemini 모델 사이에서 자동으로 처리됩니다!

---

### 💡 참고 사항

*   **Grounding**: 이렇게 LLM이 외부 데이터(검색 결과 등)에 기반하여 답변하는 것을 **그라운딩(Grounding)**이라고 합니다. 이를 통해 할루시네이션(거짓 답변)을 줄이고 신뢰성을 높일 수 있습니다.
*   **비용**: 검색 도구를 사용하면 일반적인 텍스트 생성 외에 검색에 대한 추가적인 API 호출이나 비용(Search Grounding)이 발생할 수 있습니다. (Google AI Studio 정책 참고)

---
수고하셨습니다! 이제 여러분은 **"검색하는 AI 에이전트"**를 만들었습니다.
다음 단계에서는 우리가 직접 만든 커스텀 도구를 에이전트에게 쥐어주는 방법을 알아볼 것입니다. 🚀

---

# Search Agent (도구(Tools)를 활용한 검색 에이전트 만들기)

## English Version

Welcome! 👋 Welcome to the second session.

In the previous session, we built a basic conversational agent. However, that agent had limitations—it couldn't access information after its training cutoff date or real-time news.

In this session, we will use ADK's powerful **Tools** system to give our agent **Google Search** capabilities. Now, your agent will be able to search for and answer questions using the latest information.

### 🎯 Learning Objectives
*   Understand the **ADK Tool Interface**.
*   Integrate Google Search functionality using the `geminitool` package.
*   Implement an agent that answers questions requiring real-time information.

### 💻 Code Explanation

This code is similar in structure to Session 1, but with the addition of **Tools** configuration. We will focus on the changes.

### 1. Add Tool-related Packages
```go
import (
    // ... existing imports omitted ...
	"google.golang.org/adk/tool"
	"google.golang.org/adk/tool/geminitool"
    // ...
)
```
*   `google.golang.org/adk/tool`: This is the core interface for defining and managing tools in ADK.
*   `google.golang.org/adk/tool/geminitool`: This package contains predefined tools that the Gemini model can use, such as Google Search and code execution.

### 2. Model Initialization (Same as Session 1)
```go
	model, err := gemini.NewModel(ctx,
		"gemini-3-pro-preview",
		&genai.ClientConfig{
			APIKey: os.Getenv("GOOGLE_API_KEY"),
		})
    // ...
```
*   The Gemini model is initialized in the same way as before. Google Search functionality works very effectively with Gemini 2.5 and higher models.

### 3. Equip the Agent with Tools (Key Change) ⭐
This is the most crucial part. We add the `Tools` option when creating the agent.

```go
	// Although the variable name is timeAgent, its actual role is a search agent.
	timeAgent, err := llmagent.New(llmagent.Config{
		Name:        "search_agent", // Agent name changed
		Model:       model,
		Description: "A helpful agent that searches the web.", // Description updated
		Instruction: "You are a helpful assistant. Use Google Search to answer the user's questions.", // Instruct to use search
		
        // [Key] Define the list of tools
		Tools: []tool.Tool{
			geminitool.GoogleSearch{}, // Add Google Search tool
		},
	})
```
*   **`Tools: []tool.Tool{...}`**: This is a list of tools that the agent can use.
*   **`geminitool.GoogleSearch{}`**: With just this single line, the agent gains the ability to use the Google Search engine (Grounding with Google Search) without complex custom implementation.
*   **`Instruction`**: By explicitly stating "Use Google Search" in the prompt, the model will better determine when to use the tool.

### 4. Launcher Execution (Same as Session 1)
```go
	config := &launcher.Config{
		AgentLoader: agent.NewSingleLoader(timeAgent),
	}
    // ... same execution logic ...
```

### 🚀 Run and Test (Let's Run!)

Save the code and run it in the terminal. This time, it's important to ask questions that require real-time information.

### 1. Run in Chat Mode
```bash
go run main.go chat
```

### 2. Example Questions (Comparison)

**Q1. (Past Knowledge) "What is the capital of the United States?"**
*   Since the model already knows this, it might answer directly without searching.

**Q2. (Real-time Information) "Tell me the result of Son Heung-min's game yesterday" or "What's the weather like in Seoul today?"**
*   **Session 1 Agent**: Would have replied, "Sorry, I cannot access real-time information."
*   **Session 2 Agent**: Will internally perform a Google Search (Grounding) and generate an answer based on the latest information.

### 🔍 What Happened? (Under the Hood)

1.  The user asks, **"What's the weather like in Seoul today?"**
2.  The agent (LLM) determines that it cannot answer this question with its internal knowledge.
3.  However, it knows that a `geminitool.GoogleSearch` tool is available.
4.  The agent itself generates a search query like **"Seoul weather"** and calls the tool.
5.  Google search results are passed to the agent.
6.  The agent summarizes the search results and delivers a natural answer to the user.

All these processes are handled automatically between ADK and the Gemini model!

### 💡 Notes

*   **Grounding**: When an LLM generates answers based on external data (like search results), it's called **Grounding**. This helps reduce hallucinations (false answers) and increases reliability.
*   **Cost**: Using search tools may incur additional API calls or costs (Search Grounding) beyond general text generation. (Refer to Google AI Studio policies).

---
