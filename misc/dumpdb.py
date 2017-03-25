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

cursor.execute("select COUNT(*) as total_rows from items")
total_rows = cursor.fetchall()[0]["total_rows"]

num = 1
inc = 30000
upper_row = -1
lower_row = -1
while upper_row < total_rows:
    lower_row = upper_row
    upper_row += inc

    if upper_row > total_rows:
        upper_row = total_rows

    params = (lower_row, upper_row)
    print("lower: " + str(lower_row) + ", upper: " + str(upper_row))
    cursor.execute("select * from items where id > ? and id <= ?", params)
    results = cursor.fetchall()

    with codecs.open("items" + str(num) + ".json", "w", encoding="utf-8") as f:
        f.write('{"docs":')
        f.write(json.dumps(results, sort_keys=True, ensure_ascii=False))
        f.write("}")

    num += 1


connection.close()