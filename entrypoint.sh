#!/bin/sh
set -e

# Default values if not provided
UID=${UID:-1000}
GID=${GID:-1000}

# Create group if it doesn't exist
if ! getent group omnomgroup >/dev/null; then
    addgroup --gid $GID omnomgroup
fi

# Create user if it doesn't exist
if ! id omnomuser >/dev/null 2>&1; then
    adduser --disabled-password --gecos "" --shell /bin/sh \
      --uid $UID --gid $GID omnomuser
fi

LOCAL_UID=$(id -u omnomuser)
LOCAL_GID=$(getent group omnomgroup | cut -d ":" -f 3)

if [ ! "$UID" = "$LOCAL_UID" ] || [ ! "$GID" = "$LOCAL_GID" ]; then
    echo "Warning: User with differing UID $LOCAL_UID/GID $LOCAL_GID already exists, most likely this container was started before with a different UID/GID. Re-create it to change UID/GID."
fi

# Change ownership of the application directory
chown -R $UID:$GID /omnom

# Run as the specified user
exec gosu $UID:$GID "$@"