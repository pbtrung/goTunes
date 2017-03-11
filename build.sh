#!/bin/bash

go install
cp -r data $GOPATH/bin
cp .env $GOPATH/bin
cp .aws $GOPATH/bin
cp -r public $GOPATH/bin
rm $GOPATH/bin/data/.gitkeep
