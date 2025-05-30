FROM golang:1.24

# Switch workdir do build directory
WORKDIR /omnom/build

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build

# Switch workdir to final destination
WORKDIR /omnom

# Copy binary, remove build dir
RUN cp ./build/omnom .
RUN rm -rf ./build

# Install required utilities for user/group management
RUN apt-get update && apt-get install -y --no-install-recommends \
    gosu \
    && rm -rf /var/lib/apt/lists/*

# Create the entrypoint script
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

# Modify config.yml to provide some needed default values for running inside the container
COPY config.yml_sample config.yml
RUN sed -i -E \
    # Listen to any address by default
    -e 's/address: "127\.0\.0\.1:7331"/address: "0.0.0.0:7331"/' \
    # No browser is installed, disable bookmark creation by default
    -e 's/create_bookmark_from_webapp: true/create_bookmark_from_webapp: false/' \
    # Move database and keys to config/ by default to allow users to bind mount this directory
    -e 's@connection: "\./db\.sqlite3"@connection: "./config/db.sqlite3"@' \
    # Move the activity pub keys to config/ by default to allow users to bind mount this directory
    -e 's#(pubkey|privkey): "\./([^"]+)"#\1: "./config/\2"#g' config.yml

VOLUME /omnom/config
VOLUME /omnom/static/data

EXPOSE 7331

ENTRYPOINT ["/entrypoint.sh"]
CMD ["/omnom/omnom", "listen"]