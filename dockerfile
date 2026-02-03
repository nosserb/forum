FROM ubuntu:24.04
RUN apt-get update && apt-get install -y sqlite3 && rm -rf /var/lib/apt/lists/*
WORKDIR /
COPY main .
COPY model ./model
COPY view ./view
CMD ["./main"]
