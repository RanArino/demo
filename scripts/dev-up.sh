#!/bin/bash

# Start development environment with hot reloading
echo "Starting development environment with hot reloading..."
docker-compose -f docker-compose.yml -f docker-compose.dev.yml up -d

echo ""
echo "Development environment started!"
echo ""
echo "To restart a service after code changes:"
echo "  docker-compose -f docker-compose.yml -f docker-compose.dev.yml restart <service_name>"
echo ""
echo "To view logs:"
echo "  docker-compose -f docker-compose.yml -f docker-compose.dev.yml logs -f <service_name>"
echo ""
echo "To stop the development environment:"
echo "  docker-compose -f docker-compose.yml -f docker-compose.dev.yml down"
echo ""