#!/bin/sh

(cd embed_assets/;set -e;go build;./embed_assets)
