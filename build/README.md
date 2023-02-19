# crv_frame_build

#导出镜像包命令
docker save -o smartlock.tar wangzhsh/smartlock:0.1.0
#导入镜像包命令
docker load -i smartlock.tar

#run smartlock in docker
docker run -d --name smartlock -p8301:80 -v /root/smartlock/conf:/services/smartlock/conf wangzhsh/smartlock:0.1.0

2023-02-19
1、补充了三个关联表：
    sl_application_sl_key_authoriza
    sl_application_sl_lock
    sl_application_sl_person
2、sl_application模型配置修改，修改了发下智能锁按钮的功能配置
3、external_api配置修改，增加了writekey接口
4、smartlockservice修改，增加writekey接口逻辑
5、修改配置文件，增加了和钥匙管理机通信的两个topic