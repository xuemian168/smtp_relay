// MongoDB初始化脚本
// 该脚本在MongoDB容器启动时执行

// 切换到smtp_relay数据库
db = db.getSiblingDB('smtp_relay');

// 创建管理员用户（如果需要）
// db.createUser({
//   user: 'smtp_admin',
//   pwd: 'password',
//   roles: [
//     {
//       role: 'readWrite',
//       db: 'smtp_relay'
//     }
//   ]
// });

// 创建初始系统配置
db.system_configs.insertMany([
  {
    key: 'smtp_server_name',
    value: 'SMTP Relay Server',
    type: 'string'
  },
  {
    key: 'default_daily_quota',
    value: 1000,
    type: 'int'
  },
  {
    key: 'default_hourly_quota',
    value: 100,
    type: 'int'
  },
  {
    key: 'max_email_size',
    value: 26214400, // 25MB in bytes
    type: 'int'
  },
  {
    key: 'max_recipients_per_email',
    value: 100,
    type: 'int'
  },
  {
    key: 'rate_limit_enabled',
    value: true,
    type: 'bool'
  },
  {
    key: 'registration_enabled',
    value: true,
    type: 'bool'
  }
]);

// 创建示例IP信誉记录
db.ip_reputation.insertOne({
  ip: '127.0.0.1',
  reputation_score: 100.0,
  success_rate: 0.0,
  total_sent: 0,
  total_failed: 0,
  last_checked: new Date(),
  status: 'good',
  updated_at: new Date()
});

print('MongoDB初始化完成');
print('数据库名称: smtp_relay');
print('创建了以下集合:');
print('- system_configs (系统配置)');
print('- ip_reputation (IP信誉监控)'); 