#!/bin/sh

gcc -fPIC -DPIC -shared -rdynamic -o pam_remote2.so module/pam_remote2.c
mv pam_remote2.so /usr/lib/x86_64-linux-gnu/security/pam_remote2.so
chown root:root /usr/lib/x86_64-linux-gnu/security/pam_remote2.so
chmod 755 /usr/lib/x86_64-linux-gnu/security/pam_remote2.so
