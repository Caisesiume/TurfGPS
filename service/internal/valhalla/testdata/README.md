# Test extract — central Göteborg

`goteborg-inom-vallgraven.osm.pbf` is a small OpenStreetMap extract that CI
builds Valhalla's tiles from, so that the Valhalla adapter's integration tests
(`../valhalla_integration_test.go`) run against a real engine. It is committed
by the Owner's decision on #6 (`#issuecomment-6334699894`, 22 October 2026). The
full extract of `Architecture.md § D5` is a different thing, and arrives with the
server (`docs/adr/ADR-0006-deployment-topology.md`).

## Licence and attribution

**© OpenStreetMap contributors.** The data is available under the
[Open Database License (ODbL) 1.0](https://opendatacommons.org/licenses/odbl/1-0/);
see [openstreetmap.org/copyright](https://www.openstreetmap.org/copyright).

This file is a derivative database of OpenStreetMap. It is licensed under the
ODbL, and **not** under this repository's `LICENSE`, which covers the code. Anything
derived from it is subject to the ODbL's share-alike terms.

## What it holds

| | |
|---|---|
| Area | 57.6980–57.7120 N, 11.9500–11.9850 E — Inom Vallgraven and the streets around it, about 1.6 km by 2.1 km (3.2 km²) |
| Content | every way tagged `highway`, every `type=restriction` relation, and the nodes and member ways they reference, which reach slightly past the box |
| Size | 193,605 bytes; SHA-256 `0b878c989152be85784eea44c01e67d363140937629f27813caa0a3ac4f9aab9` |
| Data as of | `osm_base` 2026-10-22T07:05:28Z, as the Overpass API reported it |
| Retrieved | 22 October 2026, from `https://overpass-api.de/api/interpreter` |
| Query | `goteborg-inom-vallgraven.overpassql`, beside this file |

Only the street network is kept because it is all Valhalla builds a routing
graph from; buildings, land use and the rest would add size and nothing a route
can use.

**The car-restricted street** is Kungsgatan from the end of its residential part,
near Kaserntorget (57.7036 N, 11.9611 E), to Östra Larmgatan (57.7052 N,
11.9714 E): `highway=pedestrian` on ways 222674401, 304998012 and 304998017,
about 600 m long, with a street a car may use at each end. The
integration test plans a car leg between those two ends and checks that it does
not travel along the pedestrian street, after first checking that a walking
route between the same ends does — so the test proves its own premise against
whatever extract it is given.

## How CI uses it

The `valhalla` job in `.github/workflows/gates.yml` copies the file into a
scratch directory, builds tiles from it with the pinned Valhalla image's
`valhalla_build_config` and `valhalla_build_tiles`, serves them with
`valhalla_service` on loopback, and runs the integration tests against it. The
tiles are built without an admin or a time-zone database: this extract carries
no administrative boundaries, and nothing tested here depends on either.

## Regenerating it

Regenerating changes the data — OpenStreetMap is edited daily — so the new file
has a new date and hash, and both rows above are updated with it. Run from this
directory:

```bash
curl -sS -A 'TurfGPS-test-extract/1.0 (https://github.com/Caisesiume/TurfGPS)' \
  --data-urlencode data@goteborg-inom-vallgraven.overpassql \
  https://overpass-api.de/api/interpreter -o goteborg-inom-vallgraven.osm
grep -m1 osm_base goteborg-inom-vallgraven.osm      # the "Data as of" row
```

The Overpass API refuses a request without a `User-Agent` of its own (it answered
`406` to curl's default on 22 October 2026). Then convert the XML to PBF. The
committed file was converted with pyosmium 4.3.2 (`pip install osmium==4.3.2`):

```bash
python -c "import osmium, sys; w = osmium.SimpleWriter(sys.argv[2]); [w.add(o) for o in osmium.FileProcessor(sys.argv[1])]; w.close()" \
  goteborg-inom-vallgraven.osm goteborg-inom-vallgraven.osm.pbf
```

`osmium cat goteborg-inom-vallgraven.osm -o goteborg-inom-vallgraven.osm.pbf`
(osmium-tool) does the same with the same library. Delete the `.osm` afterwards;
only the PBF is committed. Then run the integration tests against tiles built
from the new file — the CI job does exactly that on the pull request that
carries it.
