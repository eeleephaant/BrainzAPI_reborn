import json
from abc import ABC, abstractmethod
from datetime import datetime, time


class LessonTimings:
    def __init__(self, lesson_time: dict):
        self.lesson_time = lesson_time

    def get_saturday_time(self, pair_num: int) -> tuple[str, str]:
        return self.lesson_time["saturday"]["start_time"][str(pair_num)], self.lesson_time["saturday"]["end_time"][
            str(pair_num)]

    def get_weekly_time(self, pair_num: int) -> tuple[str, str]:
        return self.lesson_time["weekly"]["start_time"][str(pair_num)], self.lesson_time["weekly"]["end_time"][
            str(pair_num)]

    @classmethod
    def from_json(cls, json_path: str):
        with open(json_path, 'r') as f:
            data = json.load(f)
        return cls(data)


def _parse_time(time_str: str) -> time:
    """Parses a time string in the format HH:MM into a datetime.time object."""
    return datetime.strptime(time_str, "%H:%M").time()


class Lesson:

    def __repr__(self):
        return (f"Lesson(date={self.date}, group={self.group}, lesson_num={self.lesson_num}, "
                f"cab_num={self.cab_num}, teacher={self.teacher}, name={self.name}, start_time={self.start_time}, "
                f"end_time={self.end_time})")

    def __init__(self, lesson_timings, date: datetime, group: str, lesson_num: int = 0,
                 cab_num: str = "-", teacher: str = "-", name: str = '',
                 institution_id: int = 1):
        self.date = date
        self.group = group
        self.lesson_timings = lesson_timings

        start_end_time = lesson_timings.get_saturday_time(lesson_num) if date.isoweekday() == 6 else \
            lesson_timings.get_weekly_time(lesson_num)

        start_time_str, end_time_str = start_end_time

        self.start_time = datetime.combine(date, _parse_time(start_time_str))
        self.end_time = datetime.combine(date, _parse_time(end_time_str))

        self.lesson_num = lesson_num
        self.name = name
        self.teacher = teacher
        self.cab_num = cab_num
        self.institution_id = institution_id


class AbstractScheduleParser(ABC):

    def __init__(self, path_to_file: str):
        self.path_to_file = path_to_file

    @abstractmethod
    def extract_date(self) -> datetime:
        pass

    @abstractmethod
    def extract_groups(self) -> list[str]:
        pass

    @abstractmethod
    def extract_lessons(self, lesson_timings: LessonTimings) -> list[Lesson]:
        pass
