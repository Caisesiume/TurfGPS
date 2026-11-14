# Test extract — central Göteborg

`goteborg-inom-vallgraven.osm.pbf` is a small OpenStreetMap extract that CI
builds Valhalla's tiles from, so that the Valhalla adapter's integration tests
(`../valhalla_integration_test.go`) run against a real engine, and loads into
the access-marking tables beside them, so that the access-marking check's
(`../../accessstore/integration_test.go`, #255) run against both. It is committed
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
| Area | 57.6980–57.7120 N, 11.9500–11.9850 E — Inom Vallgraven and the streets around it, about 1.6 km by 2.1 km (3.2 km²); and four small islands elsewhere, below |
| Content | every way tagged `highway`, every `type=restriction` relation, and the nodes and member ways they reference, which reach slightly past the box; and the access-marking cases below, with the nodes of their ways |
| Size | 196,273 bytes; SHA-256 `927c833ad8fbb35e002d1c7a0a8898eb165c381af2bdbcd85cbb33d8fa76e727` — 13,452 nodes, 3,862 ways, 57 relations |
| Data as of | the box: `osm_base` 2026-10-22T07:05:28Z; the cases: `osm_base` 2026-11-14T20:55:17Z, each as the Overpass API reported it |
| Retrieved | the box on 22 October 2026, the cases on 14 November 2026, from `https://overpass-api.de/api/interpreter` |
| Query | the box: `goteborg-inom-vallgraven.overpassql`; the cases: `access-marking-cases.overpassql`; both beside this file |

**The box alone** was 193,605 bytes, SHA-256
`0b878c989152be85784eea44c01e67d363140937629f27813caa0a3ac4f9aab9`; that file is
what the cases were merged into, and it is the "other extract" #255's AC5 test
names.

## The access-marking cases

#255's criteria name cases the box holds no instance of, and its acceptance
criteria let the extract be extended where it does not. Each was found by an
Overpass search, nearest Göteborg first, and added with every node its ways
reference, so a test reads it from the same file the tiles are built from:

| Case | OSM | Where | Taken with |
|---|---|---|---|
| a way carrying `access=military` | way 25683224 | Copenhagen, 55.7176 N 12.5742 E | the way alone |
| a way carrying `motor_vehicle=private` with `access=yes` | way 5179080, Klarabergsgatan | Stockholm, 59.3325 N 18.0637 E | the way alone |
| a `lift_gate` carrying `access=residents`, on the only road between two unmarked ones | node 8691377189, on way 444469850 (Frönäs sjöväg) | south of Stockholm, 58.8934 N 17.8009 E | the ways through the gate, and the `highway` ways meeting them |
| a way carrying `access=no` with `motor_vehicle=yes`, between unmarked ways | way 107135272, Ringvägen | Torslanda, Göteborg, 57.7168 N 11.8276 E | the way, and the `highway` ways meeting it |

The box already holds a way carrying `access=private` (way 976307072, among
others) and Viktoriagatan's block 32709123, carrying `vehicle:backward=no`.

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

Besides the tiles, the `valhalla` job applies `migrations/` to a PostGIS beside
the engine, loads this file into the access-marking tables with
`turfgps load-access-markings`, fails unless the load records the SHA-256 of the
file the tiles were built from, and reads the file a second time with
`osmium cat -f opl`, so the access-marking tests compare the tables with a
reading independent of the loader's.


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
(osmium-tool) does the same with the same library.

Then fetch the access-marking cases the same way, from
`access-marking-cases.overpassql`, into `access-marking-cases.osm`, and merge
them into the box's PBF. pyosmium 4.3.2's merge sorts and de-duplicates, and
gave the committed file byte for byte on two runs:

```bash
python -c "import osmium, sys; r = osmium.MergeInputReader(); [r.add_file(f) for f in sys.argv[1:-1]]; h = osmium.WriteHandler(sys.argv[-1]); r.apply(h); h.close()" \
  box.osm.pbf access-marking-cases.osm goteborg-inom-vallgraven.osm.pbf
```

where `box.osm.pbf` is the box converted as above, under that name. Delete the
`.osm` files and `box.osm.pbf` afterwards; only the merged PBF is committed.
The access-marking tests establish each case from the file before measuring
it, so a regeneration that loses one fails them rather than passing.

Then run the integration tests against tiles built
from the new file — the CI job does exactly that on the pull request that
carries it.
