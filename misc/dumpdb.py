import sqlite3
import json
import codecs


def dict_factory(cursor, row):
    d = {}
    for idx, col in enumerate(cursor.description):
        d[col[0]] = row[idx]
    return d


connection = sqlite3.connect("lib.db")
connection.row_factory = dict_factory

cursor = connection.cursor()

#cursor.execute("select * from items where id <= 40000")
#cursor.execute("select * from items where id > 40000 and id <= 80000")
#cursor.execute("select * from items where id > 80000 and id <= 120000")
cursor.execute("select * from items where id > 120000")

results = cursor.fetchall()

for r in results:
    r["path"] = r["path"].decode("utf-8")

with codecs.open("items.json", "w", encoding="utf-8") as f:
    f.write('{"docs":')
    f.write(json.dumps(results, sort_keys = True, ensure_ascii=False))
    f.write("}")

connection.close()