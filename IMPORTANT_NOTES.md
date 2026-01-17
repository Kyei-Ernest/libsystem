# Important Notes for Local Development

## Infrastructure Configuration

> [!IMPORTANT]
> **Kafka**: Ensure Kafka is running on port **9093**.
> We changed this from the default 9092 to avoid conflicts with a system process.
> - **Command**: `bin/kafka-server-start.sh config/server.properties --override listeners=PLAINTEXT://:9093`

> [!IMPORTANT]
> **Elasticsearch**: Ensure Elasticsearch is running on port **9200** with **security disabled**.
> - **Config**: Set `xpack.security.enabled: false` in `config/elasticsearch.yml`.
> - **Verify**: `curl http://localhost:9200` should return JSON without authentication.

## Running the System

-   **Start All Services**:
    ```bash
    ./run-services.sh
    ```
    This script builds and starts the User, Collection, Document, Search, and Indexer services. It waits for them to become healthy before proceeding.

-   **Nginx Gateway**:
    The Nginx configuration (`nginx.conf`) has been updated to include the Search Service.
    -   **Reload**: `sudo nginx -s reload` (if running locally) or restart your Nginx container.
    -   **Search endpoint**: `http://localhost:8080/api/v1/search` (via Nginx) or `http://localhost:8084/api/v1/search` (direct).

## Verification

-   **End-to-End Test**:
    ```bash
    ./verify-search.sh
    ```
    This script registers a user, creates a collection, uploads a document, waits for indexing, and verifies that the document appears in search results.

## Troubleshooting

-   **Service Logs**: Logs are written to the `logs/` directory.
    -   `tail -f logs/*.log` to monitor all services.
-   **"Address already in use"**: If you see this for Kafka (9092), remember to use port **9093**.
-   **Elasticsearch Connection Refused**: Check if Elasticsearch is running and if `xpack.security.enabled` is false.
