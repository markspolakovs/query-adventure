#!/usr/bin/env bash
set -eu

function find_cb_binary() {
  cmd=$1
  if command -v "$cmd" >/dev/null; then
      env "$cmd"
      return 0
  elif [ -d /opt/couchbase/bin ]; then
    echo "/opt/couchbase/bin/$cmd"
    return 0
  elif [ -d '/Applications/Couchbase Server.app/Contents/Resources/couchbase-core/bin' ]; then
    echo "/Applications/Couchbase Server.app/Contents/Resources/couchbase-core/bin/couchbase-cli/$cmd"
    return 0
  elif command -v docker >/dev/null; then
    echo "docker run --rm -it -v $(pwd):$(pwd) --network container:cb couchbase/server:7.1.1 $cmd"
    return 0
  fi
  echo "couchbase-cli not found" >&2
  return 1
}

cbc="$(find_cb_binary couchbase-cli)"
cbi="$(find_cb_binary cbimport)"

cbh=${COUCHBASE_HOST:-http://localhost:8091}
cbu=${COUCHBASE_USER:-Administrator}
cbp=${COUCHBASE_PASSWORD:-password}

function create_bucket() {
  if ! bc_out=$($cbc bucket-create -c "$cbh" -u "$cbu" -p "$cbp" --bucket "$1" --bucket-type couchbase --bucket-ramsize 512 --bucket-replica 1 --wait); then
    if echo "$bc_out" | grep -q 'already exists'; then
      echo "Bucket $1 already exists, skipping creation"
    else
      echo "Failed to create bucket $1" >&2
      echo "$bc_out" >&2
      exit 1
    fi
  fi
}

function create_collection() {
  if ! cc_out=$($cbc collection-manage -c "$cbh" -u "$cbu" -p "$cbp" --bucket tfgm --create-collection "_default.$1"); then
    if echo "$cc_out" | grep -q 'already exists'; then
      echo "Collection $1 already exists, skipping creation"
    else
      echo "Failed to create collection $1" >&2
      echo "$cc_out" >&2
      exit 1
    fi
  fi
}

mkdir -p _tmp

if [ ! -f ~/Downloads/TfGMgtfsnew.zip ]; then
  echo "Downloading TfGMgtfsnew.zip"
  curl -o ~/Downloads/TfGMgtfsnew.zip https://odata.tfgm.com/opendata/downloads/TfGMgtfsnew.zip
fi

mkdir -p _tmp/tfgm
unzip -d _tmp/tfgm ~/Downloads/TfGMgtfsnew.zip

echo "Importing TfGM data..."

create_bucket tfgm
create_collection agency
$cbi csv --infer-types -c "$cbh" -u "$cbu" -p "$cbp" -b tfgm --scope-collection-exp "_default.agency" -g "%agency_id%" -d "file://$(pwd)/_tmp/tfgm/agency.txt"
create_collection calendar_dates
$cbi csv --infer-types -c "$cbh" -u "$cbu" -p "$cbp" -b tfgm --scope-collection-exp "_default.calendar_dates" -g "%service_id%::%date%" -d "file://$(pwd)/_tmp/tfgm/calendar_dates.txt"
create_collection calendar
$cbi csv --infer-types -c "$cbh" -u "$cbu" -p "$cbp" -b tfgm --scope-collection-exp "_default.calendar" -g "%service_id%" -d "file://$(pwd)/_tmp/tfgm/calendar.txt"
create_collection routes
$cbi csv --infer-types -c "$cbh" -u "$cbu" -p "$cbp" -b tfgm --scope-collection-exp "_default.routes" -g "%route_id%" -d "file://$(pwd)/_tmp/tfgm/routes.txt"
create_collection stop_times
$cbi csv --infer-types -c "$cbh" -u "$cbu" -p "$cbp" -b tfgm --scope-collection-exp "_default.stop_times" -g "%trip_id%::%stop_id%::%stop_sequence%" -d "file://$(pwd)/_tmp/tfgm/stop_times.txt"
create_collection stops
$cbi csv --infer-types -c "$cbh" -u "$cbu" -p "$cbp" -b tfgm --scope-collection-exp "_default.stops" -g "%stop_id%" -d "file://$(pwd)/_tmp/tfgm/stops.txt"
create_collection trips
$cbi csv --infer-types -c "$cbh" -u "$cbu" -p "$cbp" -b tfgm --scope-collection-exp "_default.trips" -g "%trip_id%" -d "file://$(pwd)/_tmp/tfgm/trips.txt"

echo "Cleaning up..."
rm -rf _tmp
