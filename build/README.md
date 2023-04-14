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

2023-03-10 增加锁信息同步功能，可以将系统中的所有锁信息同步到后台状态监控模块
1、修改以下配置文件
    sl_lock   增加同步锁信息按钮
    external_api配置修改，增加了syncLockList接口
2、smartlockservice修改，增加syncLockList接口逻辑

2023-04-01 补充智能锁网关配置
1、增加智能锁网关表sl_lock_hub
2、sl_lock表增加两个一对多关联字段，分别对应主从网关
3、增加sl_lock_hub模型配置
4、修改菜单配置，增加sl_lock_hub对应的入口菜单
5、sl_lock模型配置中增加网关字段

2023-04-01 钥匙管理机的名称改称IP地址
1、修改模型sl_key_controller的配置，将name字段改为ip
2、sl_key表增加钥匙管理机字段
3、开锁申请表增加管理机字段？暂时没有添加这个字段，再确认后增加

2023-04-02 和锁的通信机制的修改
1、通过tcp直接链接智能锁网关，并通过后台线程监控锁的状态
2、当锁状态发生变化时，通过MQTT消息触发将锁状态同步到mysql
3、修改锁状态记录表中的状态取值枚举范围，保持和锁厂家提供的状态码一致，
   注意这里修改了数据库表sl_lock_status_record.status字段长度从2改到了3
   升级时需要替换sl_lock_status_record模型配置
4、下发开锁指令时直接通过tcp方式下发

2023-04-05 细节改进
1、listview的数据排序按照时间倒序
    修改相关视图的默认排序字段配置，用last_udpate字段倒序
    sl_application
    sl_key_authorization
    sl_lock_authorization
2、下发开锁指令时，数据库中的锁的编号是10进制的，需要转换为16进制
3、远程开锁的闭锁延时允许再开锁时指定，下拉选择，选项包括：10s、30s、60s
    修改模型sl_lock_authorization，增加闭锁延时字段close_delay
    修改开锁接口逻辑，下发开锁指令前先下发闭锁延时指令

2023-04-12 细节改进
1、申请授权中的2个时间去掉，日期改为带时间的日期
    修改模型sl_application的相关配置
    后端服务逻辑修改
2、智能钥匙授权的时候需要弹出对话框，让用户选择一个钥匙管理机
    修改模型sl_application的相关配置
    后端服务逻辑修改

2023-04-14 增加钥匙授权记录功能
1、修改sl_key_authorization模型表结构和对应配置
2、后台逻辑增加mqtt收取信息并入库的功能

