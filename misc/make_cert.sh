#!/bin/sh

function make_cert() {
    openssl genrsa -out $1.key 4096
    openssl req -new -out $1.csr -key $1.key -config $1.cnf -utf8 \
        -nameopt utf8
    openssl x509 -req -sha256 -days 3650 -in $1.csr -signkey $1.key\
        -out $1.crt -extensions v3_req -extfile $1.cnf
    rm $1.csr
}

function usage() {
    echo "$0 <configfile>"
}

if [ $# -ne 1 ];
then
    usage
    exit 1
fi

make_cert ${1%.cnf}
