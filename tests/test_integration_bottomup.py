"""
集成测试 - 自底向上
从 Student 底层开始，逐步向上集成到 StudentManager
"""
import sys
import os
sys.path.insert(0, os.path.join(os.path.dirname(__file__), '..'))

from student_score_manager import Student, StudentManager


# ========== 第1步：测试底层模块 Student ==========

def step1_test_student():
    """
    第1步：从最底层 Student 模块开始测试
    编写 driver 直接创建和操作 Student 对象
    """
    print("\n--- 步骤1：底层 Student 模块 ---")
    
    # Driver：直接构造 Student
    s1 = Student("A01", "S001", 85.5)
    assert s1.class_id == "A01"
    assert s1.student_id == "S001"
    assert s1.score == 85.5
    assert repr(s1) == "[A01] S001 : 85.5"
    print("  ✅ Student 构造 + __repr__ 正确")
    
    # Driver：分数可修改（属性可变）
    s1.score = 95.0
    assert s1.score == 95.0
    print("  ✅ Student 分数属性可修改")
    
    # Driver：边界测试
    s2 = Student("", "", 0)
    assert s2.class_id == ""
    assert s2.student_id == ""
    assert s2.score == 0
    print("  ✅ Student 边界值：空字符串、零分")


# ========== 第2步：测试 StudentManager 的增删改（依赖 Student） ==========

def step2_test_crud_with_student():
    """
    第2步：向上集成到 StudentManager 的增删改方法
    验证 Student 与 StudentManager 之间的接口
    """
    print("\n--- 步骤2：StudentManager 增删改 + Student ---")
    mgr = StudentManager()
    
    # Driver 调用 add()
    mgr.add("A01", "S001", 85.5)
    assert mgr.count == 1
    
    # 验证内部存储的 Student 对象
    assert isinstance(mgr._students[0], Student)
    assert isinstance(mgr._hash_index["S001"], Student)
    assert mgr._students[0] is mgr._hash_index["S001"]
    print("  ✅ add() 正确创建 Student 并同步到列表和哈希索引")
    
    # 测试 update_score
    mgr.update_score("S001", 90.0)
    assert mgr._students[0].score == 90.0
    assert mgr._hash_index["S001"].score == 90.0
    print("  ✅ update_score() 同步更新 Student 对象（引用一致）")
    
    # 测试 remove
    mgr.remove("S001")
    assert mgr.count == 0
    assert "S001" not in mgr._hash_index
    print("  ✅ remove() 正确删除 Student 并从哈希索引移除")


# ========== 第3步：测试哈希快查模块（依赖 StudentManager 内部状态） ==========

def step3_test_hash_search():
    """
    第3步：集成哈希快查方法
    """
    print("\n--- 步骤3：哈希快查与 Student 集成 ---")
    mgr = StudentManager()
    mgr.add("A01", "S001", 85.5)
    mgr.add("A01", "S002", 92.0)
    
    # find 返回 Student 对象
    s = mgr.find("S001")
    assert s is not None
    assert isinstance(s, Student)
    assert s.score == 85.5
    assert s.class_id == "A01"
    print("  ✅ find() 返回完整的 Student 对象")
    
    # find_score 返回分数
    score = mgr.find_score("S002")
    assert score == 92.0
    print("  ✅ find_score() 正确返回分数")
    
    # 找不到时返回 None
    assert mgr.find("NONEXIST") is None
    assert mgr.find_score("NONEXIST") is None
    print("  ✅ 查找不存在的学号返回 None")


# ========== 第4步：完整集成 - 排序模块 ==========

def step4_test_sort_full_integration():
    """
    第4步：完整集成 - 排序模块与其他模块的协作
    """
    print("\n--- 步骤4：排序模块完整集成 ---")
    mgr = StudentManager()
    
    # 通过增删改准备数据
    mgr.add("A01", "S001", 70)
    mgr.add("A01", "S002", 85)
    mgr.add("A01", "S003", 60)
    
    # 排序
    sorted_list = mgr.sort_by_score(reverse=False)
    
    # 验证排序结果中的 Student 对象完整性
    assert all(isinstance(s, Student) for s in sorted_list)
    assert sorted_list[0].student_id == "S003"  # 60分
    assert sorted_list[1].student_id == "S001"  # 70分
    assert sorted_list[2].student_id == "S002"  # 85分
    print("  ✅ 排序结果 Student 对象完整")
    
    # 验证排序后更新仍然正确关联
    mgr.update_score("S003", 90)
    assert mgr.find_score("S003") == 90
    mgr.add("A01", "S004", 55)
    assert mgr.count == 4
    
    # 重新排序
    sorted_list2 = mgr.sort_by_score(reverse=True)
    assert sorted_list2[0].student_id == "S003"  # 90分
    print("  ✅ 排序与增删改集成：修改后重新排序正确")
    
    # 大规模集成测试
    import random
    mgr2 = StudentManager()
    # 添加100个学生
    for i in range(100):
        mgr2.add(f"A{i//20+1:02d}", f"S{i:04d}", random.randint(0, 100))
    assert mgr2.count == 100
    
    sorted_asc = mgr2.sort_by_score(reverse=False)
    sorted_desc = mgr2.sort_by_score(reverse=True)
    
    # 验证排序正确性
    for i in range(99):
        assert sorted_asc[i].score <= sorted_asc[i+1].score
        assert sorted_desc[i].score >= sorted_desc[i+1].score
    print("  ✅ 大规模集成（100个学生）：升序+降序全部正确")


if __name__ == "__main__":
    print("=" * 60)
    print("集成测试 - 自底向上")
    print("=" * 60)
    step1_test_student()
    step2_test_crud_with_student()
    step3_test_hash_search()
    step4_test_sort_full_integration()
    print("\n" + "=" * 60)
    print("自底向上集成测试完成！")
    print("=" * 60)
