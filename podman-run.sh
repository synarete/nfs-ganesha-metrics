#!/bin/sh
podman run -it \
  --network=host \
  --env DBUS_SESSION_BUS_ADDRESS="$DBUS_SESSION_BUS_ADDRESS" \
  --user $(id -u):$(id -g) \
  --volume /run/user/$(id -u)/bus:/run/user/$(id -u)/bus \
  --volume /var/run/dbus:/var/run/dbus \
  "$@"
