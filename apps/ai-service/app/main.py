"""FastAPI application entrypoint for SMARTREVERSAL AI Service."""

from fastapi import FastAPI

app = FastAPI(
    title="SMARTREVERSAL AI Intelligence Service",
    version="0.1.0",
)


@app.get("/health")
def health_check():
    return {"status": "operational", "service": "ai-service"}
