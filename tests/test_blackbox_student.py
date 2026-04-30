"""
黑盒测试 - Student 类
测试方法：等价类划分 + 边界值分析
"""
import sys
import os
sys.path.insert(0, os.path.join(os.path.dirname(__file__), '..'))

from student_score_manager import Student

def test_student_init():
    """等价类划分测试 - Student 构造"""
    test_cases = [
        # (class_id, student_id, score, 描述)
        ("A01", "S001", 85.5, "有效等价类：正常字符串+正分数"),
        ("B02", "S002", 0, "边界值：分数为0"),
        ("C03", "S003", 100, "边界值：分数为100（满分）"),
        ("", "S004", 59.9, "边界值：空班号字符串"),
        ("D04", "", 60, "边界值：空学号字符串"),
        ("E05", "S005", -10.5, "无效等价类：负分数"),
        ("F06", "S006", 999.99, "边界值：超大分数"),
    ]
    
    for class_id, student_id, score, desc in test_cases:
        try:
            s = Student(class_id, student_id, score)
            assert s.class_id == class_id, f"{desc}: class_id 不匹配"
            assert s.student_id == student_id, f"{desc}: student_id 不匹配"
            assert s.score == score, f"{desc}: score 不匹配"
            print(f"  ✅ {desc}")
        except Exception as e:
            print(f"  ❌ {desc}: {e}")

def test_student_repr():
    """测试 __repr__ 输出格式"""
    s = Student("A01", "S001", 85.5)
    expected = "[A01] S001 : 85.5"
    result = repr(s)
    assert result == expected, f"__repr__ 期望 '{expected}', 实际 '{result}'"
    print(f"  ✅ __repr__ 格式正确: {result}")

if __name__ == "__main__":
    print("=" * 60)
    print("黑盒测试 - Student 类")
    print("=" * 60)
    test_student_init()
    test_student_repr()
