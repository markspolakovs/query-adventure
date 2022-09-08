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
    echo "/Applications/Couchbase Server.app/Contents/Resources/couchbase-core/bin/$cmd"
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
  if [[ "$cbh" == *"cloud.couchbase.com" ]]; then
    echo "Not creating bucket against Capella."
    return 0
  fi
  name=$1
  quota=${2:-512}
  backend=${3:-couchstore}
  if ! bc_out=$("$cbc" bucket-create -c "$cbh" -u "$cbu" -p "$cbp" --bucket "$name" --bucket-type couchbase --storage-backend "$backend" --bucket-ramsize "$quota" --bucket-replica 1 --wait); then
    if echo "$bc_out" | grep -q 'already exists'; then
      echo "Bucket $name already exists, skipping creation"
    else
      echo "Failed to create bucket $name" >&2
      echo "$bc_out" >&2
      exit 1
    fi
  fi
}

function create_collection() {
  bucket=$1
  collection=$2
  if ! cc_out=$("$cbc" collection-manage --cacert ~/Downloads/capella.pem -c "$cbh" -u "$cbu" -p "$cbp" --bucket "$bucket" --create-collection "_default.$collection"); then
    if echo "$cc_out" | grep -q 'already exists'; then
      echo "Collection $bucket._default.$collection already exists, skipping creation"
    else
      echo "Failed to create collection $1" >&2
      echo "$cc_out" >&2
      exit 1
    fi
  fi
}

mkdir -p _tmp

if [ ! -d "_tmp/f1" ]; then
  echo "Please download the F1 dataset from TK and unzip it into _tmp/f1"
  exit 1
fi

go run ./scripts/process_f1.go
create_bucket f1 200 couchstore
"$cbi" json --cacert ~/Downloads/capella.pem -c "$cbh" -u "$cbu" -p "$cbp" -f list -d file://races.json -b f1 --generate-key "%raceId%"

if [ ! -f ~/Downloads/TfGMgtfsnew.zip ]; then
  echo "Downloading TfGMgtfsnew.zip"
  curl -o ~/Downloads/TfGMgtfsnew.zip https://odata.tfgm.com/opendata/downloads/TfGMgtfsnew.zip
fi

mkdir -p _tmp/tfgm
unzip -d _tmp/tfgm ~/Downloads/TfGMgtfsnew.zip

echo "Importing TfGM data..."

create_bucket tfgm 1024 magma
create_collection tfgm agency
"$cbi" csv  --cacert ~/Downloads/capella.pem --infer-types -c "$cbh" -u "$cbu" -p "$cbp" -b tfgm --scope-collection-exp "_default.agency" -g "%agency_id%" -d "file://$(pwd)/_tmp/tfgm/agency.txt"
create_collection tfgm calendar_dates
"$cbi" csv  --cacert ~/Downloads/capella.pem --infer-types -c "$cbh" -u "$cbu" -p "$cbp" -b tfgm --scope-collection-exp "_default.calendar_dates" -g "%service_id%::%date%" -d "file://$(pwd)/_tmp/tfgm/calendar_dates.txt"
create_collection tfgm calendar
"$cbi" csv  --cacert ~/Downloads/capella.pem --infer-types -c "$cbh" -u "$cbu" -p "$cbp" -b tfgm --scope-collection-exp "_default.calendar" -g "%service_id%" -d "file://$(pwd)/_tmp/tfgm/calendar.txt"
create_collection tfgm routes
"$cbi" csv  --cacert ~/Downloads/capella.pem --infer-types -c "$cbh" -u "$cbu" -p "$cbp" -b tfgm --scope-collection-exp "_default.routes" -g "%route_id%" -d "file://$(pwd)/_tmp/tfgm/routes.txt"
create_collection tfgm stop_times
"$cbi" csv  --cacert ~/Downloads/capella.pem --infer-types -c "$cbh" -u "$cbu" -p "$cbp" -b tfgm --scope-collection-exp "_default.stop_times" -g "%trip_id%::%stop_id%::%stop_sequence%" -d "file://$(pwd)/_tmp/tfgm/stop_times.txt"
create_collection tfgm stops
"$cbi" csv  --cacert ~/Downloads/capella.pem --infer-types -c "$cbh" -u "$cbu" -p "$cbp" -b tfgm --scope-collection-exp "_default.stops" -g "%stop_id%" -d "file://$(pwd)/_tmp/tfgm/stops.txt"
create_collection tfgm trips
"$cbi" csv  --cacert ~/Downloads/capella.pem --infer-types -c "$cbh" -u "$cbu" -p "$cbp" -b tfgm --scope-collection-exp "_default.trips" -g "%trip_id%" -d "file://$(pwd)/_tmp/tfgm/trips.txt"

if [ -n "${CLEANUP:-}" ]; then
  echo "Cleaning up..."
  rm -rf _tmp
fi
