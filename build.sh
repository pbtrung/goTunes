#!/bin/bash

go build --tags "fts5"
mv goTunes $GOPATH/bin
cp -r data $GOPATH/bin
cp .env $GOPATH/bin
cp -r public $GOPATH/bin
rm $GOPATH/bin/data/.gitkeep
