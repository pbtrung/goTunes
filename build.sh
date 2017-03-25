#!/bin/bash

go install
cp -r data ../../bin
cp .env ../../bin
cp -r public ../../bin
rm ../../bin/data/.gitkeep
