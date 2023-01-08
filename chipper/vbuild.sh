#!/bin/bash

source vbuild.env

/usr/local/go/bin/go build -tags nolibopusfile $1
