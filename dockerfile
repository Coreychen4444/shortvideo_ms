# 第一阶段: 构建Go应用
# 使用官方的golang镜像作为构建环境。这个镜像包含了Go编译器和所有你需要构建Go应用程序的工具。
FROM golang:1.20 AS build-env

# 设置工作目录。这是后续命令默认的运行目录。
WORKDIR /app

# 拷贝go.mod和go.sum到容器中。这些文件定义了应用程序的依赖。
COPY go.mod go.sum ./

# 下载所有依赖。
RUN go mod download

# 拷贝源代码文件到容器中。
COPY . .

# 使用Go构建应用程序。这将在/app目录下生成一个名为"api_service"的二进制文件。
RUN go build -o api_service

# 第二阶段: 运行阶段
# 使用轻量级的alpine Linux镜像作为基础。
FROM alpine:latest

# 从build-env阶段拷贝构建好的Go应用到新的镜像中。
COPY --from=build-env /app/api_service /app/api_service

# 设置工作目录为/app。这样CMD或ENTRYPOINT中的命令都会在这个目录下执行。
WORKDIR /app

# 指定容器启动时运行的命令。这里是启动你的视频服务。
CMD ["./api_service"]
