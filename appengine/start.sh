#!/bin/sh

exec dev_appserver.py --admin_host=0.0.0.0 --enable_host_checking=false --host=0.0.0.0 app.yaml
