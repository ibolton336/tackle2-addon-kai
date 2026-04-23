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

## How to proceed
1. Run `find . -name "*.java" | head -20` to understand project structure
2. Start with the most-referenced service classes
3. Fix one file at a time, verify it compiles before moving on
4. Commit progress every 5-10 files with message "chore: migrate <package> to quarkus"
