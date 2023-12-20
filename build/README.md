# crv_frame_build

#更新镜像流程
1、停止并删除原有镜像实例
    docker stop smartlock
    docker rm smartlock
2、删除原有镜像
    docker rmi smartlock:0.1.0
3、加载新的镜像
    docker load -i smartlockXXXXX.tar
4、启动实例
    docker run 。。。

#导出镜像包命令
docker save -o smartlock.tar wangzhsh/smartlock:0.1.0
#导入镜像包命令
docker load -i smartlock.tar

#run smartlock in docker
docker run -d --name smartlock -p8301:80 -v /root/smartlock/conf:/services/smartlock/conf wangzhsh/smartlock:0.1.1

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

2023-04-23 应对现场hub可能存在环路问题
1、修改发送指令的逻辑，连接hub后先收取旧的数据并丢弃，然后发出命令，再收取反馈信息。

2023-05-03 优化开锁速度
1、增加了批量开锁操作openBatch
2、需要修改开锁按钮的配置和external_api配置

2023-05-07
1、修改了报表中第三个chat图的取数逻辑，增加了基于区域的过滤参数，让下方的折线图和第一个饼图数据保持一致
2、sl_lock模型配置中去掉了同步锁信息按钮
3、菜单中去掉系统管理下的日志管理和文件管理菜单
4、修改了框架中不可用控件的背景颜色

2023-05-13
1、修改了dashboard的配置，解决了统计数据不正确的问题，包括开锁状态错误，按日期统计开锁次数时没有按日期分组汇总，优化了对日期条件的框选方式
2、sl_lock_authorization、sl_key_authorization、sl_application补充了快速检索字段
3、框架增加了数据校验功能，sl_application补充了对结束日期必须大于开始日期的校验,sl_lock补充了主用网关和备用网关不能相同的校验

2023-11-17 i6000接口
1、数据库增加三个表存储工单相关信息
    sl_work_ticket
    sl_involve_system
    sl_involve_device
2、为新增加的表增加响应的模型配置
3、增加菜单项打开工单信息页面
4、后端逻辑增加从I6000获取工单的逻辑和响应接口

2023-11-29 更新
1、修改获取工单的逻辑，增加基于不同状态工单的处理。
    更新接口程序
    修改接口程序配置文件，增加了2个配置项，用于过滤有效的工单
2、需改sl_application配置，增加申请状态和对应的视图
    {"value":"4","label":"工单变更"}

2023-12-20 新客户版本
1、创建smartlockv3应用，并需改对用配置
2、后台服务中增加配置参数控制定时读取状态逻辑是否启动