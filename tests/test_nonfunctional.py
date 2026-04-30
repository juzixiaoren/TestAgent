"""
系统测试 - 非功能需求测试（性能 + 健壮性）
"""
import sys
import os
sys.path.insert(0, os.path.join(os.path.dirname(__file__), '..'))

from student_score_manager import StudentManager
import time

def test_performance_large_dataset():
    """性能测试：大数据集下的操作耗时"""
    print("\n--- 性能测试：大数据集 ---")
    mgr = StudentManager()
    
    # 测试添加性能
    n = 1000
    start = time.time()
    for i in range(n):
        mgr.add(f"A{i//100+1:02d}", f"S{i:06d}", i % 100)
    add_time = time.time() - start
    print(f"  ✅ 添加 {n} 个学生耗时: {add_time:.3f}秒 ({add_time/n*1000:.1f}ms/个)")
    
    # 测试哈希查找性能 (O(1))
    start = time.time()
    for i in range(n):
        s = mgr.find(f"S{i:06d}")
        assert s is not None
    find_time = time.time() - start
    print(f"  ✅ 哈希查找 {n} 次耗时: {find_time:.3f}秒 ({find_time/n*1000:.1f}ms/次)")
    
    # 测试排序性能 (O(n log n))
    start = time.time()
    sorted_list = mgr.sort_by_score(reverse=False)
    sort_time = time.time() - start
    print(f"  ✅ 排序 {n} 个学生耗时: {sort_time:.3f}秒")
    assert len(sorted_list) == n
    
    # 验证排序正确
    for i in range(n - 1):
        assert sorted_list[i].score <= sorted_list[i+1].score
    print(f"  ✅ 排序结果正确性验证通过")
    
    # 测试删除性能
    start = time.time()
    for i in range(n):
        mgr.remove(f"S{i:06d}")
    remove_time = time.time() - start
    print(f"  ✅ 删除 {n} 个学生耗时: {remove_time:.3f}秒")
    assert mgr.count == 0


def test_performance_worst_case_sort():
    """性能测试：最坏情况排序（已排序和逆序）"""
    print("\n--- 性能测试：最坏情况排序 ---")
    
    # 已排序（最坏情况测试三数取中法的效果）
    mgr = StudentManager()
    for i in range(500):
        mgr.add("A01", f"S{i:04d}", i)
    
    start = time.time()
    mgr.sort_by_score(reverse=False)
    t1 = time.time() - start
    print(f"  ✅ 已排序数据（500个）排序耗时: {t1:.3f}秒")
    
    # 逆序
    start = time.time()
    mgr.sort_by_score(reverse=True)
    t2 = time.time() - start
    print(f"  ✅ 逆序数据（500个）排序耗时: {t2:.3f}秒")
    
    # 全部相同分数
    mgr2 = StudentManager()
    for i in range(500):
        mgr2.add("A01", f"S{i:04d}", 50)
    
    start = time.time()
    mgr2.sort_by_score(reverse=False)
    t3 = time.time() - start
    print(f"  ✅ 同分数数据（500个）排序耗时: {t3:.3f}秒")


def test_robustness_extreme_values():
    """健壮性测试：极端值处理"""
    print("\n--- 健壮性测试：极端值 ---")
    mgr = StudentManager()
    
    # 极大学号
    mgr.add("A01", "X" * 1000, 85)
    assert mgr.find("X" * 1000) is not None
    print("  ✅ 极长学号（1000字符）处理正常")
    
    # Unicode/特殊字符
    mgr.add("A01", "测试-学号-①", 90)
    assert mgr.find("测试-学号-①") is not None
    print("  ✅ Unicode学号处理正常")
    
    # 极大极小分数
    mgr.add("A01", "MAX", 1e10)
    mgr.add("A01", "MIN", -1e10)
    assert mgr.find_score("MAX") == 1e10
    assert mgr.find_score("MIN") == -1e10
    print("  ✅ 极大/极小分数处理正常")
    
    # 排序含极端值
    result = mgr.sort_by_score(reverse=False)
    assert result[0].score == -1e10
    assert result[-1].score == 1e10
    print("  ✅ 含极端值排序正确")


if __name__ == "__main__":
    print("=" * 60)
    print("系统测试 - 非功能需求")
    print("=" * 60)
    test_performance_large_dataset()
    test_performance_worst_case_sort()
    test_robustness_extreme_values()
    print("\n" + "=" * 60)
    print("非功能测试完成！")
    print("=" * 60)
