FROM alpine
# 添加时区文件
ADD zoneinfo.tar.gz /
# 解压时区文件
RUN cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime
# 设置工作目录
WORKDIR /app
# 添加可执行文件
COPY cmd/cmd /app/
# 添加所需的配置文件
COPY configs/ /app/configs/
# 开放端口
EXPOSE 9000
# 运行
#CMD ["/cmd", "-conf", "/configs"]
CMD ["./cmd"]