import sys
import json
from collections import defaultdict
import matplotlib.pyplot as plt
import numpy as np
import csv
import sqlite3
import re

def fetch_date(d):
    date = d['date'].split('T')[0]
    y, m, d = date.split('-')
    return int(y), int(m), int(d)


def monthly_visits(rec, ans):
    y, m, _ = fetch_date(rec)
    ans[(y, m)] += 1

def num2month(v):
    return ('Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec')[v-1]

def display_monthly_visits(data, syn_data):
    fig, ax = plt.subplots()
    index = np.arange(len(data))
    bar_width = 0.35
    opacity = 0.4
    print(syn_data)
    items = sorted(data.items(), key=lambda x: x[0][0] * 100 + x[0][1])
    syn_items = sorted(syn_data.items(), key=lambda x: x[0][0] * 100 + x[0][1])
    fig.suptitle('Monthly visits (total: {0})'.format(sum(v[1] for v in items)), fontsize=14, fontweight='bold')
    rects1 = ax.bar(index, [v[1] for v in items], bar_width,
                alpha=opacity, color='b', label='Num. of queries (IC)')
    rects2 = ax.bar(index + bar_width, [v[1] for v in syn_items], bar_width,
                alpha=opacity, color='r', label='Num. of queries (SYN*)')
    ax.set_xticklabels(['{0} {1}'.format(num2month(v[0][1]), v[0][0]) for v in items])
    ax.legend()
    ax.set_xticks(index + bar_width / 2)
    ax.set_xlabel('Month')
    fig.tight_layout()
    plt.show()

def get_corp_lang(c):
    return c.split('_')[-1]

def total_used_corpora(rec, ans):
    cn = get_corp_lang(rec['params']['corpname'])
    ac = rec['alignedCorpora']
    if len(ac) == 0:
        ans[cn][None] += 1
    else:
        for v in ac:
            ans[cn][get_corp_lang(v)] += 1

def filter_corpora_table(data):
    rowsums = defaultdict(lambda: 0)
    colsums = defaultdict(lambda: 0)
    for c1, row in data.items():
        for c2, cell in row.items():
            rowsums[c1] += cell
            colsums[c2] += cell
    s1 = set(k for (k, v) in rowsums.items() if v > 100)
    s2 = set(k for (k, v) in colsums.items() if v > 100)
    return s1.intersection(s2).union(set([None]))


def display_total_used_corpora(data, out_path):
    allc = filter_corpora_table(data)
    print(allc)
    labels = sorted(allc)
    labels2 = labels[1:]
    tdata = np.zeros((len(labels2), len(labels)))
    lbmap = dict((v[1], v[0]) for v in enumerate(labels))
    lbmap2 = dict((v[1], v[0]) for v in enumerate(labels2))
    for b1 in labels2:
        for b2 in labels:
            tdata[lbmap2[b1], lbmap[b2]] = data[b1][b2]

    fig, ax = plt.subplots()
    im = ax.imshow(tdata)
    ax.set_xticks(np.arange(len(labels)))
    ax.set_yticks(np.arange(len(labels2)))
    ax.set_xticklabels(labels)
    ax.set_yticklabels(labels2)

    with open(out_path, 'wb') as fw:
        cw = csv.writer(fw, delimiter=';', quotechar='"', quoting=csv.QUOTE_ALL)
        cw.writerow(labels)

        for i in range(len(labels2)):
            for j in range(len(labels)):
                text = ax.text(j, i, int(tdata[i, j]), ha="center", va="center", color="w")
            cw.writerow([labels2[i]] + list(tdata[i]))
    fig.tight_layout()
    plt.show()

def parse_corp_version(cname):
    items = cname.split('_')
    if len(items) == 2:
        return 'v6'
    else:
        return items[1]

def used_versions(rec, ans):
    cn = rec['params']['corpname']
    ac = rec['alignedCorpora']
    mainv = parse_corp_version(cn)
    ans[mainv] += 1
    for v in ac:
        av = parse_corp_version(v)
        ans[av] += 1

def display_used_versions(data):
    fig, ax = plt.subplots()
    index = np.arange(len(data))
    bar_width = 0.3
    opacity = 0.4

    items = sorted(data.items(), key=lambda x: int(x[0][1:]))
    ax.bar(index, [v[1] for v in items], bar_width,
           alpha=opacity, color='b', label='Num. of queries')
    ax.set_xticklabels([v[0] for v in items])

    ax.legend()
    ax.set_xticks(index + bar_width / 2)
    ax.set_xlabel('InterCorp version')
    fig.tight_layout()
    plt.show()

def analyze_subc(db, user_id, subc):
    cur = db.cursor()
    cur.execute("select cql FROM subc_archive where subcname = ? AND user_id = ? AND corpname LIKE 'intercorp%'",
                (subc, user_id))
    row = cur.fetchone()
    if row:
        cql = row[0]
        if 'group' in cql:
            m = re.findall(r'group="([^"]+)"', cql)
            if m:
                return m
    return []

def used_div_groups(rec, div_groups, expression_lens, db):
    args = rec.get('params', {})

    if 'sca_div.group' in args:
        div_groups[args['sca_div.group']] += 1
    elif args.get('usesubcorp', None):
        #div_groups['__subc__'] += 1
        #subcorpora[args['usesubcorp']] += 1
        sc_list = analyze_subc(db, rec.get('user_id', None), args['usesubcorp'])
        if len(sc_list) > 0:
            expression_lens.append(len(sc_list))
            for sc in sc_list:
                div_groups['SUBC_{0}'.format(sc)] += 1
        else:
            div_groups['none'] += 1
    else:
        div_groups['none'] += 1

def display_subcorpora(data):
    print(data)

def display_div_groups(data):
    print(data)

def rec_match(rec):
    date = rec['date'].split('T')[0]
    y, m, _ = date.split('-')
    y, m = int(y), int(m)
    return y == 2017 and m >= 9 or y == 2018 and m < 9

def proc_file(fpath, syn_data, csv_path, subc_db):
    daily_v = defaultdict(lambda: 0)
    used_corpora = defaultdict(lambda: defaultdict(lambda: 0))
    versions = defaultdict(lambda: 0)
    div_groups = defaultdict(lambda: 0)
    expression_lens = []
    with open(fpath, 'rb') as fr:
        for line in fr:
            rec = json.loads(line)
            if rec_match(rec):
                #monthly_visits(rec, daily_v)
                #total_used_corpora(rec, used_corpora)
                #used_versions(rec, versions)
                used_div_groups(rec, div_groups, expression_lens, subc_db)
    #display_monthly_visits(daily_v, syn_data)
    #display_total_used_corpora(used_corpora, csv_path)
    #display_used_versions(versions)
    display_div_groups(div_groups)
    print('avg len: {0}'.format(float(sum(expression_lens) / float(len(expression_lens)))))


def proc_syn_file(fpath):
    monthly_v = defaultdict(lambda: 0)
    with open(fpath, 'rb') as fr:
        for line in fr:
            rec = json.loads(line)
            if rec_match(rec):
                monthly_visits(rec, monthly_v)
    return monthly_v


if __name__ == '__main__':
    with open(sys.argv[1], 'rb') as fr:
        conf = json.load(fr)
        syn = dict() #proc_syn_file(sys.argv[2])
        subc_db = sqlite3.connect(conf['subclogFile'])
        proc_file(conf['iclogs'], syn, conf['corpmapFile'], subc_db)