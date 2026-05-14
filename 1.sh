#!/bin/bash
set -e

PROJECT_DIR="/www/www***/v2.dokey.cf/action"
REPO="git@github.com:dansanyu/testwebhook.git"
BRANCH="master"   # ⚠️ 如果你 GitHub 是 main，这里要改 main

mkdir -p $PROJECT_DIR
cd $PROJECT_DIR

# 如果第一次部署
if [ ! -d ".git" ]; then
    git init
    git remote add origin $REPO
fi

# 强制同步远程分支
git fetch origin

# 删除所有本地状态，直接对齐远程
git checkout -B $BRANCH origin/$BRANCH

# 清理脏文件
git clean -fd

# Docker
docker compose down || true
docker compose up -d --build