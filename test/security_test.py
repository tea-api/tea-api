#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
Tea API 安全防护测试脚本
用于验证安全防护机制是否有效工作
"""

import requests
import json
import time
import random
import string
from concurrent.futures import ThreadPoolExecutor, as_completed

# 测试配置
BASE_URL = "http://localhost:3000"  # 修改为你的API地址
API_KEY = "sk-test-key"  # 修改为你的测试API密钥

def generate_random_content(length):
    """生成随机内容"""
    return ''.join(random.choices(string.ascii_letters + string.digits, k=length))

def test_normal_request():
    """测试正常请求"""
    print("🧪 测试正常请求...")
    
    headers = {
        "Authorization": f"Bearer {API_KEY}",
        "Content-Type": "application/json"
    }
    
    data = {
        "model": "gpt-3.5-turbo",
        "messages": [
            {"role": "user", "content": "Hello, how are you?"}
        ],
        "max_tokens": 100
    }
    
    try:
        response = requests.post(f"{BASE_URL}/v1/chat/completions", 
                               headers=headers, json=data, timeout=10)
        print(f"✅ 正常请求状态码: {response.status_code}")
        return response.status_code == 200
    except Exception as e:
        print(f"❌ 正常请求失败: {e}")
        return False

def test_large_prompt_attack():
    """测试超长Prompt攻击"""
    print("🧪 测试超长Prompt攻击...")
    
    headers = {
        "Authorization": f"Bearer {API_KEY}",
        "Content-Type": "application/json"
    }
    
    # 生成超长随机内容
    large_content = generate_random_content(60000)
    
    data = {
        "model": "gpt-3.5-turbo",
        "messages": [
            {"role": "user", "content": large_content}
        ],
        "max_tokens": 100,
        "stream": True
    }
    
    try:
        response = requests.post(f"{BASE_URL}/v1/chat/completions", 
                               headers=headers, json=data, timeout=10)
        print(f"🛡️ 超长Prompt攻击状态码: {response.status_code}")
        
        if response.status_code in [413, 429, 403]:
            print("✅ 超长Prompt攻击被成功阻止")
            return True
        else:
            print("❌ 超长Prompt攻击未被阻止")
            return False
    except Exception as e:
        print(f"🛡️ 超长Prompt攻击被阻止: {e}")
        return True

def test_high_frequency_attack():
    """测试高频请求攻击"""
    print("🧪 测试高频请求攻击...")
    
    headers = {
        "Authorization": f"Bearer {API_KEY}",
        "Content-Type": "application/json"
    }
    
    data = {
        "model": "gpt-3.5-turbo",
        "messages": [
            {"role": "user", "content": "test"}
        ],
        "max_tokens": 10
    }
    
    blocked_count = 0
    total_requests = 20
    
    print(f"发送 {total_requests} 个高频请求...")
    
    for i in range(total_requests):
        try:
            response = requests.post(f"{BASE_URL}/v1/chat/completions", 
                                   headers=headers, json=data, timeout=5)
            if response.status_code in [429, 403]:
                blocked_count += 1
                print(f"🛡️ 请求 {i+1} 被阻止 (状态码: {response.status_code})")
            else:
                print(f"✅ 请求 {i+1} 通过 (状态码: {response.status_code})")
            
            # 短间隔发送请求
            time.sleep(0.05)
            
        except Exception as e:
            blocked_count += 1
            print(f"🛡️ 请求 {i+1} 被阻止: {e}")
    
    print(f"📊 高频攻击结果: {blocked_count}/{total_requests} 请求被阻止")
    return blocked_count > total_requests * 0.5  # 超过50%被阻止认为防护有效

def test_malicious_stream_attack():
    """测试恶意流攻击"""
    print("🧪 测试恶意流攻击...")
    
    headers = {
        "Authorization": f"Bearer {API_KEY}",
        "Content-Type": "application/json",
        "Accept": "text/event-stream"
    }
    
    # 生成大量随机内容的流请求
    large_content = generate_random_content(80000)
    
    data = {
        "model": "gpt-3.5-turbo",
        "messages": [
            {"role": "user", "content": large_content}
        ],
        "max_tokens": 1000,
        "stream": True
    }
    
    try:
        response = requests.post(f"{BASE_URL}/v1/chat/completions", 
                               headers=headers, json=data, 
                               timeout=10, stream=True)
        
        print(f"🛡️ 恶意流攻击状态码: {response.status_code}")
        
        if response.status_code in [413, 429, 403]:
            print("✅ 恶意流攻击被成功阻止")
            return True
        else:
            print("❌ 恶意流攻击未被阻止")
            return False
            
    except Exception as e:
        print(f"🛡️ 恶意流攻击被阻止: {e}")
        return True

def test_concurrent_streams():
    """测试并发流攻击"""
    print("🧪 测试并发流攻击...")
    
    def create_stream_request():
        headers = {
            "Authorization": f"Bearer {API_KEY}",
            "Content-Type": "application/json",
            "Accept": "text/event-stream"
        }
        
        data = {
            "model": "gpt-3.5-turbo",
            "messages": [
                {"role": "user", "content": "Generate a long story"}
            ],
            "max_tokens": 1000,
            "stream": True
        }
        
        try:
            response = requests.post(f"{BASE_URL}/v1/chat/completions", 
                                   headers=headers, json=data, 
                                   timeout=10, stream=True)
            return response.status_code
        except Exception as e:
            return 0  # 连接被拒绝
    
    # 并发创建多个流请求
    blocked_count = 0
    total_streams = 10
    
    with ThreadPoolExecutor(max_workers=total_streams) as executor:
        futures = [executor.submit(create_stream_request) for _ in range(total_streams)]
        
        for i, future in enumerate(as_completed(futures)):
            status_code = future.result()
            if status_code in [0, 429, 403]:
                blocked_count += 1
                print(f"🛡️ 流 {i+1} 被阻止")
            else:
                print(f"✅ 流 {i+1} 通过 (状态码: {status_code})")
    
    print(f"📊 并发流攻击结果: {blocked_count}/{total_streams} 流被阻止")
    return blocked_count > total_streams * 0.3  # 超过30%被阻止认为防护有效

def test_security_api():
    """测试安全管理API"""
    print("🧪 测试安全管理API...")
    
    headers = {
        "Authorization": f"Bearer {API_KEY}",
        "Content-Type": "application/json"
    }
    
    try:
        # 测试获取安全统计
        response = requests.get(f"{BASE_URL}/api/security/stats", headers=headers)
        print(f"📊 安全统计API状态码: {response.status_code}")
        
        if response.status_code == 200:
            stats = response.json()
            print(f"📈 安全统计: {json.dumps(stats, indent=2, ensure_ascii=False)}")
            return True
        else:
            print("❌ 安全统计API访问失败")
            return False
            
    except Exception as e:
        print(f"❌ 安全API测试失败: {e}")
        return False

def main():
    """主测试函数"""
    print("🚀 开始Tea API安全防护测试")
    print("=" * 50)
    
    test_results = []
    
    # 执行各项测试
    test_results.append(("正常请求", test_normal_request()))
    test_results.append(("超长Prompt攻击防护", test_large_prompt_attack()))
    test_results.append(("高频请求攻击防护", test_high_frequency_attack()))
    test_results.append(("恶意流攻击防护", test_malicious_stream_attack()))
    test_results.append(("并发流攻击防护", test_concurrent_streams()))
    test_results.append(("安全管理API", test_security_api()))
    
    # 输出测试结果
    print("\n" + "=" * 50)
    print("📋 测试结果汇总:")
    print("=" * 50)
    
    passed = 0
    for test_name, result in test_results:
        status = "✅ 通过" if result else "❌ 失败"
        print(f"{test_name}: {status}")
        if result:
            passed += 1
    
    print(f"\n📊 总体结果: {passed}/{len(test_results)} 项测试通过")
    
    if passed >= len(test_results) * 0.8:
        print("🎉 安全防护系统工作正常！")
    else:
        print("⚠️ 安全防护系统可能存在问题，请检查配置")

if __name__ == "__main__":
    main()
