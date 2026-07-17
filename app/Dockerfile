# Stage 1: Build the application
FROM maven:3.9.6-eclipse-temurin-21-alpine AS build
WORKDIR /app

# Build acceleration strategy: Copy only pom.xml first and download
# dependencies. This layer is cached as long as pom.xml doesn't change,
# so subsequent builds skip the expensive dependency download.
COPY pom.xml .
RUN mvn dependency:go-offline -B

# Copy the source code and package the application.
# This layer only re-runs when source code changes.
COPY src ./src
RUN mvn package -DskipTests -B

# Stage 2: Create a minimal, secure runtime image
FROM eclipse-temurin:21-jre-alpine

LABEL maintainer="crastle" \
      description="Sample Java App for DevOps Case Studies" \
      version="1.0"

# Security Best Practice: Run as non-root user
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Create log directory (for async file logging - Step 1.7)
RUN mkdir -p /var/log/app && chown appuser:appgroup /var/log/app

USER appuser:appgroup
WORKDIR /app

# Copy the built jar from the build stage
COPY --from=build --chown=appuser:appgroup /app/target/*.jar app.jar

EXPOSE 8080

# Health check for container orchestrators
HEALTHCHECK --interval=30s --timeout=3s --retries=3 \
  CMD wget -qO- http://localhost:8080/actuator/health || exit 1

ENTRYPOINT ["java", \
  "-XX:+UseG1GC", \
  "-XX:MaxRAMPercentage=75.0", \
  "-Djava.security.egd=file:/dev/./urandom", \
  "-jar", "app.jar"]
