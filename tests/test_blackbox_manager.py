"""
黑盒测试 - StudentManager 类
测试方法：等价类划分 + 边界值分析 + 判定表
"""
import sys
import os
sys.path.insert(0, os.path.join(os.path.dirname(__file__), '..'))

from student_score_manager import StudentManager, Student


def test_add():
    """等价类划分 + 边界值 - add() 方法"""
    print("\n--- test_add ---")
    mgr = StudentManager()

    # 有效等价类：正常添加
    mgr.add("A01", "S001", 85.5)
    assert mgr.count == 1, "添加后 count 应为 1"
    print("  ✅ 有效等价类：正常添加学生")

    # 边界值：分数为0
    mgr.add("A01", "S002", 0)
    assert mgr.count == 2, "添加后 count 应为 2"
    print("  ✅ 边界值：分数为0")

    # 边界值：负分数
    mgr.add("A01", "S003", -10)
    assert mgr.count == 3
    print("  ✅ 边界值：负分数（语法上允许）")

    # 边界值：超大分数
    mgr.add("A01", "S004", 1e6)
    assert mgr.count == 4
    print("  ✅ 边界值：超大分数")

    # 判定表：重复学号添加
    mgr.add("A01", "S001", 90.0)  # 重复学号
    assert mgr.count == 5, "重复学号应该被允许添加（未做去重）"
    # 但哈希索引会被覆盖
    found = mgr.find("S001")
    assert found.score == 90.0, "重复学号添加后，哈希索引应指向最新记录"
    print("  ✅ 判定表：重复学号添加（哈希索引被覆盖）")


def test_remove():
    """等价类划分 - remove() 方法"""
    print("\n--- test_remove ---")
    mgr = StudentManager()
    mgr.add("A01", "S001", 85)
    mgr.add("A01", "S002", 90)

    # 有效等价类：删除存在的学号
    result = mgr.remove("S001")
    assert result == True, "删除存在的学号应返回 True"
    assert mgr.count == 1, "删除后 count 应为 1"
    assert mgr.find("S001") is None, "删除后哈希索引也应移除"
    print("  ✅ 有效等价类：删除存在的学号")

    # 无效等价类：删除不存在的学号
    result = mgr.remove("NONEXIST")
    assert result == False, "删除不存在的学号应返回 False"
    assert mgr.count == 1, "count 应不变"
    print("  ✅ 无效等价类：删除不存在的学号")

    # 边界值：空学号
    result = mgr.remove("")
    assert result == False, "删除空学号应返回 False"
    print("  ✅ 边界值：空学号删除")


def test_update_score():
    """等价类划分 - update_score() 方法"""
    print("\n--- test_update_score ---")
    mgr = StudentManager()
    mgr.add("A01", "S001", 85)

    # 有效等价类：更新存在学生
    result = mgr.update_score("S001", 95.5)
    assert result == True, "更新存在的学生应返回 True"
    assert mgr.find_score("S001") == 95.5, "分数应更新为 95.5"
    print("  ✅ 有效等价类：更新存在学生的分数")

    # 无效等价类：更新不存在学生
    result = mgr.update_score("NONEXIST", 80)
    assert result == False, "更新不存在的学生应返回 False"
    print("  ✅ 无效等价类：更新不存在学生")

    # 边界值：更新分数为负数
    result = mgr.update_score("S001", -20)
    assert result == True
    assert mgr.find_score("S001") == -20
    print("  ✅ 边界值：更新分数为负数")

    # 边界值：更新分数为0
    mgr.update_score("S001", 0)
    assert mgr.find_score("S001") == 0
    print("  ✅ 边界值：更新分数为0")


def test_find_and_find_score():
    """等价类划分 - find() 和 find_score() 方法"""
    print("\n--- test_find_and_find_score ---")
    mgr = StudentManager()
    mgr.add("A01", "S001", 85.5)

    # 有效等价类：查找存在学号
    s = mgr.find("S001")
    assert s is not None
    assert s.student_id == "S001"
    assert s.score == 85.5
    print("  ✅ 有效等价类：find() 查找存在的学号")

    score = mgr.find_score("S001")
    assert score == 85.5
    print("  ✅ 有效等价类：find_score() 查找存在的学号")

    # 无效等价类：查找不存在学号
    s = mgr.find("NONEXIST")
    assert s is None
    print("  ✅ 无效等价类：find() 查找不存在的学号")

    score = mgr.find_score("NONEXIST")
    assert score is None
    print("  ✅ 无效等价类：find_score() 查找不存在的学号")

    # 边界值：空学号
    s = mgr.find("")
    assert s is None
    print("  ✅ 边界值：find() 查找空学号")


def test_sort_by_score():
    """判定表 + 边界值 - sort_by_score() 方法"""
    print("\n--- test_sort_by_score ---")

    # 判定表 Case 1: 空列表排序
    mgr = StudentManager()
    result = mgr.sort_by_score()
    assert result == [], "空列表排序应返回空列表"
    print("  ✅ 判定表1：空列表排序")

    # 判定表 Case 2: 单元素
    mgr.add("A01", "S001", 85)
    result = mgr.sort_by_score()
    assert len(result) == 1 and result[0].score == 85
    print("  ✅ 判定表2：单元素排序")

    # 判定表 Case 3: 多元素升序
    mgr2 = StudentManager()
    mgr2.add("A01", "S003", 90)
    mgr2.add("A01", "S001", 60)
    mgr2.add("A01", "S002", 75)
    result_asc = mgr2.sort_by_score(reverse=False)
    scores_asc = [s.score for s in result_asc]
    assert scores_asc == [60, 75, 90], f"升序应为 [60, 75, 90], 实际 {scores_asc}"
    print("  ✅ 判定表3：多元素升序排序")

    # 判定表 Case 4: 多元素降序
    result_desc = mgr2.sort_by_score(reverse=True)
    scores_desc = [s.score for s in result_desc]
    assert scores_desc == [90, 75, 60], f"降序应为 [90, 75, 60], 实际 {scores_desc}"
    print("  ✅ 判定表4：多元素降序排序")

    # 边界值 Case 5: 分数相同的元素
    mgr3 = StudentManager()
    mgr3.add("A01", "S001", 80)
    mgr3.add("A01", "S002", 80)
    result = mgr3.sort_by_score()
    assert len(result) == 2
    assert result[0].score == 80 and result[1].score == 80
    print("  ✅ 边界值：相同分数排序")

    # 边界值 Case 6: 含负数分数
    mgr4 = StudentManager()
    mgr4.add("A01", "S001", -10)
    mgr4.add("A01", "S002", 0)
    mgr4.add("A01", "S003", 50)
    result = mgr4.sort_by_score()
    scores = [s.score for s in result]
    assert scores == [-10, 0, 50], f"含负数升序应为 [-10, 0, 50], 实际 {scores}"
    print("  ✅ 边界值：含负数分数排序")


def test_list_all():
    """测试 list_all() 返回原顺序"""
    print("\n--- test_list_all ---")
    mgr = StudentManager()
    mgr.add("A01", "S001", 85)
    mgr.add("A01", "S002", 90)
    mgr.add("A01", "S003", 75)
    students = mgr.list_all()
    assert len(students) == 3
    assert students[0].student_id == "S001"
    assert students[1].student_id == "S002"
    assert students[2].student_id == "S003"
    print("  ✅ list_all() 保持插入顺序")


def test_count_property():
    """测试 count 属性"""
    print("\n--- test_count ---")
    mgr = StudentManager()
    assert mgr.count == 0, "空管理器 count 应为 0"
    mgr.add("A01", "S001", 85)
    assert mgr.count == 1
    mgr.remove("S001")
    assert mgr.count == 0
    print("  ✅ count 属性正确")


if __name__ == "__main__":
    print("=" * 60)
    print("黑盒测试 - StudentManager 类")
    print("=" * 60)
    test_add()
    test_remove()
    test_update_score()
    test_find_and_find_score()
    test_sort_by_score()
    test_list_all()
    test_count_property()
    print("\n" + "=" * 60)
    print("黑盒测试全部完成！")
    print("=" * 60)
