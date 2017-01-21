#!/bin/bash

BACKUP_DIR=${1:-"$HOME/PodsyncBackups"}

mkdir -p $BACKUP_DIR
docker-machine scp podsync:/data/redis/appendonly.aof $BACKUP_DIR/redis.$(date '+%Y_%m_%d__%H_%M_%S').aof