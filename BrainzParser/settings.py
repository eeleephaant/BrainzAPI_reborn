import os
import json

settings_file = "settings.json"


def get_db_dir() -> str:
    with open(settings_file, 'r') as file:
        return json.load(file)["paths"]["db_dir"]


def get_vkapi_version() -> str:
    with open(settings_file, 'r') as file:
        return json.load(file)["settings"]["vkapi_version"]


def get_db_file_path() -> str:
    with open(settings_file, 'r') as file:
        return os.path.join(json.load(file)["paths"]["db_dir"], "main.db")


def get_logs_dir() -> str:
    with open(settings_file, 'r') as file:
        return json.load(file)["paths"]["logs_dir"]


def get_vk_token() -> str:
    with open(settings_file, 'r') as file:
        parsed_data = json.load(file)
        return parsed_data["settings"]["vk_token"]


def get_timings_file_path() -> str:
    with open(settings_file, 'r') as file:
        parsed_data = json.load(file)
        return parsed_data["paths"]["timings_json"]


def get_temp_path() -> str:
    with open(settings_file, 'r') as file:
        parsed_data = json.load(file)
        return parsed_data["paths"]["temp_dir"]


def get_link_teacher_file() -> str:
    with open(settings_file, 'r') as file:
        parsed_data = json.load(file)
        return parsed_data["paths"]["link_to_teacher"]


def get_actual_version_mobile_app() -> str:
    with open(settings_file, 'r') as file:
        parsed_data = json.load(file)
        return parsed_data["settings"]["actual_mobile_app_version"]

def get_db_connect_string() -> str:
    with open(settings_file, 'r') as file:
        parsed_data = json.load(file)
        name = parsed_data["database"]["name"]
        host = parsed_data["database"]["host"]
        user = parsed_data["database"]["user"]
        password = parsed_data["database"]["password"]
        port = parsed_data["database"]["port"]
        return f"dbname={name} user={user} password={password} host={host} port={port}"