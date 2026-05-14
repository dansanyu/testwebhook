#!/bin/bash
set -e  # 出错就退出

PROJECT_DIR="/www/wwwroot/v2.dokey.cf/action"
REPO="git@github.com:dansanyu/testwebhook.git"
BRANCH="master"   # 如果仓库默认分支是 master，就改为 master

# 1️⃣ 创建目录（如果不存在）
mkdir -p $PROJECT_DIR
cd $PROJECT_DIR

# 2️⃣ 初始化 Git 仓库（第一次部署需要）
if [ ! -d ".git" ]; then
    git init
    git remote add origin $REPO
    git fetch origin $BRANCH
    git checkout -b $BRANCH FETCH_HEAD
fi

# 3️⃣ 拉取最新代码
git fetch origin master
git reset --hard origin/master

# 4️⃣ Docker 部署
# 检查 docker-compose.yml 是否存在
if [ ! -f "docker-compose.yml" ]; then
    echo "docker-compose.yml not found, please add it to the repo."
    exit 1
fi

docker compose down || true
docker compose up -d --build