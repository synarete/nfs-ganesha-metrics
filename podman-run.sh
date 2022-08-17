#!/bin/sh
DBUS_SOCKET=/var/run/dbus/system_bus_socket
podman run -it --network=host --volume ${DBUS_SOCKET}:${DBUS_SOCKET} "$@"
  
