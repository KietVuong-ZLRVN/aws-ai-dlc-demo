# Unit 3: AI Styling Engine — DDD Domain Model Plan

## Objective

Design the DDD domain model (aggregates, entities, value objects, domain events, policies, repositories, domain services) for Unit 3: AI Styling Engine. Output will be written to `/construction/unit_3_ai_styling_engine/domain_model.md`.

---

## Steps

- [x] **Step 1**: Analyse Unit 3 user stories and API contracts to identify domain concepts, behaviours, and invariants
- [x] **Step 2**: Define the bounded context and its responsibilities
- [x] **Step 3**: Identify aggregates and aggregate roots
- [x] **Step 4**: Identify entities within each aggregate
- [x] **Step 5**: Identify value objects
- [x] **Step 6**: Identify domain events
- [x] **Step 7**: Identify policies (business rules that react to domain events or conditions)
- [x] **Step 8**: Identify domain services (operations that don't naturally belong to any single aggregate)
- [x] **Step 9**: Identify repositories and anti-corruption layers for external dependencies (Unit 2, Platform APIs)
- [x] **Step 10**: Create `/construction/unit_3_ai_styling_engine/` directory and write `domain_model.md`

---

## Clarification Questions

**[Question 1]** US-301b states the shopper can "save their preferences for reuse", but the Unit 3 spec also says it has "no persistent storage of its own". Should saved style preferences be stored by this unit (requiring some form of persistence), or should preference persistence be delegated to a different unit or external service?

**[Answer 1]**
The preference to be saved by the users is actually the result gotten from this unit, so there will be no storing for this unit at all. Just providing suggestion for the user to choose and store on other unit corespondingly
---

**[Question 2]** Domain events in a stateless, BFF-style service can serve different purposes: (a) purely conceptual/documentation only, (b) in-process event handling for decoupling application logic, or (c) published to an event bus for cross-service consumption. Which intent applies here — and does this affect whether we model domain events at all for Unit 3?

**[Answer 2]**
it will be (b)
---

**[Question 3]** When the AI scores item compatibility and generates combo reasoning, should this be modelled as a single `StyleInferenceService` that does everything, or should scoring (compatibility ranking) and explanation generation (natural language reasoning) be separated into distinct domain services?

**[Answer 3]**
scoring is more technical in terms of fashion senses and all. So one for the scoring and one for the explanation for human readbability
---

> **Please review and answer the questions above (mark [Answer] sections), then approve to proceed with execution.**

---

## Phase 2 — Logical Design

### Objective

Generate a logical design for software source code implementation of Unit 3, based on `/construction/unit_3_ai_styling_engine/domain_model.md` and `/inception/units/integration_contract.md`. Output will be written to `/construction/unit_3_ai_styling_engine/logical_design.md`.

### Steps

- [x] **Step 11**: Review domain model and integration contract to confirm all components to be designed
- [x] **Step 12**: Define the layered architecture and how DDD layers map to source folders
- [x] **Step 13**: Design the API layer — controllers, request/response models, and routing
- [x] **Step 14**: Design the application layer — use cases, command/query objects, and orchestration flow
- [x] **Step 15**: Design the domain layer — aggregate lifecycle, entity methods, event raising, and policy wiring
- [x] **Step 16**: Design the infrastructure layer — ACL implementations for all three external dependencies, in-process event dispatcher
- [x] **Step 17**: Design the AI integration layer — interfaces for scoring service and reasoning service
- [x] **Step 18**: Write `/construction/unit_3_ai_styling_engine/logical_design.md`

---

### Clarification Questions

**[Question 4]** The logical design will reference the AI/ML integration for `ComboCompatibilityScoringService` and the LLM integration for `ComboReasoningGenerationService` and `PreferenceInterpretationService`. Should the design target a specific AI platform (e.g., AWS Bedrock, OpenAI, a custom model endpoint), or should it define a technology-agnostic interface that an implementation can plug into?

**[Answer 4]**
Can target a specific platform to avoid abstraction confusion later on
---

**[Question 5]** The logical design could include or exclude cross-cutting concerns such as authentication middleware, request logging, distributed tracing, and error handling patterns. Should these be included in the logical design, or scoped out (focused purely on functional components)?

**[Answer 5]**
should be incluđe to make sure cases are covered
---

**[Question 6]** The `ComboCompatibilityScoringService` may need to call the Platform Complete-the-Look API as a base signal before running its own scoring. Should these two steps (fetch platform signals, then run AI scoring) be designed as a single sequential operation inside the service, or as two explicitly separate steps with their own interfaces?

**[Answer 6]**
Should be tưo explicitly separate steps to make clear of the functional scope and aid in testability
---

> **Please review and answer Questions 4–6 above, then approve to proceed with Steps 11–18.**
