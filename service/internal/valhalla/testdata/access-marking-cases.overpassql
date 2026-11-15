/* Extends the central-Göteborg extract with the access-marking cases it holds
   no instance of (#255): README.md beside this file says which and why. */
[out:xml][timeout:120];
node(8691377189)->.gate;
way(bn.gate)["highway"]->.gate_ways;
node(w.gate_ways)->.gate_nodes;
way(bn.gate_nodes)["highway"]->.gate_ring;
way(107135272)->.permitted;
node(w.permitted)->.permitted_nodes;
way(bn.permitted_nodes)["highway"]->.permitted_ring;
way(14175890)->.backward;
node(w.backward)->.backward_nodes;
way(bn.backward_nodes)["highway"]->.backward_ring;
node(312500257)->.keyless;
way(bn.keyless)["highway"]->.keyless_ways;
node(w.keyless_ways)->.keyless_nodes;
way(bn.keyless_nodes)["highway"]->.keyless_ring;
way(4313847)->.conditional;
node(w.conditional)->.conditional_nodes;
way(bn.conditional_nodes)["highway"]->.conditional_ring;
(
  way(25683224);
  way(5179080);
  .gate_ways;
  .gate_ring;
  .permitted;
  .permitted_ring;
  .backward;
  .backward_ring;
  .keyless_ways;
  .keyless_ring;
  .conditional;
  .conditional_ring;
);
(._; >;);
out body;
