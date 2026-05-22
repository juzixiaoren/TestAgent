"""
白盒测试 - StudentManager._quick_sort 内部逻辑
测试方法：条件组合覆盖 + 路径覆盖
"""
import sys
import os
sys.path.insert(0, os.path.join(os.path.dirname(__file__), '..'))

from student_score_manager import StudentManager, Student


def test_path_empty_list():
    """路径1：空列表 → 直接触发 low >= high → return"""
    print("\n--- 路径1：空列表 (low >= high) ---")
    mgr = StudentManager()
    result = mgr.sort_by_score()
    assert result == [], "空列表应返回 []"
    print("  ✅ 路径覆盖：空列表排序")


def test_path_single_element():
    """路径2：单元素 → low >= high → return（不进入排序循环）"""
    print("\n--- 路径2：单元素 ---")
    mgr = StudentManager()
    mgr.add("A01", "S001", 85)
    result = mgr.sort_by_score()
    assert len(result) == 1 and result[0].score == 85
    print("  ✅ 路径覆盖：单元素排序")


def test_path_two_elements_asc():
    """路径3：两个元素升序（覆盖三数取中和分区逻辑）"""
    print("\n--- 路径3：两个元素升序 ---")
    mgr = StudentManager()
    mgr.add("A01", "S002", 90)
    mgr.add("A01", "S001", 60)
    result = mgr.sort_by_score(reverse=False)
    assert [s.score for s in result] == [60, 90], f"升序应为 [60, 90], 实际 {[s.score for s in result]}"
    print("  ✅ 路径覆盖：两元素升序")


def test_path_two_elements_desc():
    """路径4：两个元素降序"""
    print("\n--- 路径4：两个元素降序 ---")
    mgr = StudentManager()
    mgr.add("A01", "S001", 60)
    mgr.add("A01", "S002", 90)
    result = mgr.sort_by_score(reverse=True)
    assert [s.score for s in result] == [90, 60], f"降序应为 [90, 60], 实际 {[s.score for s in result]}"
    print("  ✅ 路径覆盖：两元素降序")


def test_condition_combination_all_same():
    """条件组合覆盖：所有分数相同 → pivot比较全走等号分支"""
    print("\n--- 条件组合：所有分数相同 ---")
    mgr = StudentManager()
    mgr.add("A01", "S001", 80)
    mgr.add("A01", "S002", 80)
    mgr.add("A01", "S003", 80)
    mgr.add("A01", "S004", 80)
    # 升序
    result = mgr.sort_by_score(reverse=False)
    assert all(s.score == 80 for s in result), "所有分数应为 80"
    print("  ✅ 条件组合：全部相同分数（升序）")
    # 降序
    result = mgr.sort_by_score(reverse=True)
    assert all(s.score == 80 for s in result), "所有分数应为 80"
    print("  ✅ 条件组合：全部相同分数（降序）")


def test_condition_combination_boundary():
    """条件组合覆盖：选取边界条件组合"""
    print("\n--- 条件组合：边界条件 ---")
    
    # 组合1：所有条件为真（降序，所有 >= pivot）
    mgr1 = StudentManager()
    mgr1.add("A01", "S001", 50)
    mgr1.add("A01", "S002", 70)
    mgr1.add("A01", "S003", 90)
    result = mgr1.sort_by_score(reverse=True)
    assert [s.score for s in result] == [90, 70, 50]
    print("  ✅ 条件组合1：降序（条件 not reverse=False，走 else 分支）")

    # 组合2：升序（条件 not reverse=True，走 if 分支）
    result = mgr1.sort_by_score(reverse=False)
    assert [s.score for s in result] == [50, 70, 90]
    print("  ✅ 条件组合2：升序（条件 not reverse=True，走 if 分支）")

    # 组合3：逆序排列输入（测试分区逻辑）
    mgr2 = StudentManager()
    mgr2.add("A01", "S001", 90)
    mgr2.add("A01", "S002", 70)
    mgr2.add("A01", "S003", 50)
    result = mgr2.sort_by_score(reverse=False)
    assert [s.score for s in result] == [50, 70, 90]
    print("  ✅ 条件组合3：逆序输入升序排序")

    # 组合4：已排序输入（测试分区不交换）
    mgr3 = StudentManager()
    mgr3.add("A01", "S001", 50)
    mgr3.add("A01", "S002", 70)
    mgr3.add("A01", "S003", 90)
    result = mgr3.sort_by_score(reverse=False)
    assert [s.score for s in result] == [50, 70, 90]
    print("  ✅ 条件组合4：已排序输入升序")


def test_path_multi_level():
    """路径覆盖：多层递归排序"""
    print("\n--- 路径：多层递归 ---")
    mgr = StudentManager()
    # 5个元素，确保多级递归
    scores = [85, 60, 90, 75, 95]
    for i, s in enumerate(scores):
        mgr.add("A01", f"S00{i+1}", s)
    
    # 升序
    result = mgr.sort_by_score(reverse=False)
    assert [s.score for s in result] == [60, 75, 85, 90, 95]
    print("  ✅ 路径覆盖：5元素升序（多层递归）")

    # 降序
    result = mgr.sort_by_score(reverse=True)
    assert [s.score for s in result] == [95, 90, 85, 75, 60]
    print("  ✅ 路径覆盖：5元素降序（多层递归）")


def test_remove_integrity():
    """白盒：删除后哈希索引与列表一致性"""
    print("\n--- 内部状态一致性 ---")
    mgr = StudentManager()
    mgr.add("A01", "S001", 85)
    mgr.add("A01", "S002", 90)
    mgr.add("A01", "S003", 75)
    
    # 删除中间元素
    mgr.remove("S002")
    # 验证内部 _hash_index 和 _students 一致
    remaining = mgr.list_all()
    assert len(remaining) == 2
    assert remaining[0].student_id == "S001"
    assert remaining[1].student_id == "S003"
    assert mgr._hash_index.get("S002") is None
    print("  ✅ 内部状态一致性：删除后哈希与列表同步")


def test_update_integrity():
    """白盒：更新分数后排序正确性"""
    print("\n--- 更新后排序 ---")
    mgr = StudentManager()
    mgr.add("A01", "S001", 85)
    mgr.add("A01", "S002", 90)
    mgr.add("A01", "S003", 75)
    
    mgr.update_score("S001", 100)
    result = mgr.sort_by_score(reverse=False)
    assert [s.score for s in result] == [75, 90, 100]
    print("  ✅ 更新后排序：更新分数后排序正确")


if __name__ == "__main__":
    print("=" * 60)
    print("白盒测试 - 条件组合覆盖 + 路径覆盖")
    print("=" * 60)
    test_path_empty_list()
    test_path_single_element()
    test_path_two_elements_asc()
    test_path_two_elements_desc()
    test_condition_combination_all_same()
    test_condition_combination_boundary()
    test_path_multi_level()
    test_remove_integrity()
    test_update_integrity()
    print("\n" + "=" * 60)
    print("白盒测试全部完成！")
    print("=" * 60)
