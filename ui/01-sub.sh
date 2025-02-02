#!/usr/bin/env ash

if [ ! -z "$LOCAL_API_PORT" ]; then
  export LOCAL_API_URL="http://localhost:${LOCAL_API_PORT}/"
fi

cat << EOF > /etc/nginx/conf.d/01-sub_filter.conf
sub_filter_types application/javascript;
sub_filter 'currency:"EUR"' 'currency:"${CURRENCY_ISO:-EUR}"';
sub_filter 'fetch("api/' 'fetch("${LOCAL_API_URL}api/';
sub_filter_once off;
EOF
