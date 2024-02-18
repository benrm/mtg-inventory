#!/bin/sh

if [ ! -e ./.git/hooks/pre-commit ] ; then
	ln -s ../../check.sh ./.git/hooks/pre-commit
fi

(cd ./golang/mtg-inventory && ./check.sh)
