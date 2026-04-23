---
name: java-ee-to-quarkus
description: >
  Migration guide for Java EE to Quarkus. Use when migrating a JBoss EAP /
  Java EE application to Quarkus 3.x. Covers EJB → CDI, JMS → SmallRye
  Reactive Messaging, JAX-RS → RESTEasy Reactive, JPA patterns.
---

# Java EE to Quarkus Migration Guide

## What this migration means
Migrating from JBoss EAP 7.x / Java EE 8 to Quarkus 3.x. Key changes:
- @Stateless → @ApplicationScoped
- @Stateful → @ApplicationScoped with explicit state management  
- @Remote EJBs → REST endpoints
- JMS/JMSContext → SmallRye Reactive Messaging
- JAX-RS → RESTEasy Reactive (keep annotations, change imports)
- @PersistenceContext EntityManager → @Inject EntityManager
- JNDI lookups → CDI injection or application.properties

## Key patterns to apply
1. Replace javax.* imports with jakarta.* equivalents
2. Convert @Stateless/@Stateful service beans to @ApplicationScoped
3. Add @Transactional where @Stateless provided it implicitly
4. Replace JMS producers/consumers with @Incoming/@Outgoing channels
5. Update persistence.xml to Quarkus datasource config in application.properties

## Organizational defaults (override if user specifies otherwise)
- Messaging: SmallRye Reactive Messaging with in-memory connector (unless RabbitMQ/Kafka specified)
- REST: RESTEasy Reactive
- ORM: Hibernate ORM with Panache where applicable
- Preserve all test files unless they directly import EE APIs

## Known compatibility requirements
- Use **Quarkus 3.8.6+** — earlier versions have Byte Buddy issues with Java 21+
- Set `maven.compiler.source/target/release` to **17** (even if running Java 21+)
- `io.smallrye.reactive:smallrye-reactive-messaging-in-memory` must be added explicitly (not in Quarkus BOM)
- Local/system-scoped JAR dependencies (e.g. `<scope>system</scope>`) need manual resolution — flag these to the user
- Add `net.bytebuddy.experimental=true` JVM arg if Hibernate enhancement fails on newer JVMs

## How to proceed
1. Run `find . -name "*.java" | sort` and `cat pom.xml` to understand project structure
2. Update `pom.xml` first — replace Java EE deps with Quarkus BOM, set compiler to Java 17
3. Migrate in this order: models → services → REST endpoints → utils → remove deleted files
4. Run `mvn compile -q` after each group — fix errors before moving on
5. Commit progress per group with message "chore: migrate <layer> to quarkus"
6. Final: run `mvn package -DskipTests -q` to verify full build succeeds
