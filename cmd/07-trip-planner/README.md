# Go ADK 핸즈온 Session 7: 멀티 에이전트 워크플로우 (Parallel & Sequential)

환영합니다! 👋 대망의 일곱 번째 세션입니다.

지금까지는 혼자 일하는 에이전트를 만들었습니다. 하지만 현실의 복잡한 업무는 혼자서 처리하기 어렵습니다. 마치 여행을 갈 때 친구 한 명은 맛집을 찾고, 다른 친구는 관광지를 찾은 뒤, 계획 짜는 친구가 이를 취합하여 일정을 만드는 것처럼 말이죠.

이번 시간에는 **Parallel Agent(병렬)**와 **Sequential Agent(직렬)**를 조합하여, **"여행 계획 어벤져스"** 팀을 만들어 보겠습니다.

## 🎯 학습 목표
*   **Multi-Agent Architecture**: 여러 에이전트를 레고 블록처럼 조립하는 방법 이해하기
*   **Parallel Agent**: 두 개 이상의 에이전트를 동시에 실행시켜 속도를 높이는 방법 (`CityScouts`)
*   **Sequential Agent**: 앞선 에이전트의 결과를 받아 다음 작업을 수행하는 파이프라인 구축 (`TripPlannerPipeline`)
*   **OutputKey & Context Sharing**: 에이전트끼리 데이터를 주고받는 핵심 메커니즘 (`{variable}` 문법)

---

## 💻 코드 상세 분석

이번 코드는 **"정찰조(Scouts)"**가 먼저 정보를 수집하고, **"계획가(Planner)"**가 이를 정리하는 구조입니다.

### 1. 정찰조 에이전트 정의 (The Specialists)
먼저 특정 정보만 전문적으로 수집하는 두 명의 에이전트를 만듭니다.

```go
	restaurantScout, _ := llmagent.New(llmagent.Config{
		Name:  "RestaurantScout",
		// ... 모델 설정 ...
		Instruction: `... Extract the city name ... find the top 3 restaurants ...`,
		Tools:     []tool.Tool{geminitool.GoogleSearch{}},
		// [핵심] 이 에이전트가 찾은 결과는 'restaurant_list'라는 변수에 저장됩니다.
		OutputKey: "restaurant_list",
	})

	activityScout, _ := llmagent.New(llmagent.Config{
		// ... 설정 ...
		// [핵심] 이 에이전트가 찾은 결과는 'activity_list'라는 변수에 저장됩니다.
		OutputKey: "activity_list",
	})
```
*   **`OutputKey`**: 가장 중요한 설정입니다. 이 에이전트가 수행한 결과(검색된 맛집 목록 등)를 공유 메모리(Context)의 **어떤 변수명**으로 저장할지 지정합니다.

### 2. 병렬 실행 그룹 (Parallel Agent) ⚡
두 정찰조는 서로의 결과가 필요 없습니다. 따라서 동시에 실행하는 것이 효율적입니다.

```go
	scouts, _ := parallelagent.New(parallelagent.Config{
		AgentConfig: agent.Config{
			Name:        "CityScouts",
			// 하위 에이전트 목록
			SubAgents:   []agent.Agent{restaurantScout, activityScout},
		},
	})
```
*   **`parallelagent`**: 등록된 `SubAgents`를 동시에 실행합니다. 맛집 검색과 관광지 검색이 동시에 일어나므로 전체 응답 시간이 단축됩니다.

### 3. 계획가 에이전트 (The Consumer)
앞선 정찰조가 모아온 정보를 바탕으로 최종 일정을 짭니다.

```go
	planner, _ := llmagent.New(llmagent.Config{
		Name:  "ItineraryPlanner",
		// ...
		// [핵심] 앞선 에이전트가 OutputKey로 저장한 변수를 {변수명} 형태로 가져다 씁니다.
		Instruction: `...
        Restaurants: {restaurant_list}
        Activities: {activity_list}
        Combine them into a logical schedule.`,
	})
```
*   **`{placeholder}`**: 프롬프트 내에 중괄호를 사용하면, ADK는 공유 메모리에서 해당 키(`restaurant_list`, `activity_list`)에 담긴 값을 찾아 자동으로 채워 넣습니다. 이것이 에이전트 간 데이터 전달 방식입니다.

### 4. 직렬 파이프라인 (Sequential Agent) 🔗
마지막으로 "정찰조 실행(병렬) -> 계획가 실행" 순서로 흐름을 연결합니다.

```go
	tripPlanner, _ := sequentialagent.New(sequentialagent.Config{
		AgentConfig: agent.Config{
			Name:        "TripPlannerPipeline",
			// [순서 중요] scouts가 먼저 실행되어야 planner가 데이터를 쓸 수 있습니다.
			SubAgents:   []agent.Agent{scouts, planner},
		},
	})
```
*   **`sequentialagent`**: 리스트에 적힌 순서대로 에이전트를 실행합니다. 앞 에이전트가 작업을 마쳐야 다음 에이전트가 시작됩니다.

---

## 🚀 실행 및 테스트 (Let's Travel!)

이 워크플로우는 입력값(도시 이름) 하나만 주면 알아서 검색하고 계획까지 짜줍니다.

### 1. 실행 명령어
"도쿄 여행 계획 짜줘"라고 명령해 봅시다. (영어 프롬프트가 더 정확할 수 있습니다)

```bash
go run main.go run "Plan a trip to Tokyo"
```

### 2. 내부 동작 흐름 (Visualized)

```text
Input: "Plan a trip to Tokyo"
      │
      ▼
[ Sequential Agent: TripPlannerPipeline ]
      │
      ├── Step 1: [ Parallel Agent: CityScouts ] ⚡
      │     ├───> [RestaurantScout] : "Tokyo Restaurants" 검색
      │     │        └─> 결과 저장: OutputKey="restaurant_list"
      │     │
      │     └───> [ActivityScout]   : "Tokyo Activities" 검색
      │              └─> 결과 저장: OutputKey="activity_list"
      │
      │     (두 에이전트가 모두 끝날 때까지 대기)
      │
      └── Step 2: [ Agent: ItineraryPlanner ] 📝
            │  프롬프트 완성: 
            │  "Restaurants: [스시집, 라멘집...]"
            │  "Activities: [도쿄타워, 시부야...]"
            │
            └─> 최종 결과: "오전엔 도쿄타워 갔다가 점심엔 스시를 드세요..."
```

### 3. 결과 확인
콘솔에 최종적으로 정리된 **하루 여행 일정표**가 출력되는지 확인하세요.

---

## 🔍 핵심 포인트 (Key Takeaways)

1.  **분업의 힘**: 하나의 거대한 프롬프트로 모든 걸 처리하려는 것보다, 역할을 쪼개고(검색 담당, 요약 담당) 전문화된 에이전트를 연결하는 것이 훨씬 성능이 좋고 관리가 쉽습니다.
2.  **속도 최적화**: 서로 의존성이 없는 작업(맛집 찾기 vs 관광지 찾기)은 `ParallelAgent`로 묶어서 실행 시간을 획기적으로 줄일 수 있습니다.
3.  **데이터 파이프라인**: `OutputKey`와 프롬프트 내 `{Key}` 문법을 통해, 코드를 수정하지 않고도 에이전트끼리 자연스럽게 데이터를 주고받을 수 있습니다.

---
수고하셨습니다! ✈️ 이제 여러분은 단순한 챗봇 개발자를 넘어, 복잡한 **AI 에이전트 시스템을 설계하는 아키텍트**가 되셨습니다. 이 구조를 활용하면 뉴스 요약, 법률 문서 분석, 자동화된 리포트 생성 등 무궁무진한 응용 프로그램을 만들 수 있습니다.