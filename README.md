# 基础平台
Forked from: [Casdoor](https://github.com/casdoor/casdoor)  
提供基础服务， 包括
- 用户管理
- 角色管理
- 权限管理
- 对象资源管理(图片、文件、视频)




## 用户管理
### 新增用户
前提：
1. 已创建[组织](#组织管理-1)
2. 组织内存在[APP](#app管理)（商城、后台管理）

接口定义:




## 组织管理
### 组织管理
- 每个组织代表一个公司、企业。
- 组织间资源完全隔离。
- 组织内存在多个分组group
#### 新增组织






### group管理
 - 每个分组代表一个部门、子公司
 - 分组用于用户管理，group:user= n:n


## app管理
- 每个APP代表组织的一个应用。
- 当前一个组织至少存在两个APP： 后台和商城





## 商城管理
在商城系统中
- 商城:app:group= 1:1:1
- app记录商城自身的属性：商城url、图标、登录方式
- group记录商城的用户信息



## 权限管理
## 角色管理



# 服务部署
## 重新生成swagger
https://casdoor.org/zh/docs/developer-guide/swagger/
```bash
mybee generate docs --tags "Group API,Auth API,Application API"
```



## 初始化服务
根据`init_data.json`生成默认组织、角色

### 使用CLI命令进行初始化

可以通过如下命令执行初始化流程（仅初始化数据库和基础数据，不启动Web服务）：

```bash
./casdoor init --createDatabase  --initFile ./init_data.json 
```

执行后会自动完成数据库表结构和基础数据的初始化。

# 环境变量配置

本项目支持通过 `.env` 文件加载环境变量。你可以参考 `.env.example` 文件，复制为 `.env` 并根据实际需求修改。

.env 配置内容会覆盖app.conf配置

项目启动时会自动加载 `.env` 文件中的变量。

