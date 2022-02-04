# Copyright 2018 Tomas Machalek <tomas.machalek@gmail.com>
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

"""
A module for writing SkE user interaction log to a Redis queue

How to use (CNC-specific stuff) -

Please note that this script expects a modified version of
run.cgi (with provided database connection).

-----------------------------
import redislog

# ...

if __name__ == '__main__':
    t1 = time.time()
    # ... orig code ...
    redislog.log_action(conn, time.time() - t1)
    conn.close()
-----------------------------
"""

import json
import redis
import os
import datetime
import urllib
import urlparse


KLOGPROC_CONF_PATH = '/home/tomas/work/go/src/klogproc/conf.json'
DEFAULT_DATETIME_FORMAT = '%Y-%m-%d %H:%M:%S'
ERROR_LOG_PATH = '/var/log/klogproc/ske_errors.txt'

if KLOGPROC_CONF_PATH:
    with open(KLOGPROC_CONF_PATH) as fr:
        data = json.load(fr)
        rc = data['logRedis']
        REDIS_SERVER, REDIS_PORT = rc['address'].split(':')
        REDIS_PORT = int(REDIS_PORT)
        REDIS_DB = rc['database']
        REDIS_QUEUE_KEY = rc['queueKey']
else:
    REDIS_SERVER = '127.0.0.1'
    REDIS_PORT = 6379
    REDIS_DB = 1
    REDIS_QUEUE_KEY = 'ske_log_queue'


class QueryValues(object):

    def __init__(self, url):
        self._action = None
        self._args = {}
        if url:
            action, tmp = urllib.splitquery(url)
            self._args = urlparse.parse_qs(tmp if tmp else '')
            if 'run.cgi/' in action:
                _, self._action = action.rsplit('/', 1)

    @property
    def action(self):
        return self._action

    @property
    def args(self):
        ans = {}
        for k, v in self._args.items():
            if len(v) == 1:
                ans[k] = v[0]
            elif len(v) > 1:
                ans[k] = v
        return ans


def get_env(s):
    return os.environ.get(s, None)


def find_user_id(conn, username):
    cur = conn.cursor()
    cur.execute('SELECT id FROM user WHERE user = %s', (username,))
    ans = cur.fetchone()
    return ans[0] if ans else None


def store_log_to_redis(rec):
    conn = redis.StrictRedis(host=REDIS_SERVER, port=REDIS_PORT, db=REDIS_DB)
    conn.rpush(REDIS_QUEUE_KEY, json.dumps(rec))


def create_log_record(mysql_conn, proc_time, log_date):
    log_data = {}
    log_data['user_id'] = find_user_id(mysql_conn, get_env('REMOTE_USER'))
    log_data['proc_time'] = round(proc_time, 3)
    log_data['settings'] = {}
    log_data['date'] = log_date
    log_data['request'] = {
        'HTTP_USER_AGENT': get_env('HTTP_USER_AGENT'),
        'REMOTE_ADDR': get_env('REMOTE_ADDR')
    }
    qv = QueryValues(get_env('REQUEST_URI'))
    log_data['params'] = qv.args
    log_data['action'] = qv.action
    return log_data


def log_action(mysql_conn, proc_time):
    log_date = datetime.datetime.today().strftime('%s.%%f' % DEFAULT_DATETIME_FORMAT)
    try:
        data = create_log_record(mysql_conn, proc_time, log_date)
        store_log_to_redis(data)
    except Exception as ex:
        if ERROR_LOG_PATH:
            with open(ERROR_LOG_PATH, 'a') as fw:
                fw.write('{0} ERROR: {1}'.format(log_date, ex))
