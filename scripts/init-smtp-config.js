// MongoDB初始化脚本 - 添加默认SMTP配置
// 用于配置上游SMTP服务器（容器化Postfix）

// 连接数据库
db = db.getSiblingDB('smtp_relay');

// 创建默认的SMTP配置
const defaultSMTPConfig = {
    name: "本地Postfix服务器",
    host: "postfix",  // 容器名称
    port: 25,         // 标准SMTP端口
    username: "",     // Postfix不需要认证（内网信任）
    password: "",
    tls: false,       // 内网通信不需要TLS
    active: true,
    max_connections: 10,
    timeout: 30,      // 30秒超时
    rate_limit: 100,  // 每分钟100封邮件
    priority: 1,      // 优先级
    created_at: new Date(),
    updated_at: new Date()
};

// 检查是否已存在配置
const existingConfig = db.smtp_configs.findOne({host: "postfix"});

if (!existingConfig) {
    // 插入默认配置
    const result = db.smtp_configs.insertOne(defaultSMTPConfig);
    print("已添加默认SMTP配置:", result.insertedId);
} else {
    print("默认SMTP配置已存在，跳过创建");
}

// 创建索引
db.smtp_configs.createIndex({host: 1, port: 1}, {unique: true});
db.smtp_configs.createIndex({active: 1});
db.smtp_configs.createIndex({priority: 1});

print("SMTP配置初始化完成");

// 显示当前配置
print("\n当前SMTP配置:");
db.smtp_configs.find().forEach(function(config) {
    print("- " + config.name + " (" + config.host + ":" + config.port + ") - " + 
          (config.active ? "启用" : "禁用"));
}); 