#!/bin/sh

wwctl node add -G 10.100.100.1 -H 00:6e:6f:64:65:31 --ipaddr 10.100.100.10 --ipmiaddr 10.100.100.254 --ipmipass password --ipmiport 6231 --ipmiuser admin --ipmiinterface lanplus --netdev eth0 --netmask 255.255.255.0 --nettagadd="DNS1=10.100.100.1" n1
wwctl node add -G 10.100.100.1 -H 00:6e:6f:64:65:32 --ipaddr 10.100.100.11 --ipmiaddr 10.100.100.254 --ipmipass password --ipmiport 6232 --ipmiuser admin --ipmiinterface lanplus --netdev eth0 --netmask 255.255.255.0 --nettagadd="DNS1=10.100.100.1" n2
