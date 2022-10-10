FROM golang:1.18.4 AS base

ENV GOPROXY "https://goproxy.cn"

RUN mkdir /minio-operator

ADD . /minio-operator

RUN  cd /minio-operator/cmd/miniooperator && go build .


FROM alpine 

COPY --from=base /minio-operator/cmd/miniooperator/miniooperator /opt/

# RUN echo "nameserver 114.114.114.114" > /etc/resolv.conf &&  sed -i 's/dl-cdn.alpinelinux.org/mirrors.ustc.edu.cn/g' /etc/apk/repositories && apk add tzdata && cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && echo "Asia/Shanghai" > /etc/timezone && apk add libc6-compat && chmod +x /opt/miniooperator
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.ustc.edu.cn/g' /etc/apk/repositories && apk add tzdata && cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && echo "Asia/Shanghai" > /etc/timezone && apk add libc6-compat && chmod +x /opt/miniooperator



ENTRYPOINT ["/opt/miniooperator"]
