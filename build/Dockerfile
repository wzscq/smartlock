FROM alpine:3.15
RUN set -eux &&sed -i 's/dl-cdn.alpinelinux.org/mirrors.ustc.edu.cn/g' /etc/apk/repositories
RUN apk add nginx
COPY nginx.conf /etc/nginx/nginx.conf
ADD ./package/service/smartlockservice /services/smartlock/smartlockservice
copy entrypoint.sh /entrypoint.sh
ENTRYPOINT ["sh","entrypoint.sh"]