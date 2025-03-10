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

if [ -n "$LOCAL_API_PROTOCOL" ]; then
  cat << EOF >> "$CONFIG_FILE"
sub_filter 'window.location.protocol.replace(":","")' '"${LOCAL_API_PROTOCOL}"';
EOF
fi

if [ -n "$LOCAL_API_HOSTNAME" ]; then
  cat << EOF >> "$CONFIG_FILE"
sub_filter 'window.location.hostname.toLowerCase()' '"${LOCAL_API_HOSTNAME}"';
EOF
fi

if [ -n "$LOCAL_API_PORT" ]; then
  cat << EOF >> "$CONFIG_FILE"
sub_filter 'window.location.port.trim()' '"${LOCAL_API_PORT}"';
EOF
fi

if [ "$USE_SINGLE_COLUMN" = true ]; then
  cat << EOF >> "$CONFIG_FILE"
sub_filter 'md:grid-cols-2 ' '';
EOF
fi
