"""Anomaly detection endpoints."""

from fastapi import APIRouter

router = APIRouter(prefix="/anomaly", tags=["anomaly"])
