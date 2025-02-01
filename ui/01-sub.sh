#!/usr/bin/env ash

cat << EOF > /etc/nginx/conf.d/01-sub_filter.conf
sub_filter_types application/javascript;
sub_filter 'currency:"EUR"' 'currency:"${CURRENCY_ISO:-EUR}"';
sub_filter_once off;
EOF
