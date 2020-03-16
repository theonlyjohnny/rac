#!/bin/sh

set -e

./run_k3d.sh
./run_registry.sh
./run_notifications.sh
