FROM --platform=$BUILDPLATFORM golang:1.23.1-bullseye

WORKDIR /app/espresso/

# Create application directories
RUN mkdir -p /app/espresso/configs \
    && mkdir -p /app/espresso/output \
    && mkdir -p /app/espresso/service/tmp \
    && mkdir -p /app/espresso/service/inputfiles

# Configure Go to bypass proxy for private repositories
# Set ENABLE_UI to false to disable the UI
ENV GOSUMDB=off \
    GO111MODULE=on \
    ROD_BROWSER_BIN=/usr/bin/chromium \
    ENABLE_UI=true
    

# Copy go.mod and go.sum first to leverage Docker cache
COPY service/go.mod service/go.sum ./service/

# Download dependencies
RUN cd service && \
    go mod download
# Download dependencies

# Install browser dependencies, Chromium, and netcat
RUN apt-get update && apt-get install -y \
    fonts-liberation \
    libappindicator3-1 \
    libasound2 \
    libatk-bridge2.0-0 \
    libatk1.0-0 \
    libcups2 \
    libnss3 \
    libxcomposite1 \
    libxdamage1 \
    libxrandr2 \
    libdrm2 \
    libgbm1 \
    libxshmfence1 \
    libx11-xcb1 \
    chromium \
    netcat \
    --no-install-recommends \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app/espresso/service

COPY ./service /app/espresso/service/
COPY ./service/configs/ /app/espresso/configs/ 

# Add a non-root user for running Chrome with sandbox
RUN groupadd -r chrome && useradd -r -g chrome -G audio,video chrome \
    && mkdir -p /home/chrome/Downloads \
    && chown -R chrome:chrome /home/chrome \
    && chown -R chrome:chrome /app/espresso


# Set proper permissions
RUN chown -R chrome:chrome /app

# Set the user to chrome for the container
USER chrome

# To download dependencies as non root user
ENV GOPATH=/home/chrome/go
RUN mkdir -p $GOPATH/pkg/mod && chmod -R 777 $GOPATH/pkg/mod

EXPOSE 8081

# Change the CMD to run the Go application directly
CMD ["go", "run","-mod=mod", "/app/espresso/service/main.go"]
