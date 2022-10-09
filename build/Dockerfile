FROM golang:1.18.4 AS base

RUN apk add git && cd / && https://github.com/3Xpl0it3r/minio-operator.git 

ENV GOPROXY="https://goproxy.cn"

RUN cd /minio-operator/cmd/minio-operator

FROM alpine

RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.ustc.edu.cn/g' /etc/apk/repositories && apk add tzdata && cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && echo "Asia/Shanghai" > /etc/timezone && apk add libc6-compat

ADD miniooperator /opt/miniooperator

RUN chmod +x /opt/miniooperator

ENTRYPOINT ["/opt/miniooperator"]
