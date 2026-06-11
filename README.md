# Gin + Kubernetes SDK Project

一个基于 Gin 框架并集成 Kubernetes SDK 的 Go 项目。

## 功能特性

- ✅ Gin Web 框架
- ✅ Kubernetes SDK 集成
- ✅ 支持集群内外运行
- ✅ RESTful API 接口

## 项目结构

```
.
├── main.go          # 主程序入口
├── go.mod           # Go 模块依赖
└── README.md        # 项目说明
```

## 快速开始

### 1. 安装依赖

```bash
go mod download
```

### 2. 运行项目

#### 在 Kubernetes 集群内运行

直接部署到集群,会自动使用 ServiceAccount 访问 API。

#### 在集群外运行 (本地开发)

需要配置 kubeconfig:

```bash
export KUBECONFIG=/path/to/kubeconfig
go run main.go
```

或使用默认的 `~/.kube/config`:

```bash
go run main.go
```

### 3. API 接口

#### 健康检查
```bash
curl http://localhost:8080/health
```

#### 获取所有命名空间
```bash
curl http://localhost:8080/api/v1/namespaces
```

#### 获取 Pod 列表
```bash
# 获取 default 命名空间的 pods
curl http://localhost:8080/api/v1/pods

# 获取指定命名空间的 pods
curl http://localhost:8080/api/v1/pods?namespace=kube-system
```

### 4. 环境变量

- `PORT`: 服务端口,默认 8080
- `KUBECONFIG`: Kubernetes 配置文件路径

## 构建 Docker 镜像

```bash
docker build -t gin-k8s-app:latest .
```

## 部署到 Kubernetes

创建 RBAC 配置:

```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: gin-k8s-app
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: gin-k8s-app
rules:
- apiGroups: [""]
  resources: ["namespaces", "pods"]
  verbs: ["get", "list", "watch"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: gin-k8s-app
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: gin-k8s-app
subjects:
- kind: ServiceAccount
  name: gin-k8s-app
  namespace: default
```

## 许可证

MIT License
