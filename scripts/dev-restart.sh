#!/bin/bash

# Restart specific service in development mode
if [ $# -eq 0 ]; then
    echo "Usage: $0 <service_name>"
    echo "Available services: frontend, ms_user, ms_knowledge, ms_document_process, ngrok"
    exit 1
fi

SERVICE_NAME=$1

echo "Restarting $SERVICE_NAME in development mode..."
docker-compose -f docker-compose.yml -f docker-compose.dev.yml restart $SERVICE_NAME

echo "$SERVICE_NAME restarted successfully!"
echo "To view logs: docker-compose -f docker-compose.yml -f docker-compose.dev.yml logs -f $SERVICE_NAME"