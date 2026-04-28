import psycopg2  # type: ignore

from parser_app import utils
from parser_app.models import Lesson


def get_group_ids(conn, institution_id: int, group_names: set[str]) -> dict[str, int]:
    cursor = conn.cursor()
    cursor.execute(
        """
		SELECT name, id FROM groups
		WHERE institution_id = %s AND LOWER(name) = ANY(%s)
		""",
        (institution_id, list(group_names))
    )
    return {row[0].lower(): row[1] for row in cursor.fetchall()}


def write_lessons_to_bd(lessons: list[Lesson]) -> None:
    with psycopg2.connect(
            dsn=utils.DSN
    ) as conn:
        cursor = conn.cursor()

        group_names = {lesson.group.lower() for lesson in lessons}
        group_ids = get_group_ids(conn, lessons[0].institution_id, group_names)

        valid_lessons = [
            lesson for lesson in lessons if lesson.group.lower() in group_ids
        ]

        cursor.executemany(
            """
            INSERT INTO lessons (
                start_time, end_time, num, name, teacher_name, cab_num, group_id, institution_id
            )
            VALUES (%s, %s, %s, %s, %s, %s, %s, %s)
            ON CONFLICT (start_time, group_id, num, name, teacher_name)
            DO NOTHING
            """,
            [
                (
                    lesson.start_time,
                    lesson.end_time,
                    lesson.lesson_num,
                    lesson.name,
                    lesson.teacher,
                    lesson.cab_num,
                    group_ids[lesson.group.lower()],
                    lesson.institution_id,
                )
                for lesson in valid_lessons
            ]
        )
