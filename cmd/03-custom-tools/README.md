# Go ADK 핸즈온 Session 3: 나만의 커스텀 도구(Custom Tools) 만들기

환영합니다! 👋 세 번째 세션에 오신 것을 환영합니다.

이전 세션에서는 `Google Search`라는 이미 만들어진 도구를 사용했습니다. 하지만 실제 애플리케이션을 개발할 때는 **내 데이터베이스를 조회**하거나, **내부 API를 호출**하거나, **특정 비즈니스 로직을 실행**해야 할 때가 많습니다.

이번 시간에는 **여러분이 작성한 Go 함수**를 에이전트가 스스로 호출하여 작업을 수행하도록 만드는 **Function Calling** 기능을 구현해 보겠습니다.

## 🎯 학습 목표
*   **Custom Tool**의 구조 이해하기 (입력 구조체와 핸들러 함수)
*   `functiontool` 패키지를 사용하여 Go 함수를 에이전트 도구로 변환하기
*   `jsonschema` 태그를 통해 LLM에게 인자 정보 전달하기
*   에이전트가 도구를 연속적으로 사용하는(Chaining) 과정 관찰하기

---

## 💻 코드 상세 분석

이번 코드는 `main.go`와 도구 구현부(날씨, 감정분석)로 나뉘어 있습니다.

### 1. 도구 정의하기 (The "Logic")
에이전트가 사용할 실제 함수를 정의하는 부분입니다. ADK는 Go의 구조체 태그를 분석하여 LLM에게 함수 사용법을 알려줍니다.

**A. 날씨 조회 도구 (`getWeather`)**
```go
// 입력 인자를 정의하는 구조체
type getWeatherArgs struct {
    // jsonschema 태그: LLM이 이 필드가 무엇인지 이해하는 설명서 역할을 합니다.
	City string `json:"city" jsonschema:"The city to get weather for."`
}

// 실제 실행될 함수
func getWeather(ctx tool.Context, args getWeatherArgs) (string, error) {
	fmt.Printf("[Tool] Getting weather for %s...\n", args.City) // 실행 확인용 로그
	return fmt.Sprintf("The weather in %s is Sunny, 25°C", args.City), nil // Mock 데이터 반환
}
```
*   **`struct`와 `jsonschema`**: LLM은 이 태그를 보고 "아, `city`라는 인자에 도시 이름을 넣어서 호출해야 하는구나"라고 판단합니다.
*   **함수 시그니처**: `func(ctx tool.Context, args T) (string, error)` 형태를 따라야 합니다.

**B. 감정 분석 도구 (`analyzeSentiment`)**
```go
type analyzeSentimentArgs struct {
	Text string `json:"text" jsonschema:"The text to analyze."`
}

func analyzeSentiment(ctx tool.Context, args analyzeSentimentArgs) (string, error) {
	fmt.Printf("[Tool] Analyzing sentiment for: %s\n", args.Text)
	// 실제로는 외부 API를 부르거나 복잡한 로직이 들어갈 자리입니다.
	return "Positive Sentiment", nil
}
```

### 2. 도구 등록하기 (`main.go`)
작성한 함수를 에이전트가 사용할 수 있는 `Tool` 객체로 포장합니다.

```go
	// 1. 날씨 도구 생성
	weatherTool, _ := functiontool.New(functiontool.Config{
		Name: "get_weather", 
        Description: "Get weather for a city" // LLM이 언제 이 도구를 쓸지 판단하는 기준
    }, getWeather)

	// 2. 감정 분석 도구 생성
	sentimentTool, _ := functiontool.New(
		functiontool.Config{
            Name: "analyze_sentiment", 
            Description: "Analyze text sentiment"
        }, analyzeSentiment)
```
*   **`functiontool.New`**: 일반 Go 함수(`getWeather`)를 ADK 호환 도구로 변환해 줍니다.
*   **`Description`**: 매우 중요합니다. 에이전트는 이 설명을 읽고 사용자의 질문과 매칭하여 도구 사용 여부를 결정합니다.

### 3. 에이전트에 도구 쥐어주기
```go
	myAgent, err := llmagent.New(llmagent.Config{
		Name:  "helper_agent",
		Model: model,
		// 지시사항: 날씨를 확인하고, 그에 대한 반응을 분석하라는 복합적인 지시
		Instruction: "You are a helper. If asked about weather, use get_weather. " +
			"Then analyze the user's reaction using analyze_sentiment.", 
		Tools: []tool.Tool{weatherTool, sentimentTool}, // 두 개의 도구 등록
	})
```
*   **`Tools`**: 배열 형태로 여러 개의 도구를 동시에 등록할 수 있습니다. 에이전트는 상황에 따라 하나를 쓰거나, 두 개를 연달아 쓸 수도 있습니다.

---

## 🚀 실행 및 테스트 (Let's Run!)

코드가 여러 파일에 나뉘어 있거나 하나의 패키지 안에 있다면 아래 명령어로 실행하세요.

### 1. 실행 명령어
```bash
# 현재 폴더의 모든 go 파일을 빌드하여 실행
go run main.go
```

### 2. 시나리오 테스트

에이전트에게 다음과 같이 질문해 보세요.

**User:** "서울 날씨 어때?" (How is the weather in Seoul?)

**예상되는 내부 동작 흐름:**
1.  **Agent**: 사용자의 질문("서울 날씨")을 분석 -> `get_weather` 도구가 필요하다고 판단.
2.  **Agent -> Code**: `getWeather(City="Seoul")` 호출.
3.  **Code**:
    *   콘솔 출력: `[Tool] Getting weather for Seoul...`
    *   반환: `"The weather in Seoul is Sunny, 25°C"`
4.  **Agent**: 날씨 정보를 바탕으로 사용자에게 답변 생성.
    *   Answer: "서울 날씨는 맑고 25도입니다."

**복합 시나리오 (Chaining):**
프롬프트에 "반응을 분석하라"는 내용이 있으므로, 대화가 이어질 때 감정 분석 도구가 호출되는지 확인해 보세요.

**User:** "와, 날씨 정말 좋네! 기분 최고야."
1.  **Agent**: 사용자의 텍스트("기분 최고야")를 분석 -> `analyze_sentiment` 도구 호출.
2.  **Code**:
    *   콘솔 출력: `[Tool] Analyzing sentiment for: 와, 날씨 정말 좋네! 기분 최고야.`
    *   반환: `"Positive Sentiment"`
3.  **Agent**: "긍정적인 기분이시군요! 즐거운 하루 되세요."

---

## 🔍 핵심 포인트 (Key Takeaways)

1.  **Type-Safe Function Calling**: Go의 강력한 타입 시스템(`struct`)을 그대로 사용하여 LLM의 입출력을 정의합니다. 복잡한 JSON 파싱을 직접 할 필요가 없습니다.
2.  **Description is Key**: 함수 이름과 `Description`이 명확해야 에이전트가 도구를 올바르게 선택합니다.
3.  **Mocking**: 예제에서는 단순한 문자열을 리턴했지만, 실제로는 `db.Query`나 `http.Get` 등을 사용하여 무한한 기능을 확장할 수 있습니다.

---
수고하셨습니다! 🎉 이제 여러분은 LLM이 실행할 수 있는 **커스텀 함수**를 만들고 연동하는 방법까지 익혔습니다. 이것이 바로 "Agentic Workflow"의 기초입니다.