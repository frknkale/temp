# Application Logging Strategy

This document addresses Step 1.7 of the assignment requirements regarding the design and implementation of application logging.

## Problem Statement

Currently, the system reads logs through `stdout` and forwards them to Elasticsearch. Printing logs to `stdout` in high-throughput applications can negatively impact performance due to blocking I/O operations and context switching.

## Proposed Design & Implementation

To resolve this issue, the application will be configured to write logs asynchronously directly to a file, applying specific retention and rotation policies. Since the application is built using Spring Boot, we can leverage the built-in `Logback` logging framework.

### a. Write application logs to a file asynchronously

**Design:** 
By wrapping the traditional `FileAppender` with an `AsyncAppender` in Logback, logging calls from the application threads become non-blocking. The `AsyncAppender` acts as a buffer (using a blocking queue). A separate worker thread then dequeues these log events and writes them to the file system.

**Implementation (logback-spring.xml):**
```xml
<appender name="FILE" class="ch.qos.logback.core.rolling.RollingFileAppender">
    <file>/var/log/app/application.log</file>
    <encoder>
        <pattern>%d{yyyy-MM-dd HH:mm:ss} [%thread] %-5level %logger{36} - %msg%n</pattern>
    </encoder>
    <!-- ... rotation policies ... -->
</appender>

<appender name="ASYNC_FILE" class="ch.qos.logback.classic.AsyncAppender">
    <!-- Queue size and behavior configuration -->
    <queueSize>512</queueSize>
    <discardingThreshold>0</discardingThreshold>
    <appender-ref ref="FILE" />
</appender>

<root level="INFO">
    <appender-ref ref="ASYNC_FILE" />
</root>
```

### b. Ensure the log file size does not exceed 1GB
### c. Rotate log files daily

**Design:**
To achieve both constraints simultaneously (size-based and time-based rolling), we will use the `SizeAndTimeBasedRollingPolicy` provided by Logback. This policy rotates the file every day, but will also trigger a rotation if the file exceeds 1GB before the day is over.

**Implementation (added to the FILE appender):**
```xml
<rollingPolicy class="ch.qos.logback.core.rolling.SizeAndTimeBasedRollingPolicy">
    <!-- Daily rollover pattern with a sequence number -->
    <fileNamePattern>/var/log/app/application-%d{yyyy-MM-dd}.%i.log</fileNamePattern>
    
    <!-- Each file should be at most 1GB -->
    <maxFileSize>1GB</maxFileSize>
    
    <!-- Keep 30 days of history, capped at total 10GB -->
    <maxHistory>30</maxHistory>
    <totalSizeCap>10GB</totalSizeCap>
</rollingPolicy>
```

### Infrastructure Integration
Once the application writes logs to `/var/log/app/application.log` (mounted as an emptyDir volume in Kubernetes), a **Fluent-bit** sidecar container or DaemonSet will be configured to tail this specific file asynchronously instead of reading from Docker/containerd stdout streams. This completely eliminates the standard output bottleneck while preserving log delivery to Elasticsearch.
