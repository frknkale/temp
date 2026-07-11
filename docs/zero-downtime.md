# Zero-Downtime Deployment Approach

This document explains the strategy for achieving zero-downtime deployments for the Java application (Step 2.3.c).

## Strategy: Rolling Updates

Kubernetes supports zero-downtime deployments out of the box using **RollingUpdate** strategies for Deployments.

### Key Components

1. **`maxUnavailable` and `maxSurge`:**
   In the Deployment specification, the update strategy should be explicitly configured (or relying on defaults of 25% for both). 
   - `maxSurge`: Allows Kubernetes to spin up extra Pods above the desired replica count during the rollout.
   - `maxUnavailable`: Ensures that a minimum number of Pods are always running and available to handle traffic.

2. **Readiness Probes:**
   We have configured a `readinessProbe` pointing to the `/actuator/health` endpoint of the Spring Boot app.
   - **Why this is critical:** Kubernetes will not send traffic to the new Pods until the readiness probe passes. This guarantees that traffic is only routed to the new Pod when the Java application has fully started up and connected to its dependencies.
   
3. **Liveness Probes:**
   The `livenessProbe` ensures that if a Pod enters a broken state post-startup, it is automatically restarted.

4. **Graceful Shutdown (SIGTERM):**
   When Kubernetes scales down old Pods, it sends a `SIGTERM` signal. The Spring Boot application should gracefully handle this by stopping new connections and finishing processing in-flight requests.
   - In `application.properties`: `server.shutdown=graceful` is recommended.

### Rollout Process
During deployment, Kubernetes will:
1. Create a new ReplicaSet.
2. Spin up new Pods based on the `maxSurge` value.
3. Wait for the `readinessProbe` of the new Pods to pass.
4. Add the new Pods to the Service endpoints.
5. Send `SIGTERM` to the old Pods (up to `maxUnavailable`).
6. Repeat until all Pods are running the new version.

This ensures the user never experiences a dropped request or a 502 Bad Gateway error during the deployment pipeline.
