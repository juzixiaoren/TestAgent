"""
集成测试 - 自顶向下
从 StudentManager 顶层开始，逐步替换为学生模块
"""
import sys
import os
sys.path.insert(0, os.path.join(os.path.dirname(__file__), '..'))

from student_score_manager import StudentManager, Student


# ========== 第1步：测试 StudentManager + Stub（用构造的Student对象代替） ==========

def step1_test_with_fake_students():
    """
    第1步：自顶向下第1层
    用 stub（手动构造 Student 对象）测试 StudentManager 的顶层方法
    """
    print("\n--- 步骤1：StudentManager + Stub（构造Student对象） ---")
    mgr = StudentManager()
    
    # 直接操作内部列表（模拟添加）
    stub_student = Student("STUB", "STUB001", 99.9)
    mgr._students.append(stub_student)
    mgr._hash_index["STUB001"] = stub_student
    
    # 测试顶层方法
    assert mgr.count == 1
    assert mgr.find("STUB001") is stub_student
    assert mgr.find_score("STUB001") == 99.9
    print("  ✅ 第1层集成：StudentManager + Stub 通过")


# ========== 第2步：替换stub为真实Student模块 ==========

def step2_integrate_real_student():
    """
    第2步：自顶向下第2层
    替换stub为真实Student类，通过add()方法添加
    """
    print("\n--- 步骤2：集成真实Student类 ---")
    mgr = StudentManager()
    
    # 使用真实 Student 类
    mgr.add("A01", "S001", 85.5)
    mgr.add("A01", "S002", 92.0)
    
    assert mgr.count == 2
    assert isinstance(mgr.find("S001"), Student)
    assert mgr.find("S001").class_id == "A01"
    assert mgr.find("S001").score == 85.5
    print("  ✅ 第2层集成：集成真实 Student 类通过")


# ========== 第3步：集成完整功能链 ==========

def step3_full_integration():
    """
    第3步：完整功能链集成测试
    增→查→改→排序→删→查的全流程
    """
    print("\n--- 步骤3：完整功能链集成 ---")
    mgr = StudentManager()
    
    # 添加多个学生
    mgr.add("A01", "S001", 60)
    mgr.add("A01", "S002", 90)
    mgr.add("A01", "S003", 75)
    assert mgr.count == 3
    
    # 哈希快查
    s = mgr.find("S002")
    assert s.score == 90
    
    # 更新分数
    mgr.update_score("S002", 95)
    assert mgr.find_score("S002") == 95
    
    # 排序（验证排序结果与哈希索引一致）
    sorted_list = mgr.sort_by_score(reverse=False)
    assert [s.score for s in sorted_list] == [60, 75, 95]
    assert mgr.find_score("S002") == 95  # 排序不应改变原数据
    
    # 删除
    mgr.remove("S001")
    assert mgr.count == 2
    assert mgr.find("S001") is None
    
    # 再次排序
    sorted_list2 = mgr.sort_by_score(reverse=True)
    assert [s.score for s in sorted_list2] == [95, 75]
    
    print("  ✅ 第3层集成：完整功能链通过")


def step4_integration_sort_edge_cases():
    """
    第4步：排序与数据一致性的集成测试
    """
    print("\n--- 步骤4：排序与数据一致性 ---")
    mgr = StudentManager()
    
    # 大量数据（测试递归深度和稳定性）
    import random
    students_data = [(f"A{i//10+1:02d}", f"S{i:04d}", random.randint(0, 100)) 
                     for i in range(50)]
    for cid, sid, score in students_data:
        mgr.add(cid, sid, score)
    
    assert mgr.count == 50
    
    # 升序排序
    sorted_asc = mgr.sort_by_score(reverse=False)
    for i in range(len(sorted_asc) - 1):
        assert sorted_asc[i].score <= sorted_asc[i+1].score, \
            f"升序失败: {sorted_asc[i].score} > {sorted_asc[i+1].score}"
    print("  ✅ 第4层集成：50个元素升序排序正确")
    
    # 降序排序
    sorted_desc = mgr.sort_by_score(reverse=True)
    for i in range(len(sorted_desc) - 1):
        assert sorted_desc[i].score >= sorted_desc[i+1].score, \
            f"降序失败: {sorted_desc[i].score} < {sorted_desc[i+1].score}"
    print("  ✅ 第4层集成：50个元素降序排序正确")
    
    # 验证原数据未被排序修改
    assert mgr.count == 50
    print("  ✅ 第4层集成：排序不修改原数据")


if __name__ == "__main__":
    print("=" * 60)
    print("集成测试 - 自顶向下")
    print("=" * 60)
    step1_test_with_fake_students()
    step2_integrate_real_student()
    step3_full_integration()
    step4_integration_sort_edge_cases()
    print("\n" + "=" * 60)
    print("自顶向下集成测试完成！")
    print("=" * 60)
