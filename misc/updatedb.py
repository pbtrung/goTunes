#!/usr/bin/python

import sqlite3
import sys


def addTable(db, md5List, gdList):
    conn = sqlite3.connect(db)
    c = conn.cursor()

    stmt = "CREATE TABLE md5 (md5 TEXT, path BLOB)"
    c.execute(stmt)
    with open(md5List) as md5File:
        for md5Line in md5File:
            fileMD5 = md5Line[:33].strip()
            filePath = md5Line[33:].strip().encode("utf-8")

            t = (fileMD5, filePath)
            stmt = "INSERT INTO md5 (md5, path) VALUES (?, ?)"
            c.execute(stmt, t)

    conn.commit()

    stmt = "CREATE TABLE gdr (fileId TEXT, md5 TEXT)"
    c.execute(stmt)
    with open(gdList) as gdFile:
        for gdLine in gdFile:
            gdFileId = gdLine.split(' ')[0].strip()
            gdFileMD5 = gdLine.split(' ')[1].strip()

            t = (gdFileId, gdFileMD5)
            stmt = "INSERT INTO gdr (fileId, md5) VALUES (?, ?)"
            c.execute(stmt, t)

    conn.commit()
    conn.close()


def updatedb(db, md5List, gdList):

    addTable(db, md5List, gdList)

    conn = sqlite3.connect(db)
    c = conn.cursor()

    print("Add COLUMN to items")
    stmt = "ALTER TABLE items ADD COLUMN artpath BLOB"
    c.execute(stmt)
    conn.commit()
    stmt = "ALTER TABLE items ADD COLUMN md5 TEXT"
    c.execute(stmt)
    conn.commit()
    stmt = "ALTER TABLE items ADD COLUMN fileId TEXT"
    c.execute(stmt)
    conn.commit()

    print("Add COLUMN to albums")
    stmt = "ALTER TABLE albums ADD COLUMN md5 TEXT"
    c.execute(stmt)
    conn.commit()
    stmt = "ALTER TABLE albums ADD COLUMN fileId TEXT"
    c.execute(stmt)
    conn.commit()

    print("Add MD5 to albums")
    stmt = "UPDATE albums SET md5 = (SELECT md5 FROM md5 WHERE path = albums.artpath)"
    c.execute(stmt)
    conn.commit()

    print("Add fileId to albums")
    stmt = "UPDATE albums SET fileId = (SELECT fileId FROM gdr WHERE md5 = albums.md5)"
    c.execute(stmt)
    conn.commit()

    print("Add MD5 to items")
    stmt = "UPDATE items SET md5 = (SELECT md5 FROM md5 WHERE path = items.path)"
    c.execute(stmt)
    conn.commit()

    print("Add fileId to items")
    stmt = "UPDATE items SET fileId = (SELECT fileId FROM gdr WHERE md5 = items.md5)"
    c.execute(stmt)
    conn.commit()

    print("Add artpath to items")
    stmt = "UPDATE items SET artpath = (SELECT fileId FROM albums WHERE id = items.album_id)"
    c.execute(stmt)
    conn.commit()

    conn.close()


if __name__ == '__main__':
    db = sys.argv[1]
    md5List = sys.argv[2]
    gdList = sys.argv[3]
    updatedb(db, md5List, gdList)
