#!/bin/sh
set -e

# Start supervisord (will keep processes supervised)
exec /usr/bin/supervisord -c /app/supervisord.conf
