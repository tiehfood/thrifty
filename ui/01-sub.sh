#!/usr/bin/env ash

CONFIG_FILE=/etc/nginx/conf.d/01-sub_filter.conf

cat << EOF > "$CONFIG_FILE"
sub_filter_types application/javascript;
sub_filter_once off;
EOF

if [ -n "$CURRENCY_ISO" ]; then
  cat << EOF >> "$CONFIG_FILE"
sub_filter 'currency:"EUR"' 'currency:"${CURRENCY_ISO}"';
EOF
fi

if [ -n "$LOCAL_API_PORT" ]; then
  cat << EOF >> "$CONFIG_FILE"
sub_filter 'fetch("api/' 'fetch("http://localhost:${LOCAL_API_PORT}/api/';
EOF
fi

if [ "$USE_SINGLE_COLUMN" = true ]; then
  cat << EOF >> "$CONFIG_FILE"
sub_filter 'md:grid-cols-2 ' '';
EOF
fi
