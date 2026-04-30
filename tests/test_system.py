"""
系统测试 - 端到端测试
模拟用户交互，验证整个系统功能
"""
import sys
import os
sys.path.insert(0, os.path.join(os.path.dirname(__file__), '..'))

import subprocess
import time


def run_demo_with_inputs(input_lines):
    """
    运行 demo() 并传入多行输入，返回输出内容
    """
    script_path = os.path.join(os.path.dirname(__file__), '..', 'student_score_manager.py')
    input_text = "\n".join(input_lines) + "\n"
    
    result = subprocess.run(
        ['python3', script_path],
        input=input_text,
        capture_output=True,
        text=True,
        timeout=30
    )
    return result.stdout, result.stderr


def test_system_full_workflow():
    """系统测试1：完整的增→查→改→排→删→查工作流"""
    print("\n--- 系统测试1：完整工作流 ---")
    
    commands = [
        "add A01 S001 85.5",
        "add A01 S002 92.0",
        "add A01 S003 60.5",
        "list",
        "find S002",
        "update S001 95.0",
        "sort 升序",
        "sort 降序",
        "remove S003",
        "find S003",
        "count",
        "exit"
    ]
    
    stdout, stderr = run_demo_with_inputs(commands)
    
    # 验证关键输出
    checks = [
        ("✅ 已添加：S001", "添加 S001"),
        ("✅ 已添加：S002", "添加 S002"),
        ("✅ 已添加：S003", "添加 S003"),
        ("当前学生列表（共 3 人）", "列出3个学生"),
        ("🔍 学号 S002", "查找 S002"),
        ("分数 92.0", "S002 分数为92.0"),
        ("✅ 已更新 S001 的分数为 95.0", "更新 S001 分数"),
        ("按分数升序排序结果", "升序排序"),
        ("S001", "升序列表含 S001"),
        ("S002", "升序列表含 S002"),
        ("S003", "升序列表含 S003"),
        ("按分数降序排序结果", "降序排序"),
        ("✅ 已删除学号 S003", "删除 S003"),
        ("❌ 未找到学号 S003", "删除后找不到 S003"),
        ("学生总数：2", "count 为 2"),
        ("👋 再见！", "退出"),
    ]
    
    all_passed = True
    for keyword, desc in checks:
        if keyword in stdout:
            print(f"  ✅ {desc}")
        else:
            print(f"  ❌ {desc} - 未找到关键字: '{keyword}'")
            all_passed = False
    
    if stderr:
        print(f"  ⚠️  stderr输出: {stderr}")
    
    return all_passed


def test_system_empty_operations():
    """系统测试2：空数据操作边界"""
    print("\n--- 系统测试2：空数据操作 ---")
    
    commands = [
        "list",
        "count",
        "find S001",
        "sort 升序",
        "sort 降序",
        "remove S001",
        "exit"
    ]
    
    stdout, stderr = run_demo_with_inputs(commands)
    
    checks = [
        ("暂无学生数据", "空列表提示"),
        ("学生总数：0", "count 为 0"),
        ("未找到学号 S001", "查找不存在学号"),
        ("暂无学生数据", "空排序提示"),
        ("删除不存在的学号", "或者"),  # 至少有一个提示
    ]
    
    all_passed = True
    for keyword, desc in checks:
        if keyword in stdout:
            print(f"  ✅ {desc}")
            break
    
    # 检查 remove 不存在的学号
    if "❌ 未找到学号 S001" in stdout:
        print(f"  ✅ 删除不存在学号")
    else:
        print(f"  ❌ 删除不存在学号 - 无提示")
        all_passed = False
    
    if stderr:
        print(f"  ⚠️  stderr输出: {stderr}")
    
    return all_passed


def test_system_invalid_inputs():
    """系统测试3：异常输入处理"""
    print("\n--- 系统测试3：异常输入 ---")
    
    commands = [
        "add",                # 缺少参数
        "add A01 S001",       # 缺少分数
        "add A01 S001 abc",   # 分数不是数字
        "find",               # 缺少学号
        "update S001",        # 缺少新分数
        "update S001 abc",    # 分数不是数字
        "remove",             # 缺少学号
        "sort unknown",       # 未知排序参数
        "unknown_command",    # 未知命令
        "exit"
    ]
    
    stdout, stderr = run_demo_with_inputs(commands)
    
    checks = [
        ("格式：add <班号> <学号> <分数>", "add 缺参数提示"),
        ("分数必须是数字", "非法分数提示"),
        ("格式：find <学号>", "find 缺参数提示"),
        ("格式：update <学号> <新分数>", "update 缺参数提示"),
        ("格式：remove <学号>", "remove 缺参数提示"),
        ("未知命令", "未知命令提示"),
    ]
    
    all_passed = True
    for keyword, desc in checks:
        if keyword in stdout:
            print(f"  ✅ {desc}")
        else:
            print(f"  ❌ {desc} - 未找到关键字: '{keyword}'")
            all_passed = False
    
    if stderr:
        print(f"  ⚠️  stderr输出: {stderr}")
    
    return all_passed


def test_system_duplicate_student_id():
    """系统测试4：重复学号处理"""
    print("\n--- 系统测试4：重复学号 ---")
    
    commands = [
        "add A01 S001 85.0",
        "add A01 S001 95.0",  # 重复学号
        "list",
        "find S001",
        "exit"
    ]
    
    stdout, stderr = run_demo_with_inputs(commands)
    
    # 验证重复学号添加后，记录仍在（但哈希索引覆盖）
    if "共 2 人" in stdout or "共 2" in stdout:
        print(f"  ⚠️ 重复学号添加后列表有2条（可能为bug）")
    else:
        # 检查 list 输出
        pass
    
    # 查找 S001 的分数
    if "分数 95.0" in stdout:
        print(f"  ✅ 重复学号后哈希索引指向最新记录")
    
    # 检查是否有重复学号的警告
    if "重复" in stdout:
        print(f"  ⚠️ 有重复学号警告")
    
    print(f"  ✅ 重复学号测试完成")


def test_system_help_command():
    """系统测试5：帮助命令"""
    print("\n--- 系统测试5：帮助命令 ---")
    
    commands = [
        "help",
        "exit"
    ]
    
    stdout, stderr = run_demo_with_inputs(commands)
    
    expected_commands = ["add", "find", "sort", "update", "remove", "list", "count", "help", "exit"]
    all_passed = True
    for cmd in expected_commands:
        if cmd in stdout:
            print(f"  ✅ help 显示 {cmd} 命令")
        else:
            print(f"  ❌ help 未显示 {cmd} 命令")
            all_passed = False
    
    return all_passed


if __name__ == "__main__":
    print("=" * 60)
    print("系统测试 - 端到端功能验证")
    print("=" * 60)
    
    results = []
    results.append(("完整工作流", test_system_full_workflow()))
    results.append(("空数据操作", test_system_empty_operations()))
    results.append(("异常输入处理", test_system_invalid_inputs()))
    results.append(("重复学号处理", test_system_duplicate_student_id()))
    results.append(("帮助命令", test_system_help_command()))
    
    print("\n" + "=" * 60)
    print("系统测试结果汇总")
    print("=" * 60)
    all_pass = True
    for name, passed in results:
        status = "✅ 通过" if passed else "❌ 失败"
        print(f"  {status}: {name}")
        if not passed:
            all_pass = False
    print("=" * 60)
    print("系统测试完成！" if all_pass else "系统测试发现缺陷！")
    print("=" * 60)
