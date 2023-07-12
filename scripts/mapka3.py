import datetime
import json
import random
from typing import Optional, List
from dataclasses import dataclass
from dataclasses_json import dataclass_json

CLUSTER_MAX_DELTA_MIN = 2
ACTION_MAX_DELTA_SEC = 120
ACTION_MAX_COUNT = 15

ERROR_PROB = 0.1
USER_PROB = 0.8
SESSION_PROB = 0.8

ACTIONS = [
    ("Marker", "/markers", "POST"),
    ("Polygon", "/polygons", "POST"),
    ("Point", "/point", "POST"),
    ("SelectObce", "/select", "POST"),
]

@dataclass_json
@dataclass
class User:
    ip: Optional[str] = None
    session: Optional[str] = None
    id: Optional[str] = None


def create_query_line(datetime: datetime.datetime, user: User):
    action = random.choice(ACTIONS)
    return {
        "message": action[0],
        "context": {},
        "level": 200,
        "level_name": "INFO",
        "channel": "access",
        "datetime": datetime.isoformat() + "+01:00",
        "extra": {
            "user_id": None if user.id is None else str(user.id),
            "session_selector": user.session,
            "ip": user.ip,
            "url": action[1],
            "http_method": action[2],
            "server": "localhost",
            "referrer": "http://localhost:8083/",
        }
    }

def create_error_line(datetime: datetime.datetime, user: User):
    action = random.choice(ACTIONS)
    return {
        "message": action[0],
        "context": {},
        "level": 200,
        "level_name": "ERROR",
        "channel": "access",
        "datetime": datetime.isoformat() + "+01:00",
        "extra": {
            "user_id": None if user.id is None else str(user.id),
            "session_selector": user.session,
            "ip": user.ip,
            "url": action[1],
            "http_method": action[2],
            "server": "localhost",
            "referrer": "http://localhost:8083/",
        }
    }

def generate_users(num: int) -> List[User]:
    ans = []
    for i in range(num):
        u = User()
        if random.random() <= USER_PROB:
            u.id = random.randint(0, 1000)
        if u.id:
            u.session = ''.join(random.choices('0123456789abcdef', k=12))
        elif random.random() <= SESSION_PROB:
            u.session = ''.join(random.choices('0123456789abcdef', k=12))
        u.ip = '.'.join(str(random.randint(0, 255)) for _ in range(4))
        ans.append(u)
    return ans


if __name__ == "__main__":
    entries = []
    today = datetime.date.today()
    entry_time = datetime.datetime.combine(today, datetime.time(0, 0, 0))
    users = generate_users(20)
    i = 0
    while entry_time.date() == today:
        entry_time = entry_time + datetime.timedelta(
            minutes=random.randint(1, CLUSTER_MAX_DELTA_MIN), microseconds=random.randint(0, 999999))
        ux = random.choice(users)
        for _ in range(random.randint(1, ACTION_MAX_COUNT)):
            entry_time = entry_time + datetime.timedelta(
                seconds=-ACTION_MAX_DELTA_SEC + random.randint(1, 2*ACTION_MAX_DELTA_SEC),
                microseconds=random.randint(0, 999999))
            if random.random() <= ERROR_PROB:
                entries.append(create_error_line(entry_time, ux))
            else:
                entries.append(create_query_line(entry_time, ux))
            i += 1
        #if i > 1000:

    for entry in sorted(entries, key=lambda x: x['datetime']):
        print(json.dumps(entry))
