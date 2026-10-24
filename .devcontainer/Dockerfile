FROM golang:1.25-bookworm

# Non-root user: the devcontainer CLI aligns its UID/GID with the host user's at
# start-up, so files created in the bind-mounted workspace belong to that user.
RUN groupadd --gid 1000 vscode \
    && useradd --uid 1000 --gid 1000 --create-home --shell /bin/bash vscode
