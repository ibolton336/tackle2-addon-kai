---
name: java-ee-to-quarkus
description: >
  Transformation reference for Java EE to Quarkus 3.x migration.
  Consumed by the orchestrator skill — defines exact mappings.
---

# Java EE → Quarkus 3.x — Transformation Reference

## Identity

- **Source:** Java EE 7/8 (javax.* namespace), JBoss EAP 7.x
- **Target:** Quarkus 3.8+ (jakarta.* namespace)
- **Packaging:** WAR/EAR → JAR

## Build Transformation (pom.xml)

### Remove

```xml
<dependency><groupId>javax</groupId><artifactId>javaee-web-api</artifactId></dependency>
<dependency><groupId>javax</groupId><artifactId>javaee-api</artifactId></dependency>
<dependency><groupId>org.jboss.spec.javax.jms</groupId><artifactId>jboss-jms-api_2.0_spec</artifactId></dependency>
<dependency><groupId>org.jboss.spec.javax.rmi</groupId><artifactId>jboss-rmi-api_1.0_spec</artifactId></dependency>
<plugin><artifactId>maven-war-plugin</artifactId></plugin>
<plugin><artifactId>maven-ear-plugin</artifactId></plugin>
```

Also remove any `<scope>system</scope>` + `<systemPath>` dependencies — flag these in a TODO comment.

### Add

```xml
<properties>
  <quarkus.platform.version>3.8.6</quarkus.platform.version>
  <maven.compiler.source>17</maven.compiler.source>
  <maven.compiler.target>17</maven.compiler.target>
  <maven.compiler.release>17</maven.compiler.release>
</properties>

<dependencyManagement>
  <dependencies>
    <dependency>
      <groupId>io.quarkus.platform</groupId>
      <artifactId>quarkus-bom</artifactId>
      <version>${quarkus.platform.version}</version>
      <type>pom</type>
      <scope>import</scope>
    </dependency>
  </dependencies>
</dependencyManagement>

<dependencies>
  <dependency><groupId>io.quarkus</groupId><artifactId>quarkus-arc</artifactId></dependency>
  <dependency><groupId>io.quarkus</groupId><artifactId>quarkus-rest</artifactId></dependency>
  <dependency><groupId>io.quarkus</groupId><artifactId>quarkus-rest-jackson</artifactId></dependency>
  <dependency><groupId>io.quarkus</groupId><artifactId>quarkus-hibernate-orm</artifactId></dependency>
  <dependency><groupId>io.quarkus</groupId><artifactId>quarkus-jdbc-postgresql</artifactId></dependency>
  <dependency><groupId>io.quarkus</groupId><artifactId>quarkus-narayana-jta</artifactId></dependency>
  <dependency><groupId>io.quarkus</groupId><artifactId>quarkus-smallrye-reactive-messaging</artifactId></dependency>
  <dependency><groupId>io.quarkus</groupId><artifactId>quarkus-hibernate-validator</artifactId></dependency>
</dependencies>

<build>
  <plugins>
    <plugin>
      <groupId>io.quarkus.platform</groupId>
      <artifactId>quarkus-maven-plugin</artifactId>
      <version>${quarkus.platform.version}</version>
      <extensions>true</extensions>
      <executions>
        <execution>
          <goals><goal>build</goal><goal>generate-code</goal></goals>
        </execution>
      </executions>
    </plugin>
  </plugins>
</build>
```

### Change
- `<packaging>war</packaging>` → `<packaging>jar</packaging>`
- `<source>1.8</source>` / `<target>1.8</target>` → remove (covered by maven.compiler.release=17)

## Import Replacements (all .java files)

```
javax.inject           → jakarta.inject
javax.enterprise       → jakarta.enterprise
javax.ws.rs            → jakarta.ws.rs
javax.persistence      → jakarta.persistence
javax.transaction      → jakarta.transaction
javax.validation       → jakarta.validation
javax.annotation       → jakarta.annotation
javax.json             → jakarta.json
javax.ejb              → (remove entire import — see annotation transforms)
javax.jms              → (remove — replaced by reactive messaging)
javax.servlet          → (remove — not used in Quarkus)
```

## Annotation Transformations

| Remove | Replace With | Notes |
|--------|-------------|-------|
| `@Stateless` | `@ApplicationScoped` | Add `@Transactional` if methods mutate data |
| `@Stateful` | `@ApplicationScoped` | Externalize state if needed |
| `@Singleton` (javax.ejb) | `@ApplicationScoped` | Add `@Startup` if eager init required |
| `@EJB` | `@Inject` | |
| `@LocalBean` | (delete) | |
| `@Remote` | (delete) | Convert to REST if remote access needed |
| `@Local` | (delete) | |
| `@MessageDriven(...)` | `@ApplicationScoped` | See JMS section below |
| `@ActivationConfigProperty(...)` | (delete) | Config moves to application.properties |
| `@PersistenceContext` | `@Inject` | |
| `@TransactionAttribute(REQUIRED)` | `@Transactional` | |
| `@TransactionAttribute(REQUIRES_NEW)` | `@Transactional(TxType.REQUIRES_NEW)` | |
| `@ApplicationPath("/services")` | Keep or delete | Quarkus auto-discovers; keep if custom path needed |

## JMS → Reactive Messaging

### Message-Driven Beans

**Before:**
```java
@MessageDriven(name = "OrderServiceMDB", activationConfig = {
    @ActivationConfigProperty(propertyName = "destinationLookup", propertyValue = "topic/orders"),
    @ActivationConfigProperty(propertyName = "destinationType", propertyValue = "javax.jms.Topic")
})
public class OrderServiceMDB implements MessageListener {
    public void onMessage(Message msg) { ... }
}
```

**After:**
```java
@ApplicationScoped
public class OrderServiceMDB {
    @Incoming("orders")
    public void onMessage(String orderJson) { ... }
}
```

### application.properties for messaging:
```properties
mp.messaging.incoming.orders.connector=smallrye-in-memory
# or for Kafka: mp.messaging.incoming.orders.connector=smallrye-kafka
# or for AMQP: mp.messaging.incoming.orders.connector=smallrye-amqp
```

## Configuration Migration

### persistence.xml → application.properties

Delete `src/main/resources/META-INF/persistence.xml` and add to `src/main/resources/application.properties`:

```properties
quarkus.datasource.db-kind=postgresql
quarkus.datasource.jdbc.url=jdbc:postgresql://localhost:5432/coolstore
quarkus.datasource.username=coolstore
quarkus.datasource.password=coolstore
quarkus.hibernate-orm.database.generation=none
quarkus.hibernate-orm.log.sql=false
```

### Datasource mapping from persistence.xml:
- `java:jboss/datasources/*` → extract DB kind from JNDI name or driver
- `javax.persistence.schema-generation.database.action` → `quarkus.hibernate-orm.database.generation`
- `hibernate.show_sql` → `quarkus.hibernate-orm.log.sql`

## Structural Changes

### Delete these files:
- `src/main/webapp/WEB-INF/web.xml`
- `src/main/webapp/WEB-INF/jboss-web.xml`
- `src/main/webapp/WEB-INF/beans.xml` (Quarkus uses annotated discovery by default)
- `src/main/resources/META-INF/ejb-jar.xml`
- Any `*-ds.xml` datasource descriptors
- Any WebLogic/JBoss-specific config files

### Move static resources:
- `src/main/webapp/*.html|*.jsp|*.css|*.js` → `src/main/resources/META-INF/resources/`
- JSP files: convert to static HTML or remove (Quarkus doesn't support JSP)

### Remove WebLogic/vendor-specific classes:
- Delete any `com.weblogic.*` or `weblogic.*` source files
- Delete `ApplicationLifecycleListener` implementations → use `@Observes StartupEvent` instead

### EntityManager producer:
If there's a CDI producer for EntityManager (e.g. `Resources.java`), delete it — Quarkus provides `@Inject EntityManager` out of the box.

## Validation Command

```bash
mvn compile -DskipTests -q
```

Success = structural migration complete. Test failures are acceptable — they indicate business logic review, not migration defects.

## Known Issues

- System-scoped JARs (`<scope>system</scope>`) need manual resolution — check `lib/` directory
- Flyway: replace `org.flywaydb:flyway-core` with `io.quarkus:quarkus-flyway` (and move migrations to `src/main/resources/db/migration/`)
- JSP pages cannot be migrated automatically — flag as TODO
- `@Startup` + `@Singleton` for app init → use `void onStart(@Observes StartupEvent ev)`
