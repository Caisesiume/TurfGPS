# TurfGPS

A route-planning and decision-support system for players of [Turf](https://turfgame.com), the GPS location game.

You are already driving somewhere — Örebro to Jönköping, say. TurfGPS works out which roads and which individual zones make the best Turf journey for the extra time you can spare, and explains why one journey beats another.

It is not a zone map, and it is not a navigation app. It is an optimizer with an explanation layer.

## Status

**Pre-implementation.** The documentation leads the code.

The repository previously held an unrelated Next.js zone-map prototype. It was removed on 31 July 2026, once the product definition made clear it was not a partial implementation of this system but a different application. Nothing is being ported from it; it remains in git history.

## Start here

| Document | The question it answers |
|----------|------------------------|
| [`docs/README.md`](docs/README.md) | How is the documentation organised? |
| [`docs/SPECIFICATION.md`](docs/SPECIFICATION.md) | What is the product, and how must it behave? |
| [`docs/CalculationSpecification.md`](docs/CalculationSpecification.md) | How is every number worked out? |
| [`docs/Architecture.md`](docs/Architecture.md) | How will it be built? |
| [`docs/DESIGN.md`](docs/DESIGN.md) | What is using it like? |
| [`docs/DELIVERY.md`](docs/DELIVERY.md) | How does work get tracked and reviewed? |

`docs/SPECIFICATION.md` is the source of truth for intent. Read it before writing code — it settles a great many decisions, and re-deciding one that is already settled wastes the work that produced it.

## Licence

See [LICENSE](LICENSE).
