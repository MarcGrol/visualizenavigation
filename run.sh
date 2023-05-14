#!/bin/sh

go install

for i in 5 10 15 20 50 100
do
	analyzeNavigation -input-filename=ca_may_62_users.csv -limit=$i > ca_nav_top_$i.dot && dot -Tpng -o ca_nav_top_$i.png ca_nav_top_$i.dot
done
