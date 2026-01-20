#!/bin/sh
# Wrapper script for nginx in foreground mode
# Required because supervizio's strings.Fields() doesn't handle quoted args
exec nginx -g 'daemon off;'
