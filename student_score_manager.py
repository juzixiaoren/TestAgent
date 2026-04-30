"""
学生成绩管理系统
- 内存存储：使用列表存储学生信息（班号、学号、分数）
- 快速排序：按分数排序
- 哈希快查：用学号作为 key 在 O(1) 时间内查分数
"""


class Student:
    """学生数据模型"""

    def __init__(self, class_id: str, student_id: str, score: float):
        self.class_id = class_id       # 班号
        self.student_id = student_id   # 学号
        self.score = score             # 分数

    def __repr__(self):
        return f"[{self.class_id}] {self.student_id} : {self.score}"


class StudentManager:
    """学生成绩管理器"""

    def __init__(self):
        # 主存储：学生列表
        self._students: list[Student] = []
        # 哈希索引：学号 -> Student（用于 O(1) 快速查找）
        self._hash_index: dict[str, Student] = {}

    # ───────────────────────────── 增删改 ─────────────────────────────

    def add(self, class_id: str, student_id: str, score: float) -> None:
        """添加一个学生"""
        student = Student(class_id, student_id, score)
        self._students.append(student)
        self._hash_index[student_id] = student

    def remove(self, student_id: str) -> bool:
        """按学号删除学生"""
        if student_id not in self._hash_index:
            return False
        target = self._hash_index.pop(student_id)
        self._students[:] = [s for s in self._students if s.student_id != student_id]
        return True

    def update_score(self, student_id: str, new_score: float) -> bool:
        """更新某个学生的分数"""
        student = self._hash_index.get(student_id)
        if student is None:
            return False
        student.score = new_score
        return True

    # ───────────────────────────── 哈希快查 ─────────────────────────────

    def find(self, student_id: str) -> Student | None:
        """用哈希索引 O(1) 查找学生信息"""
        return self._hash_index.get(student_id)

    def find_score(self, student_id: str) -> float | None:
        """用哈希索引快速查分数"""
        student = self._hash_index.get(student_id)
        return student.score if student else None

    # ───────────────────────────── 快速排序 ─────────────────────────────

    def sort_by_score(self, reverse: bool = False) -> list[Student]:
        """
        按分数快速排序（不修改原顺序，返回新列表）
        :param reverse: False = 升序,  True = 降序
        """
        arr = self._students[:]          # 拷贝一份
        self._quick_sort(arr, 0, len(arr) - 1, reverse)
        return arr

    def _quick_sort(self, arr: list[Student], low: int, high: int, reverse: bool) -> None:
        """快速排序（原地排序）"""
        if low >= high:
            return

        # 三数取中法选 pivot，避免最坏情况
        mid = (low + high) // 2
        if arr[low].score > arr[mid].score:
            arr[low], arr[mid] = arr[mid], arr[low]
        if arr[low].score > arr[high].score:
            arr[low], arr[high] = arr[high], arr[low]
        if arr[mid].score > arr[high].score:
            arr[mid], arr[high] = arr[high], arr[mid]
        # pivot 放到 high-1 位置
        arr[mid], arr[high] = arr[high], arr[mid]

        pivot_score = arr[high].score
        i = low - 1

        for j in range(low, high):
            if not reverse:          # 升序
                if arr[j].score <= pivot_score:
                    i += 1
                    arr[i], arr[j] = arr[j], arr[i]
            else:                    # 降序
                if arr[j].score >= pivot_score:
                    i += 1
                    arr[i], arr[j] = arr[j], arr[i]

        arr[i + 1], arr[high] = arr[high], arr[i + 1]
        pivot_idx = i + 1

        self._quick_sort(arr, low, pivot_idx - 1, reverse)
        self._quick_sort(arr, pivot_idx + 1, high, reverse)

    # ───────────────────────────── 辅助方法 ─────────────────────────────

    @property
    def count(self) -> int:
        """学生总数"""
        return len(self._students)

    def list_all(self) -> list[Student]:
        """返回所有学生列表（按插入顺序）"""
        return self._students[:]

    def __repr__(self) -> str:
        return f"StudentManager(count={self.count})"


# ───────────────────────────── 交互式演示 ─────────────────────────────

def demo():
    """演示程序（不填初始数据，由用户交互输入）"""
    manager = StudentManager()
    print("=" * 60)
    print("          学生成绩管理系统（快排 + 哈希快查）")
    print("=" * 60)
    print("可用命令：")
    print("  add   <班号> <学号> <分数>   添加学生")
    print("  find  <学号>                  哈希快查分数")
    print("  sort  [升序|降序]             快速排序并展示")
    print("  update <学号> <新分数>        更新成绩")
    print("  remove <学号>                 删除学生")
    print("  list                          列出所有学生")
    print("  count                         学生总数")
    print("  help                          显示帮助")
    print("  exit                          退出")
    print("=" * 60)

    while True:
        try:
            cmd_line = input("\n>>> ").strip()
            if not cmd_line:
                continue

            parts = cmd_line.split()
            cmd = parts[0].lower()

            if cmd == "exit":
                print("👋 再见！")
                break

            elif cmd == "add":
                if len(parts) != 4:
                    print("❌ 格式：add <班号> <学号> <分数>")
                    continue
                _, class_id, student_id, score_str = parts
                score = float(score_str)
                manager.add(class_id, student_id, score)
                print(f"✅ 已添加：{student_id} ({class_id}) = {score}")

            elif cmd == "find":
                if len(parts) != 2:
                    print("❌ 格式：find <学号>")
                    continue
                _, student_id = parts
                score = manager.find_score(student_id)
                if score is None:
                    print(f"❌ 未找到学号 {student_id}")
                else:
                    student = manager.find(student_id)
                    print(f"🔍 学号 {student_id} → 班号 {student.class_id}，分数 {score}")

            elif cmd == "sort":
                reverse = False
                if len(parts) >= 2:
                    if parts[1] in ("降序", "desc", "d"):
                        reverse = True
                result = manager.sort_by_score(reverse=reverse)
                if not result:
                    print("📭 暂无学生数据")
                else:
                    order = "降序" if reverse else "升序"
                    print(f"📊 按分数{order}排序结果：")
                    for i, s in enumerate(result, 1):
                        print(f"   {i:2d}. {s}")

            elif cmd == "update":
                if len(parts) != 3:
                    print("❌ 格式：update <学号> <新分数>")
                    continue
                _, student_id, score_str = parts
                new_score = float(score_str)
                if manager.update_score(student_id, new_score):
                    print(f"✅ 已更新 {student_id} 的分数为 {new_score}")
                else:
                    print(f"❌ 未找到学号 {student_id}")

            elif cmd == "remove":
                if len(parts) != 2:
                    print("❌ 格式：remove <学号>")
                    continue
                _, student_id = parts
                if manager.remove(student_id):
                    print(f"✅ 已删除学号 {student_id}")
                else:
                    print(f"❌ 未找到学号 {student_id}")

            elif cmd == "list":
                students = manager.list_all()
                if not students:
                    print("📭 暂无学生数据")
                else:
                    print(f"📋 当前学生列表（共 {len(students)} 人）：")
                    for i, s in enumerate(students, 1):
                        print(f"   {i:2d}. {s}")

            elif cmd == "count":
                print(f"👥 学生总数：{manager.count}")

            elif cmd == "help":
                print("可用命令：")
                print("  add   <班号> <学号> <分数>   添加学生")
                print("  find  <学号>                  哈希快查分数")
                print("  sort  [升序|降序]             快速排序并展示")
                print("  update <学号> <新分数>        更新成绩")
                print("  remove <学号>                 删除学生")
                print("  list                          列出所有学生")
                print("  count                         学生总数")
                print("  exit                          退出")
            else:
                print(f"❌ 未知命令：{cmd}，输入 help 查看帮助")

        except ValueError:
            print("❌ 分数必须是数字")
        except KeyboardInterrupt:
            print("\n👋 再见！")
            break


if __name__ == "__main__":
    demo()
