Before or alongside the architecture document, I would create a smaller document that formalizes the domain model and calculations.

This would cover:

- Route
- Route alternative
- Route leg
- Mandatory waypoint
- Zone
- Zone attribute
- Candidate zone
- Stop location
- Direct road-access zone
- Park-and-walk zone
- Access path
- Attribute preference
- Time budget
- Stretch alternative
- Accessibility confidence
- Takeover duration
- Terrain cost
- Route score

It should also formalize formulas such as:

- Walking time
- Elevation-adjusted walking time
- Stop time
- Additional journey time
- Shared-detour cost
- Route utility
- Constraint handling
- Attribute-weight conversion

This is important because the optimization logic is central enough that it should not be buried inside the architecture document.