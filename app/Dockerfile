# Stage 1: Build the application
FROM maven:3.9.6-eclipse-temurin-21-alpine AS build
WORKDIR /app

# Accelerate build process by leveraging Docker caching.
# First, copy only the pom.xml and download the dependencies.
COPY pom.xml .
RUN mvn dependency:go-offline -B

# Copy the source code and package the application.
COPY src ./src
RUN mvn package -DskipTests

# Stage 2: Create the runtime image
FROM eclipse-temurin:21-jre-alpine

# Security Best Practice: Create a non-root user and group
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser:appgroup

WORKDIR /app

# Copy the built jar from the build stage.
COPY --from=build /app/target/*.jar app.jar

# Expose the application port
EXPOSE 8080

ENTRYPOINT ["java", "-jar", "app.jar"]
