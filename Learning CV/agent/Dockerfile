FROM python:3.11-slim

WORKDIR /service

RUN apt-get update && apt-get install -y \
    libgl1 \
    libglib2.0-0 \
    && rm -rf /var/lib/apt/lists/* \
    && mkdir -p /usr/local/lib/python3.11/

COPY pyproject.toml ./
COPY src ./src
COPY config.yaml /usr/local/lib/python3.11/

RUN pip install --no-cache-dir .

EXPOSE 8000

CMD ["uvicorn", "docscan.api:app", "--host", "0.0.0.0", "--port", "8000"]
