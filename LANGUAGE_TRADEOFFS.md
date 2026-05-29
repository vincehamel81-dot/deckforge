# Language Tradeoff: Go vs Java for DeckForge

## Context

GoTo gives a binary choice: **Java** or **Go**. Both are used in production at GoTo. The assignment says
"pretend this code will become a foundational part of a new product" and explicitly states you will be
asked about design decisions in the live interview. Language choice is therefore a decision you must
defend, not just execute.

**Decision: Go. See ADR-001 in DECISIONS.md for the final reasoning.**

This document preserves the full analysis for interview reference.

---

## Java (Spring Boot)

### Strengths

| Dimension | Java / Spring Boot |
|---|---|
| .NET familiarity | Near 1-to-1 mapping. `@RestController` ≈ `[ApiController]`. `@Service` ≈ scoped service. `@Autowired` / constructor injection ≈ ASP.NET Core DI. Spring Security ≈ ASP.NET Core auth middleware. `JUnit + Mockito` ≈ xUnit + Moq. `Spring Data JPA` ≈ EF Core. |
| REST API maturity | Spring MVC / Spring Boot is the dominant enterprise Java REST stack. Request validation, exception handling, OpenAPI, and pagination are all first-class. |
| "Foundational product" argument | Strong: Spring Boot, Spring Security, Spring Data are long-lived, battle-tested, widely understood by future maintainers. |
| Clean Architecture | Natural fit: `@Service` (application), `@Repository` (infrastructure), `@RestController` (presentation). Domain layer is plain POJOs. |
| Testability | JUnit 5 + Mockito + `@WebMvcTest` / `@DataJpaTest` is a well-understood pattern. Easy to write layered tests. |
| Observability | Spring Boot Actuator + Micrometer ≈ .NET health checks + App Insights. Structured logging with Logback / SLF4J ≈ Serilog. |
| Error handling | Exception-based (familiar from C#). `@ExceptionHandler` / `@ControllerAdvice` ≈ ASP.NET Core exception filters. |
| IDE support | IntelliJ IDEA is arguably the best Java IDE in existence. Spring initializr generates scaffolding in 30 seconds. |

### Weaknesses

| Dimension | Java |
|---|---|
| Verbosity | More ceremony than Go. Builder patterns, getters/setters (Lombok helps), checked exceptions. |
| Startup / memory | JVM cold start is slow (mitigated with Spring Boot 3 Native, but adds build complexity). |
| Boilerplate risk | Without Lombok or records, entity/DTO classes can get long — leave less time for architecture. |

---

## Go

### Strengths

| Dimension | Go |
|---|---|
| Performance | Compiled binary, no JVM. Very fast startup — relevant for microservices and containers. |
| Simplicity | Language spec is small. No inheritance, no generics complexity (1.18+ has basic generics). Opinionated formatting (`gofmt`). |
| Concurrency | Goroutines + channels — excellent for VoIP/real-time use cases (which is GoTo's domain). |
| Binary size | Single static binary — Docker image is tiny (Alpine + binary, ~10 MB). |
| GoTo alignment | GoTo builds communication software where Go is widely used (Twilio, Zoom, Discord all use Go for real-time services). Choosing Go signals you know the ecosystem. |
| Standard library | `net/http` is production-grade out of the box. Gin or Echo add routing with minimal overhead. |

### Weaknesses

| Dimension | Go |
|---|---|
| Learning curve from .NET | No exceptions — explicit `if err != nil` everywhere. No classes — structs + interfaces. No DI container — manual wiring via constructors. |
| ORM | GORM is usable but has sharp edges. Raw `database/sql` is verbose. Neither maps cleanly to EF Core. |
| Testing | `testing` package is fine but test patterns (table-driven tests, mocking via interfaces) differ from xUnit/Moq. |

---

## Why Go won for this project

1. **GoTo's domain.** VoIP and real-time communication is where Go dominates. Submitting a Go
   project to the WebVoice team is on-brand.
2. **Goroutines for WebSocket.** Each connected player = one goroutine (~2 KB). Broadcasting
   game events to a table is a channel fan-out. This is Go's natural strength.
3. **Foundational product.** A single 8 MB binary in an Alpine container is a better foundation
   than a 200 MB JVM image for a product that will run at scale.
4. **Domain simplicity.** The card game logic (shuffle, deal, score) is pure functions — no ORM
   complexity, no Spring magic needed. Go's simplicity shines here.
5. **Interview signal.** "I chose Go because it aligns with GoTo's real-time infrastructure
   and because goroutines make the WebSocket layer trivially correct" is a stronger answer
   than "I used what I know."

---

## Interview answer (prepared)

> "I chose Go over Java for three reasons. First, GoTo's core product is real-time communication,
> and Go is the dominant language in that space — Twilio, Discord, and Cloudflare all use it for
> their real-time services. Second, Go's goroutine model makes WebSocket handling structurally
> correct and cheap: each player connection is a goroutine, broadcasting to a game is a channel
> fan-out — no thread pool tuning, no executor configuration. Third, the assignment explicitly
> calls this a foundational product. A single statically compiled binary in a 10 MB Alpine image
> is a better foundation than a JVM app. Java was a serious contender — its patterns map
> directly to my .NET background — but Go was the right fit for this specific domain and team."

---

*Decision made: 2026-05-29. Go chosen.*
