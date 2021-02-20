CREATE TABLE IF NOT EXISTS measurements (
  time            TIMESTAMPTZ       NOT NULL,
  battery_voltage DOUBLE PRECISION  NULL,
  soc             DOUBLE PRECISION  NULL,
  /* "in" is a reserved word */
  in_value        DOUBLE PRECISION  NULL,
  out_value       DOUBLE PRECISION  NULL,
  charge          DOUBLE PRECISION  NULL,
  load            DOUBLE PRECISION  NULL,
  regulator_state DOUBLE PRECISION  NULL
);
