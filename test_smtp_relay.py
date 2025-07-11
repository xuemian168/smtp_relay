#!/usr/bin/env python3
"""
SMTP中继服务测试脚本
测试用户创建的SMTP凭据是否能正常工作
"""

import smtplib
import ssl
import sys
import time
from email.mime.text import MIMEText
from email.mime.multipart import MIMEMultipart
from email.utils import formatdate
import argparse

# 用户的SMTP凭据信息
SMTP_CREDENTIALS = {
    'username': 'relay_687057f1_66a7',
    'password': 'f0278935404826a1d738b220ed2d6b95',
    'user_id': '687057f17058540b74a19a75',
    'credential_id': '68705c72ef7445ddb0602bd2'
}

def check_port_available(host, port, timeout=5):
    """检查端口是否可用"""
    import socket
    try:
        sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        sock.settimeout(timeout)
        result = sock.connect_ex((host, port))
        sock.close()
        return result == 0
    except Exception:
        return False

def test_smtp_connection(host, port, use_tls=False, use_ssl=False):
    """测试SMTP连接"""
    print(f"\n🔍 测试SMTP连接: {host}:{port}")
    print(f"   TLS: {'启用' if use_tls else '禁用'}")
    print(f"   SSL: {'启用' if use_ssl else '禁用'}")
    
    # 首先检查端口是否可用
    if not check_port_available(host, port):
        print(f"❌ 端口 {port} 不可达")
        print(f"   请检查服务是否运行在 {host}:{port}")
        return None
    
    print(f"✅ 端口 {port} 可达")
    
    try:
        if use_ssl:
            # 使用SSL连接 (端口465)
            context = ssl.create_default_context()
            server = smtplib.SMTP_SSL(host, port, context=context, timeout=10)
        else:
            # 使用普通连接或STARTTLS
            server = smtplib.SMTP(host, port, timeout=10)
            if use_tls:
                server.starttls()
        
        # 获取服务器问候信息
        print(f"✅ SMTP连接成功")
        ehlo_response = server.ehlo()
        if ehlo_response[0] == 250:
            print(f"   服务器响应: {ehlo_response[1].decode().strip()}")
        
        return server
    except Exception as e:
        print(f"❌ SMTP连接失败: {e}")
        return None

def test_smtp_auth(server, username, password):
    """测试SMTP认证"""
    print(f"\n🔐 测试SMTP认证")
    print(f"   用户名: {username}")
    print(f"   密码: {'*' * len(password)}")
    
    try:
        server.login(username, password)
        print(f"✅ 认证成功")
        return True
    except Exception as e:
        print(f"❌ 认证失败: {e}")
        return False

def send_test_email(server, from_email, to_email, subject="SMTP中继测试邮件"):
    """发送测试邮件"""
    print(f"\n📧 发送测试邮件")
    print(f"   发件人: {from_email}")
    print(f"   收件人: {to_email}")
    print(f"   主题: {subject}")
    
    try:
        # 创建邮件内容
        msg = MIMEMultipart()
        msg['From'] = from_email
        msg['To'] = to_email
        msg['Subject'] = subject
        msg['Date'] = formatdate(localtime=True)
        
        # 邮件正文
        body = f"""
这是一封来自SMTP中继服务的测试邮件。

测试信息:
- 发送时间: {time.strftime('%Y-%m-%d %H:%M:%S')}
- 凭据ID: {SMTP_CREDENTIALS['credential_id']}
- 用户ID: {SMTP_CREDENTIALS['user_id']}
- SMTP用户名: {SMTP_CREDENTIALS['username']}

如果您收到这封邮件，说明SMTP中继服务工作正常！

---
SMTP Relay Service Test
        """
        
        msg.attach(MIMEText(body, 'plain', 'utf-8'))
        
        # 发送邮件
        text = msg.as_string()
        server.sendmail(from_email, [to_email], text)
        
        print(f"✅ 邮件发送成功")
        return True
    except Exception as e:
        print(f"❌ 邮件发送失败: {e}")
        return False

def test_smtp_relay_service(host, ports, from_email, to_email):
    """完整的SMTP中继服务测试"""
    print("=" * 60)
    print("🚀 SMTP中继服务测试开始")
    print("=" * 60)
    
    print(f"📋 测试配置:")
    print(f"   服务器地址: {host}")
    print(f"   测试端口: {ports}")
    print(f"   SMTP用户名: {SMTP_CREDENTIALS['username']}")
    print(f"   发件人: {from_email}")
    print(f"   收件人: {to_email}")
    
    success_count = 0
    total_tests = len(ports)
    
    for port_config in ports:
        port = port_config['port']
        use_tls = port_config.get('tls', False)
        use_ssl = port_config.get('ssl', False)
        
        print(f"\n{'='*40}")
        print(f"🔧 测试端口 {port}")
        print(f"{'='*40}")
        
        # 测试连接
        server = test_smtp_connection(host, port, use_tls, use_ssl)
        if not server:
            continue
        
        # 测试认证
        if not test_smtp_auth(server, SMTP_CREDENTIALS['username'], SMTP_CREDENTIALS['password']):
            server.quit()
            continue
        
        # 发送测试邮件
        if send_test_email(server, from_email, to_email):
            success_count += 1
        
        # 关闭连接
        try:
            server.quit()
            print(f"🔌 连接已关闭")
        except:
            pass
    
    # 测试结果总结
    print(f"\n{'='*60}")
    print(f"📊 测试结果总结")
    print(f"{'='*60}")
    print(f"✅ 成功测试: {success_count}/{total_tests}")
    print(f"❌ 失败测试: {total_tests - success_count}/{total_tests}")
    
    if success_count > 0:
        print(f"\n🎉 SMTP中继服务测试通过！")
        print(f"   您的SMTP凭据可以正常使用")
        print(f"   建议使用成功的端口配置")
    else:
        print(f"\n⚠️  SMTP中继服务测试失败！")
        print(f"   请检查服务器状态和网络连接")
    
    return success_count > 0

def main():
    parser = argparse.ArgumentParser(description='SMTP中继服务测试工具')
    parser.add_argument('--host', default='localhost', help='SMTP服务器地址 (默认: localhost)')
    parser.add_argument('--from-email', required=True, help='发件人邮箱地址')
    parser.add_argument('--to-email', required=True, help='收件人邮箱地址')
    parser.add_argument('--port', type=int, help='指定单个端口测试')
    
    args = parser.parse_args()
    
    # 默认测试端口配置
    if args.port:
        # 单端口测试
        if args.port == 465:
            ports = [{'port': args.port, 'ssl': True}]
        elif args.port == 587:
            ports = [{'port': args.port, 'tls': True}]
        else:
            ports = [{'port': args.port}]
    else:
        # 多端口测试 (Docker环境端口映射)
        ports = [
            {'port': 2525, 'tls': False, 'ssl': False},  # SMTP中继端口 (Docker映射，支持认证)
            {'port': 587, 'tls': False, 'ssl': False},   # SMTP提交端口 (暂不使用TLS)
            {'port': 465, 'tls': False, 'ssl': False},   # SMTP端口 (暂不使用SSL)
        ]
    
    # 执行测试
    success = test_smtp_relay_service(args.host, ports, args.from_email, args.to_email)
    
    # 退出码
    sys.exit(0 if success else 1)

if __name__ == '__main__':
    main() 