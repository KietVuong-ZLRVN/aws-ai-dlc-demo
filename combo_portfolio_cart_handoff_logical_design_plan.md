# Logical Design Plan
## Focus: Unit 4 (Combo Portfolio) & Unit 5 (Cart Handoff)

---

## Clarifying Questions

> Please fill in the `[Answer]` tags before approving the plan. These decisions directly affect the technology choices in the logical design.

---

**[Question]** What is the target cloud platform and runtime? (e.g., AWS with Lambda + API Gateway, AWS with ECS/Fargate, or another stack?)

**[Answer]**
AWS with ECS
---

**[Question]** For Unit 4 — What is the expected persistence technology for the `Combo` aggregate? (e.g., Amazon DynamoDB, Amazon RDS/Aurora PostgreSQL, Amazon DocumentDB?)

**[Answer]**
currently the DB in core service is Amazon Aurora MySQL, however to build it as PoC and check the success rate first, how about string it in MemCached - the only storage that the service interact with
---

**[Question]** For Unit 5 — What is the expected persistence technology for the `CartHandoffRecord` audit log? (Same as Unit 4, or a separate store such as Amazon DynamoDB or Amazon S3 + Athena for analytics?)

**[Answer]**
just use the same, i think can do in Amazon Aurora MySQL
---

**[Question]** For the event bus / domain event publishing — Should domain events (e.g., `ComboCreated`, `ComboShared`, `CartHandoffRecorded`) be published to an external message bus (e.g., Amazon EventBridge, Amazon SNS/SQS) for downstream consumers such as analytics or audit systems, or should events be stored in-process only (no external bus at this stage)?

**[Answer]**
no need, just store serve the basic usage now, can decide the event publishing later on
---

**[Question]** For Unit 4 — The `ComboEnrichmentService` calls the Platform Product Catalog on every read to refresh item data. Should this enrichment be **synchronous** (called inline during the GET request, increasing latency) or **cached** (product data cached per `simpleSku` with a TTL, e.g., via ElastiCache or DynamoDB TTL)? If cached, what TTL is acceptable?

**[Answer]**
if you call by the simpleSku it already a function get the information in the Memcache layer already
---

**[Question]** For authentication and authorisation — Is there an existing API Gateway authoriser or JWT-based auth middleware that both units should integrate with, or should the logical design propose an auth mechanism?

**[Answer]**
it already have the session base authorization implemented in the service
---

## Follow-up Questions (based on your answers above)

> These arose from reviewing your answers. Please fill in the `[Answer]` tags.

---

**[Question]** For Unit 4 persistence — **Memcached is a volatile cache, not a database.** There are three concrete risks I want you to be aware of before we commit to it:
1. **Data loss on restart** — If the Memcached node restarts or is replaced (e.g., ECS task replacement, scaling event), all saved combos are permanently lost with no recovery path.
2. **Silent LRU eviction** — When memory pressure builds, Memcached silently evicts the least-recently-used items. A shopper could lose a saved combo without any error or warning.
3. **No query support** — The `ComboRepository` needs `findByShopperId(shopperId)` and `findByShareToken(token)`. Memcached only supports exact key lookups. To support these queries, we would need to maintain manual index keys (e.g., a separate key `"shopper:{id}:combo_ids"` holding a list of combo IDs), which adds complexity and introduces consistency risks.

Given that Aurora MySQL is already available and being used for Unit 5's audit log, would you like to **use Aurora MySQL for both units** instead? This avoids all three risks at no infrastructure cost. If you still want Memcached for the PoC, please confirm you accept these trade-offs and I will design the manual index pattern accordingly.

**[Answer]**
can use mySQL for both
---

**[Question]** For ECS deployment — What is the application language/framework for the service containers? (e.g., Python with FastAPI/Flask, Node.js with Express, Java with Spring Boot, Go?) This determines the file/module structure in the logical design.

**[Answer]**
the backend service is written in Go using go-chi, frontend in React framework
---

**[Question]** For Unit 4 enrichment — You mentioned that product lookups by `simpleSku` already hit a Memcached layer in the existing service. Should the `ComboEnrichmentService` call this **existing internal product lookup endpoint** directly (i.e., reuse the existing service's Memcached-backed API), or will Unit 4 have its own direct connection to the Memcached cluster to query product data?

**[Answer]**
i want to build the feature on the service have the API GET /api/v1/products/{configSku}, so we can use memcached to read the product data
---

## Follow-up Questions Round 2 (based on your latest answers)

> Three architectural questions arose that will significantly affect the service topology and file structure in the logical design.

---

**[Question]** Your answer says Unit 4 will be built on the same service that owns `GET /api/v1/products/{configSku}`. Does this mean:
- **(A) Unit 4 is a new module inside the existing Product Discovery service (Unit 1)** — same ECS task/container, same Go binary, same Memcached connection, in-process function calls for enrichment, OR
- **(B) Unit 4 is a brand-new standalone ECS service** that calls Unit 1's HTTP API (`GET /api/v1/products/{configSku}`) for enrichment over the network?

This determines whether `ComboEnrichmentService` is an in-process call or an HTTP ACL adapter, and the entire file/module layout.

**[Answer]**
discuss with teammate and choose to separate service to go with this feature, so should use the API http
---

**[Question]** Following the above — what about Unit 5 (Cart Handoff)? Is it:
- **(A) Also built into the same existing service** (same ECS container as Unit 1 / Unit 4), OR
- **(B) A separate standalone ECS service?**

**[Answer]**
B
---

**[Question]** For Aurora MySQL — Is there an **existing Aurora cluster** that the new tables should be added to (alongside existing tables), or should the logical design call for a **new dedicated Aurora cluster** for these units?

**[Answer]**
create new mySQL db for the new service
---

## Plan Steps

- [ ] **Step 1** — Re-read both domain models (`combo_portfolio/domain_model.md`, `cart_handoff/domain_model.md`) and the integration contract to confirm all inputs
- [ ] **Step 2** — Design logical architecture for **Unit 4: Combo Portfolio**
  - [ ] Define the overall service architecture (layers: API, Application, Domain, Infrastructure)
  - [ ] Define the API layer: HTTP routes, request/response contracts, auth middleware
  - [ ] Define the Application layer: Application Services (use-case handlers / command handlers)
  - [ ] Define the Domain layer: class structure for Aggregate, Value Objects, Domain Events
  - [ ] Define the Infrastructure layer: Repository implementation, Product Catalog ACL adapter, Event Publisher
  - [ ] Define the data schema for the `Combo` persistence store
  - [ ] Define the event publishing design (domain events → external bus or internal only)
  - [ ] Define the read model / projection design for `EnrichedCombo` (enrichment strategy)
  - [ ] Define the proposed file/module structure
- [ ] **Step 3** — Design logical architecture for **Unit 5: Cart Handoff**
  - [ ] Define the overall service architecture (layers: API, Application, Domain, Infrastructure)
  - [ ] Define the API layer: HTTP routes, request/response contracts, auth middleware
  - [ ] Define the Application layer: Application Services (use-case handlers)
  - [ ] Define the Domain layer: class structure for `CartHandoffRecord`, Value Objects, Domain Events
  - [ ] Define the Infrastructure layer: Repository implementation, `ComboPortfolioACL` adapter, `PlatformCartACL` adapter, Event Publisher
  - [ ] Define the data schema for the `CartHandoffRecord` audit store
  - [ ] Define the event publishing design
  - [ ] Define the proposed file/module structure
- [ ] **Step 4** — Write `/construction/combo_portfolio/logical_design.md`
- [ ] **Step 5** — Write `/construction/cart_handoff/logical_design.md`

---

> **Awaiting your review and approval before execution.**
> Please fill in the [Answer] tags above and confirm the plan is correct. Once approved, I will execute one step at a time and mark each checkbox as done upon completion.
