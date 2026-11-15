/* Extends the central-Göteborg extract with the area features of its box that
   carry any of the twelve access keys (#253): every closed way without a
   `highway` key, every closed way tagged `area=yes`, and every
   `type=multipolygon` relation, with the member ways and nodes they reference.
   README.md beside this file says which and why. */
[out:xml][timeout:120][bbox:57.6980,11.9500,57.7120,11.9850];
(
  way[!"highway"][~"^(motorcar|motor_vehicle|vehicle|access)(:forward|:backward)?$"~"."](if:is_closed());
  way["area"="yes"][~"^(motorcar|motor_vehicle|vehicle|access)(:forward|:backward)?$"~"."](if:is_closed());
  relation["type"="multipolygon"][~"^(motorcar|motor_vehicle|vehicle|access)(:forward|:backward)?$"~"."];
);
(._; >;);
out body;
