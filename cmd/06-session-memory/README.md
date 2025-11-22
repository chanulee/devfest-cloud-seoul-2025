# Go ADK 핸즈온 Session 6: 기억을 가진 에이전트 (Memory & Runner)

환영합니다! 👋 여섯 번째 세션입니다.

지금까지 만든 에이전트들은 대화가 끝나면 모든 것을 잊어버렸습니다(Stateless). 하지만 진정한 AI 비서는 사용자의 이름, 취향, 과거 대화 내용을 기억해야 합니다.

이번 시간에는 **ADK의 Runner**를 사용하여 대화 흐름을 직접 제어하고, **Memory Service**를 통해 에이전트에게 **"장기 기억(Long-term Memory)"** 능력을 부여해 보겠습니다.

## 🎯 학습 목표
*   **Runner vs Launcher**: 더 세밀한 제어를 위해 `Launcher` 대신 `Runner` 사용하기
*   **Memory Service**: 대화 내용을 저장(`session`)하고, 저장된 내용을 검색(`memory`)하는 구조 이해하기
*   **Memory Tool**: 에이전트가 자신의 기억 저장소를 검색하는 도구(`search_past_conversations`) 구현
*   **Prompt Engineering**: 한국어 기억 검색 정확도를 높이기 위한 프롬프트 기법

---

## 💻 코드 상세 분석

코드가 조금 길어졌습니다. 핵심은 **"대화 저장 -> 인덱싱 -> 검색"**의 순환 구조입니다.

### 1. 기억 검색 도구 (`memorySearchToolFunc`) ⭐
에이전트가 과거의 기억을 뒤져볼 수 있게 해주는 도구입니다.

```go
func memorySearchToolFunc(tctx tool.Context, args Args) (Result, error) {
    // ...
    // ADK가 제공하는 내장 메모리 검색 기능 호출
    searchResults, err := tctx.SearchMemory(context.Background(), args.Query)
    // ...
}
```
*   **`tctx.SearchMemory`**: ADK 프레임워크는 `tool.Context`를 통해 현재 연결된 Memory Service에 바로 접근할 수 있는 기능을 제공합니다.
*   이 함수는 에이전트가 "사용자가 내 이름을 뭐라고 했지?"라고 생각할 때 호출됩니다.

### 2. 서비스 초기화 (Service Initialization)
```go
	// 세션(현재 대화 상태)과 메모리(저장된 기억) 서비스 생성
	sessionService := session.InMemoryService()
	memoryService := memory.InMemoryService()
```
*   **`sessionService`**: 현재 진행 중인 대화의 문맥(Context)을 관리합니다.
*   **`memoryService`**: 완료된 대화를 저장하고, 나중에 검색할 수 있도록 보관하는 저장소입니다.
*   *참고: 실무에서는 `InMemory` 대신 Redis나 데이터베이스 기반의 서비스를 사용합니다.*

### 3. 에이전트 프롬프트 (Language Optimization)
한국어로 기억을 검색할 때 발생하는 문제를 해결하기 위한 프롬프트 엔지니어링입니다.

```go
	Instruction: `
    ...
    RULES FOR MEMORY:
    1. Always use 'search_past_conversations' tool when ...
    2. CRITICAL: If the conversation is in Korean, generate the search query IN KOREAN.
       - Bad Query: "user name"
       - Good Query: "내 이름", "사용자 이름"
    ...`
```
*   **문제점**: LLM은 기본적으로 영어를 선호하여, 한국어 대화 중에도 검색 쿼리를 "user's name"으로 날릴 수 있습니다. 하지만 저장된 기억은 "내 이름은 철수야"라는 한글 텍스트이므로 검색에 실패할 수 있습니다.
*   **해결책**: `Instruction`에 "한국어 대화면 한국어 키워드로 검색해라"라고 명시하여 검색 적중률을 높였습니다.

### 4. 런처(Launcher)에서 러너(Runner)로의 전환 🔄
이전 세션까지는 `l.Execute()` 한 줄로 끝났지만, 이제는 `Runner`를 통해 대화 루프를 직접 만듭니다.

```go
	// Runner 생성: Agent와 Memory 시스템을 묶어주는 실행기
	r, err := runner.New(runner.Config{
		AppName:        "MemoryApp",
		Agent:          rootAgent,
		SessionService: sessionService,
		MemoryService:  memoryService, // Runner에 메모리 서비스를 연결해야 Tool에서 접근 가능
	})
```

### 5. 수동 기억 저장 (The Learning Step)
가장 중요한 부분입니다. 에이전트가 답을 했다고 저절로 기억이 저장되지 않습니다. 우리가 직접 저장해 주어야 합니다.

```go
	for {
        // ... (사용자 입력 및 에이전트 실행) ...

        // [중요] 대화 턴이 끝난 후, 최신 세션 상태를 가져와서 메모리에 '영구 저장'
		latestSession, _ := sessionService.Get(...)
		if err := memoryService.AddSession(ctx, latestSession.Session); err != nil {
            // ...
		}
	}
```
*   **`memoryService.AddSession`**: 방금 나눈 대화를 검색 가능한 메모리 저장소에 인덱싱합니다. 이 코드가 없으면 에이전트는 방금 한 말도 기억하지 못합니다(검색 불가).

---

## 🚀 실행 및 테스트 (Scenario Test)

이 코드는 대화의 맥락을 유지하는지 확인하는 것이 핵심입니다.

### 1. 실행
```bash
go run main.go
```

### 2. 테스트 시나리오

**Step 1: 정보 주입 (기억 심기)**
```text
User: 안녕, 내 이름은 '홍길동'이고, 나는 'Go 언어'를 좋아해. 기억해줘.
Bot: 네, 안녕하세요 홍길동! Go 언어를 좋아하시는군요. 잘 기억해두겠습니다.
```
*(이 시점에서 `[시스템] 기억 저장 완료` 메시지가 떠야 합니다.)*

**Step 2: 문맥 변경 (딴청 피우기)**
```text
User: 오늘 점심 뭐 먹을까?
Bot: (점심 메뉴 추천...)
```
*(기억이 희석되는 과정을 시뮬레이션합니다.)*

**Step 3: 기억 인출 (검색 도구 작동 확인)**
```text
User: 내가 아까 내 이름이 뭐라고 했지? 그리고 내가 좋아하는 게 뭐야?
```

**예상되는 내부 동작 로그:**
```text
[Tool] 검색어: '내 이름 좋아하는 것'
 -> 1개 찾음
```

**Bot의 답변:**
```text
Bot: 아까 사용자님의 이름은 '코드깎는노인'이고, 'Go 언어'를 좋아하신다고 하셨습니다!
```

---

## 🔍 심화 개념 (Under the Hood)

이 예제는 **RAG (Retrieval-Augmented Generation)** 의 아주 기본적인 형태입니다.

1.  **Save**: 대화 내용(텍스트)을 벡터나 키워드 형태로 데이터베이스에 저장합니다.
2.  **Retrieve**: 사용자가 질문하면, 에이전트는 `search_past_conversations` 도구를 통해 관련성 높은 과거 대화 조각을 찾아냅니다.
3.  **Generate**: 찾아낸 기억 조각을 문맥(Context)에 포함시켜 답변을 생성합니다.

---

## 💡 팁 (Troubleshooting)

*   **기억을 못 해요!**: `memoryService.AddSession` 부분이 `Run` 루프 안에 있는지 확인하세요. 대화가 끝나야 기억이 저장됩니다.
*   **검색 결과가 없대요**: `Instruction`에 있는 "한국어로 검색해라" 부분이 잘 동작하는지 로그(`[Tool] 검색어: ...`)를 확인해 보세요. 검색어가 영어라면 프롬프트를 더 강하게 수정해야 합니다.

---
수고하셨습니다! 🎉 이제 여러분의 에이전트는 단순한 앵무새가 아니라, **사용자와의 추억을 간직하는 지능형 비서**로 진화했습니다. 이것으로 ADK 핸즈온의 핵심 기능을 모두 마스터하셨습니다!