# Test extract — central Göteborg

`goteborg-inom-vallgraven.osm.pbf` is a small OpenStreetMap extract that CI
builds Valhalla's tiles from, so that the Valhalla adapter's integration tests
(`../valhalla_integration_test.go`) run against a real engine, and loads into
the access-marking tables beside them, so that the access-marking check's
(`../../accessstore/integration_test.go`, #255) and the exclusion's
(`../../accessstore/exclusion_integration_test.go`, #248) run against both. It is committed
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

**Not everything in it is from OpenStreetMap.** 218 ways and 267 nodes are
synthetic, written for #255's, #248's and #271's tests: every object with an id of
9,000,000,000,000,000,000 or above. *The synthetic cases*, below, says which and
why.

## What it holds

| | |
|---|---|
| Area | 57.6980–57.7120 N, 11.9500–11.9850 E — Inom Vallgraven and the streets around it, about 1.6 km by 2.1 km (3.2 km²); seven small islands elsewhere, below; and the synthetic cases, in the Kattegat, 57.00–57.25 N |
| Content | every way tagged `highway`, every `type=restriction` relation, and the nodes and member ways they reference, which reach slightly past the box; the access-marking cases below, with the nodes of their ways; and the synthetic cases |
| Size | 200,112 bytes; SHA-256 `32d37537360f0ec1c4d4bd7dd0fa49059c36c20ef5589a3813a8f9f5558004a6` — 13,808 nodes, 4,095 ways, 57 relations |
| Data as of | the box: `osm_base` 2026-10-22T07:05:28Z; the cases: `osm_base` 2026-11-14T22:32:02Z, each as the Overpass API reported it. The synthetic cases are no OpenStreetMap data and have no such date |
| Retrieved | the box on 22 October 2026, the cases on 14 November 2026, from `https://overpass-api.de/api/interpreter` |
| Query | the box: `goteborg-inom-vallgraven.overpassql`; the cases: `access-marking-cases.overpassql`; both beside this file. The synthetic cases are `access-marking-synthetic.osm`, beside it too |

**The box alone** was 193,605 bytes, SHA-256
`0b878c989152be85784eea44c01e67d363140937629f27813caa0a3ac4f9aab9`; that file is
what the cases were merged into, and it is the "other extract" #255's AC5 test
names. The box and the access-marking cases without the synthetic ones were
197,799 bytes, SHA-256
`ab8f4dbac9e0468dbb01458c2f3c770c9f67984b40b3a652500bbf5a889fe728`; adding the
synthetic cases changed none of that file's objects, compared object by object.
With #255's synthetic cases and before #248's, the file was 198,078 bytes,
SHA-256 `35c74fb1aad95398138d731ecabf184243ffd06328b9b3e43529407896ab23f8`;
adding #248's changed none of its 17,500 objects and added 421, compared object
by object (PR #270). With #248's cases and before #271's, the file was 199,930
bytes, SHA-256
`6696a7087571a26f4e7611d8fbae20864bf86baeda14452dd74da70e33448113`; adding
#271's changed none of its 17,921 objects and added 39, compared object by
object (#271).

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
| a way carrying only `motor_vehicle:backward=no`, between unmarked ways (AC6, AC7) | way 14175890, Falkenbergsgatan | Kallebäck, Göteborg, 57.6809 N 11.9994 E | the way, and the `highway` ways meeting it |
| a node carrying `access=private` and no `barrier` key, on a road between unmarked ones (AC6, AC7) | node 312500257, on way 28443588 | Malmö, 55.6217 N 13.0519 E | the ways through the node, and the `highway` ways meeting them |
| a way carrying only `motor_vehicle:conditional=no @ (Mo-Fr 08:00-18:00)`, between unmarked ways (AC6, AC8) | way 4313847, Mornington Road | Bristol, 51.4685 N 2.6154 W | the way, and the `highway` ways meeting it |

The box already holds a way carrying `access=private` (way 976307072, among
others) and Viktoriagatan's block 32709123, carrying `vehicle:backward=no`.

## The synthetic cases — not from OpenStreetMap

Four of #255's clauses name tag sets that **OpenStreetMap holds nowhere**: an
Overpass search of the whole planet on 14 November 2026 (`osm_base`
22:05–22:32Z) found no way carrying any of them (PR #261). On the lead's ruling
on that PR, option (a), they are written by hand into
`access-marking-synthetic.osm`, beside this file, and merged into the extract,
so that each is tested at the level its criterion states, the route cases
through the engine:

- **Every id is 9,000,000,000,000,000,000 or above.** OpenStreetMap's ids are
  signed 64-bit integers handed out in sequence, and its highest were about
  1.4 × 10¹⁰ for nodes and 1.6 × 10⁹ for ways in this file; no real object can
  take one of these, and no future regeneration can overwrite one.
- **Every synthetic way carries** `note=synthetic, not from OpenStreetMap: TurfGPS #255 test case`,
  and the tests fail unless the way they measure carries it.
- **They lie in the Kattegat**, 57.00–57.25 N, 11.700–11.712 E, more than 50 km
  from every object from OpenStreetMap in the file, and share no node with one.
  The engine's location search gives up 35 km out, so a point near them cannot
  snap to a real road, nor a point on a real road to them. No route between
  real points changes.
- **Each route case is a road of three ways** running east, all
  `highway=residential`: an unmarked way, the case's way, and another unmarked
  way, so the route between points on the two outer ways runs along the case's
  way. The table-only cases are a way each, alone.

| Case | Synthetic way (id − 9×10¹⁸) | Nodes (id − 9×10¹⁸) | Where |
|---|---|---|---|
| a way carrying `motorcar=yes` with `motor_vehicle:backward=no`, between unmarked ways (AC6 third clause, AC7 third case) | 2, between unmarked ways 1 and 3 | 101–105; way 2 runs 102, 103, 104 | 57.000 N, 11.700–11.709 E |
| a way carrying `motor_vehicle=no` with `motor_vehicle:forward=yes` and `motor_vehicle:backward=yes`, between unmarked ways (AC9, its route) | 12, between unmarked ways 11 and 13 | 201–205; way 12 runs 202, 203, 204 | 57.010 N, 11.700–11.709 E |
| a way carrying `motorcar=no` with `motorcar:forward=yes` and `motorcar:backward=yes` (AC9) | 14 | 301, 302 | 57.020 N |
| a way carrying `motor_vehicle=no` with `motorcar:forward=yes` and `motorcar:backward=yes` (AC9) | 15 | 303, 304 | 57.025 N |
| a way carrying `access=private` with `motor_vehicle:conditional=yes @ (Mo-Fr 08:00-18:00)` (AC6 fifth clause) | 21 | 401, 402 | 57.030 N |

The nearest real instances of the same class, and why they were not used, are
in PR #261: there is none for AC9's, and AC6's and AC7's would have spelled the
criteria's tags differently.

### #248's cases

#248's criteria name a **road layout** as well as a tag set: a marking whose
quickest connection has a way round it, a home on a private road, a car park
of marked ways, a private road with a second way out through an estate, a gate
beyond the run. The extract held no instance of the layouts, and a real one
found by search would bring a neighbourhood of roads whose quickest
connection nothing here controls, so each is a small road network of its own.
They were written by a script for PR #270, and its output, the `#248` block of
`access-marking-synthetic.osm`, is what is committed and what a regeneration
keeps:

- **Case c lies along 57.050 + 0.006c ° N**, 11.700–11.712 E, 0.002° or more
  from the next; its node ids are 9×10¹⁸ + 10,000 + 100c + k and its way ids
  9×10¹⁸ + 1,000 + 20c + k. Every way carries
  `note=synthetic, not from OpenStreetMap: TurfGPS #248 test case`.
- **A transit case** (0–28) is a road of three `highway=residential` ways east,
  nodes at 11.700, 11.703, 11.7045, 11.706 and 11.709 E: an unmarked way, the
  case's way, an unmarked way. The case's tags are on the middle way, or on
  its middle node at 11.7045 E. Where the criterion needs a connection
  avoiding the marking, three unmarked ways run round it 0.002° to the north.
  Its ends are points at 11.7010 and 11.7080 E.

| Case | Tags | Round it | Criterion |
|---|---|---|---|
| 0–6 | `access=` `private`, `destination`, `customers`, `delivery`, `permit`, `residents`, `restricted` | yes | AC1 (0), AC2 |
| 7, 8 | `access=military`; `motor_vehicle=unknown` | yes | AC2 |
| 9–11 | `motorcar=private`, `motor_vehicle=private`, `vehicle=private`, each with `access=yes` | yes | AC2 |
| 12, 13 | node `barrier=lift_gate` `access=residents`; node `barrier=gate` `motor_vehicle=private` | yes | AC3 |
| 14 | node `barrier=gate` `access=private` `motorcar=yes` | no | AC14 |
| 15, 16 | `access=public`; `motor_vehicle=allowed` | no | AC15 |
| 17 | `access=no` | yes | AC16 |
| 18–20 | `access=private` `bicycle=yes` `foot=yes`; `motor_vehicle=private` `psv=yes` (a bus gate's); `motor_vehicle=destination` `motorcycle=yes` | yes | AC17 |
| 21 | node `barrier=gate` `access=private` `bicycle=yes` `foot=yes` | yes | AC17 |
| 22, 23 | only `motor_vehicle:backward=no`, drawn east; only `access:forward=private`, drawn west, so a leg east drives it backward | yes | AC18 |
| 24, 25 | node `access=private`; node `motor_vehicle=destination`; neither with a `barrier` key | yes | AC19 |
| 26, 27 | `motorcar=yes` `motor_vehicle:backward=no`, drawn east; `motor_vehicle=yes` `access:forward=private`, drawn west | yes | AC20 |
| 28 | `access=private` `motor_vehicle:conditional=yes @ (Mo-Fr 08:00-18:00)` | yes | AC22 |

- **Case 29** (AC4): an unmarked road east, with a private road (`access=private`)
  running 0.002° north from it at 11.703 E, a home at its middle, and a car park
  of two `highway=service` ways marked `access=customers`, north from it at
  11.706 E then east, a stop at 11.7066 E on the second.
- **Case 30** (AC6): a home on private way A, 0.0008° south of the north road it
  meets; A continues as private way B 0.002° further south to the south road,
  on which the destination lies 0.0005° east of B's end; the two roads are
  joined 0.009° east. The quickest leg from the home leaves by B.
- **Cases 31 and 32** (AC12): a home on a private road north off an unmarked
  road, on which a node `barrier=lift_gate` `access=residents` lies 0.004° east
  of the junction, before the destination; case 31 has three unmarked ways
  round the gate, 0.002° to the south, and case 32 none.

### #271's cases

#271's criteria name two layouts the file held no instance of — a marking
every connection passes with a second one further on that a way round avoids,
and two connections each through a marking of its own — and a warned leg past
both a marked way and a marked barrier, which #248's cases name only along a
run. They were written by a script for #271 and are the `#271` block of
`access-marking-synthetic.osm`; their ids are a range of their own:

- **Layout j lies along 57.034 + 0.005j ° N**, 11.700–11.712 E, between #255's
  last case (57.030 N) and #248's first (57.050 N), so the band of *The
  synthetic cases* holds; its node ids are 9×10¹⁸ + 20,000 + 100j + k and its
  way ids 9×10¹⁸ + 3,000 + 20j + k. Every way is `highway=residential` and
  carries `note=synthetic, not from OpenStreetMap: TurfGPS #271 test case`.
- **Every gate is a node `barrier=lift_gate` `access=residents`**, the tagging
  of #248's AC3 and AC12 cases, which the engine routes through where nothing
  goes round it (#248's AC12, second case). **The marked way is tagged
  `motorcar=yes` `motor_vehicle:backward=no`, drawn east**, which the check
  marks in both directions (`RD-065`) and the engine drives east, as #255's
  case on way 2 shows, so the engine's own first answer runs along it and the
  warned leg's search for less, not the engine's own avoidance, is what keeps
  it off the leg.

| Layout | Ways (k) | Gate (node k) | Round it | Criterion |
|---|---|---|---|---|
| j=0 | 1, 2 (through the gate), 3 (the marked way, 11.706–11.709 E), 4 | 3, at 11.7045 E, on way 2 | ways 5–7, 0.002° north of way 3; nothing round the gate | AC4 |
| j=1 | 1, 2 (through gate 3), 3; ways 4–6, 0.002° north, through gate 7 | 3 and 7, each at 11.7045 E | each connection is the way round the other's gate | AC5 |
| j=2 | 1, 2 (the marked way, 11.703–11.706 E), 3 (through the gate), 4 | 4, at 11.7075 E, on way 3 | nothing | AC6 |

The ends of each are points at 11.7010 E and, on the last way, 11.7080 E
(j=1) or 11.7100 E (j=0, j=2).

Each test establishes its case from the file before measuring it, and asks the
engine, with `ignore_access`, whether the quickest connection between the ends
runs along the marking, so a regeneration that loses a case fails them.

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
them and the synthetic cases into the box's PBF. pyosmium 4.3.2's merge sorts
and de-duplicates, and gave the committed file byte for byte on two runs (#248's
cases were merged the same way into the file before them, which holds the box
and #255's cases, and #271's into the file holding #248's, each with the same
result on two runs):

```bash
python -c "import osmium, sys; r = osmium.MergeInputReader(); [r.add_file(f) for f in sys.argv[1:-1]]; h = osmium.WriteHandler(sys.argv[-1]); r.apply(h); h.close()" \
  box.osm.pbf access-marking-cases.osm access-marking-synthetic.osm goteborg-inom-vallgraven.osm.pbf
```

where `box.osm.pbf` is the box converted as above, under that name. Delete
`access-marking-cases.osm`, the box's `.osm` and `box.osm.pbf` afterwards; the
merged PBF and `access-marking-synthetic.osm`, which is written by hand and
fetched from nowhere, are what is committed. A regeneration keeps the synthetic
file as it is.
The access-marking tests establish each case from the file before measuring
it, so a regeneration that loses one fails them rather than passing.

Then run the integration tests against tiles built
from the new file — the CI job does exactly that on the pull request that
carries it.
