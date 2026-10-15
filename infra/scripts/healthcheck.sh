#!/bin/sh
# Check services health
curl -f http://localhost:8080/health || exit 1
curl -f http://localhost:8000/health || exit 1
