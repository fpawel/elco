PRAGMA foreign_keys = ON;
PRAGMA encoding = 'UTF-8';
--E:\Program Data\Аналитприбор\elco\elco.sqlite
--C:\Users\fpawel\AppData\Roaming\Аналитприбор\elco\elco.sqlite

CREATE TABLE IF NOT EXISTS gas
(
  gas_name TEXT PRIMARY KEY NOT NULL,
  code     INTEGER UNIQUE   NOT NULL
);

CREATE TABLE IF NOT EXISTS units
(
  units_name TEXT PRIMARY KEY NOT NULL,
  code       INTEGER UNIQUE   NOT NULL
);

CREATE TABLE IF NOT EXISTS product_type
(
  product_type_name   TEXT PRIMARY KEY NOT NULL,
  gas_name            TEXT             NOT NULL,
  units_name          TEXT             NOT NULL,
  scale               REAL             NOT NULL,
  noble_metal_content REAL             NOT NULL,
  lifetime_months     INTEGER          NOT NULL CHECK (lifetime_months > 0),
  FOREIGN KEY (gas_name) REFERENCES gas (gas_name),
  FOREIGN KEY (units_name) REFERENCES units (units_name)
);

CREATE TABLE IF NOT EXISTS party
(
  party_id           INTEGER PRIMARY KEY NOT NULL,
  old_party_id       TEXT,
  created_at         TIMESTAMP           NOT NULL DEFAULT (datetime('now')),
  updated_at         TIMESTAMP           NOT NULL DEFAULT (datetime('now')),
  product_type_name  TEXT                NOT NULL DEFAULT '035',
  concentration1     REAL                NOT NULL DEFAULT 0 CHECK (concentration1 >= 0),
  concentration2     REAL                NOT NULL DEFAULT 50 CHECK (concentration2 >= 0),
  concentration3     REAL                NOT NULL DEFAULT 100 CHECK (concentration3 >= 0),
  note               TEXT,
  points_method      INTEGER             NOT NULL
    CHECK (points_method IN (2, 3))               DEFAULT 2,
  min_fon            REAL                         DEFAULT -1,
  max_fon            REAL                         DEFAULT 2,
  max_d_fon          REAL                         DEFAULT 3,
  min_k_sens20       REAL                         DEFAULT 0.08,
  max_k_sens20       REAL                         DEFAULT 0.335,
  min_k_sens50       REAL                         DEFAULT 110,
  max_k_sens50       REAL                         DEFAULT 150,
  min_d_temp         REAL                         DEFAULT 0,
  max_d_temp         REAL                         DEFAULT 3,
  max_d_not_measured REAL                         DEFAULT 5,

  FOREIGN KEY (product_type_name) REFERENCES product_type (product_type_name)
);

CREATE TABLE IF NOT EXISTS product
(
  product_id        INTEGER PRIMARY KEY NOT NULL,
  party_id          INTEGER             NOT NULL,
  serial            INTEGER
    CHECK ( serial ISNULL OR serial > 0 ),
  place             INTEGER             NOT NULL
    CHECK (place >= 0),
  product_type_name TEXT,
  note              TEXT,

  i_f_minus20       REAL,
  i_f_plus20        REAL,
  i_f_plus50        REAL,

  i_s_minus20       REAL,
  i_s_plus20        REAL,
  i_s_plus50        REAL,

  i13               REAL,
  i24               REAL,
  i35               REAL,
  i26               REAL,
  i17               REAL,
  not_measured      REAL,
  firmware          BLOB,
  production        BOOLEAN             NOT NULL
    CHECK (production IN (0, 1)) DEFAULT 0,

  points_method     INTEGER
    CHECK (points_method IN (2, 3)),

  old_product_id    TEXT,
  old_serial        INTEGER,

  CONSTRAINT unique_party_place UNIQUE (party_id, place),
  CONSTRAINT unique_party_serial UNIQUE (party_id, serial),

  FOREIGN KEY (product_type_name) REFERENCES product_type (product_type_name),
  FOREIGN KEY (party_id) REFERENCES party (party_id)
    ON DELETE CASCADE
);

CREATE TRIGGER IF NOT EXISTS trigger_product_party_updated_at
  AFTER INSERT
  ON product
  BEGIN
    UPDATE party
    SET updated_at = datetime('now')
    WHERE party.party_id = new.party_id;
  END;

CREATE VIEW IF NOT EXISTS party_info AS
SELECT *,
       cast(strftime('%Y', DATETIME(created_at, '+3 hours')) AS INTEGER) AS year,
       cast(strftime('%m', DATETIME(created_at, '+3 hours')) AS INTEGER) AS month,
       cast(strftime('%d', DATETIME(created_at, '+3 hours')) AS INTEGER) AS day,
       party_id IN (SELECT party_id FROM last_party)                     AS last
FROM party;


DROP VIEW  IF EXISTS product_info_1;
CREATE VIEW IF NOT EXISTS product_info_1 AS
SELECT product_id,
       product.party_id,
       created_at,
       serial,
       place,
       production,
       (CASE (product.product_type_name ISNULL)
          WHEN 1 THEN party.product_type_name
          WHEN 0
            THEN product.product_type_name END)                                AS applied_product_type_name,
       product.product_type_name                                               AS product_type_name,
       product.note                                                            AS note,
       round(i_f_minus20, 3)                                                   AS i_f_minus20,
       round(i_f_plus20, 3)                                                    AS i_f_plus20,
       round(i_f_plus50, 3)                                                    AS i_f_plus50,
       round(i_s_minus20, 3)                                                   AS i_s_minus20,
       round(i_s_plus20, 3)                                                    AS i_s_plus20,
       round(i_s_plus50, 3)                                                    AS i_s_plus50,
       round(i13, 3)                                                           AS i13,
       round(i24, 3)                                                           AS i24,
       round(i35, 3)                                                           AS i35,
       round(i26, 3)                                                           AS i26,
       round(i17, 3)                                                           AS i17,
       round(not_measured, 3)                                                  AS not_measured,

       round(i26 - i24, 3)                                                  AS variation,

       round(100 * (i_s_plus50 - i_f_plus50) / (i_s_plus20 - i_f_plus20), 3)   AS k_sens50,
       round(100 * (i_s_minus20 - i_f_minus20) / (i_s_plus20 - i_f_plus20), 3) AS k_sens_minus20,

       round((i_s_plus20 - i_f_plus20) / (concentration3 - concentration1), 3) AS k_sens20,
       round(i13 - i_f_plus20, 3)                                              AS d_fon20,
       round(i_f_plus50 - i_f_plus20, 3)                                       AS d_fon50,
       round(not_measured - i_f_plus20, 3)                                     AS d_not_measured,
       (firmware NOT NULL AND LENGTH(firmware) > 0)                            AS has_firmware,

       round(min_fon, 3)                                                       AS min_fon,
       round(max_fon, 3)                                                       AS max_fon,
       round(max_d_fon, 3)                                                     AS max_d_fon,
       round(min_k_sens20, 3)                                                  AS min_k_sens20,
       round(max_k_sens20, 3)                                                  AS max_k_sens20,
       round(min_d_temp, 3)                                                    AS min_d_temp,
       round(max_d_temp, 3)                                                    AS max_d_temp,
       round(min_k_sens50, 3)                                                  AS min_k_sens50,
       round(max_k_sens50, 3)                                                  AS max_k_sens50,
       round(max_d_not_measured, 3)                                            AS max_d_not_measured,
       product.points_method                                                   AS points_method,
       (CASE (product.points_method ISNULL)
          WHEN 1 THEN party.points_method
          WHEN 0
            THEN product.points_method END)                                    AS applied_points_method

FROM product
       INNER JOIN party ON party.party_id = product.party_id;

DROP VIEW  IF EXISTS product_info_2;
CREATE VIEW IF NOT EXISTS product_info_2 AS
SELECT q.*,

       max_d_temp ISNULL OR (d_fon50 NOTNULL) AND abs(d_fon50) < max_d_temp  AS ok_d_fon50,

       max_d_fon ISNULL OR (d_fon20 NOTNULL) AND abs(d_fon20) < max_d_fon    AS ok_d_fon20,

       min_k_sens20 ISNULL OR (k_sens20 NOTNULL) AND k_sens20 > min_k_sens20 AS ok_min_k_sens20,

       max_k_sens20 ISNULL OR (k_sens20 NOTNULL) AND k_sens20 < max_k_sens20 AS ok_max_k_sens20,

       min_k_sens50 ISNULL OR (k_sens50 NOTNULL) AND k_sens50 > min_k_sens50 AS ok_min_k_sens50,

       max_k_sens50 ISNULL OR (k_sens50 NOTNULL) AND k_sens50 < max_k_sens50 AS ok_max_k_sens50,


       min_fon ISNULL OR (i_f_plus20 NOTNULL) AND i_f_plus20 > min_fon       AS ok_min_fon20,
       max_fon ISNULL OR (i_f_plus20 NOTNULL) AND i_f_plus20 < max_fon       AS ok_max_fon20,

       min_fon ISNULL OR (i13 NOTNULL) AND i13 > min_fon                     AS ok_min_fon20_2,
       max_fon ISNULL OR (i13 NOTNULL) AND i13 < max_fon                     AS ok_max_fon20_2,

       max_d_not_measured ISNULL OR
       (d_not_measured NOTNULL) AND abs(d_not_measured) < max_d_not_measured AS ok_d_not_measured,

       gas.code                                                              AS gas_code,
       units.code                                                            AS units_code,
       gas.gas_name,
       units.units_name,
       scale,
       noble_metal_content,
       lifetime_months
FROM product_info_1 q
       INNER JOIN product_type ON product_type.product_type_name = q.applied_product_type_name
       INNER JOIN gas ON product_type.gas_name = gas.gas_name
       INNER JOIN units ON product_type.units_name = units.units_name;

DROP VIEW  IF EXISTS product_info;
CREATE VIEW IF NOT EXISTS product_info AS
SELECT *,
       ok_d_not_measured AND
       ok_d_fon50 AND ok_d_fon20 AND
       ok_min_k_sens20 AND ok_max_k_sens20 AND
       ok_min_k_sens50 AND ok_max_k_sens50 AND
       ok_min_fon20 AND ok_max_fon20 AND
       ok_min_fon20_2 AND ok_max_fon20_2 AS ok
FROM product_info_2 q;





INSERT
  OR
  IGNORE
INTO units (units_name, code)
VALUES ('мг/м3', 2),
       ('ppm', 3),
       ('об. дол. %', 7),
       ('млн-1', 5);

INSERT
  OR
  IGNORE
INTO gas (gas_name, code)
VALUES ('CO', 0x11),
       ('H₂S', 0x22),
       ('NH₃', 0x33),
       ('Cl₂', 0x44),
       ('SO₂', 0x55),
       ('NO₂', 0x66),
       ('O₂', 0x88),
       ('NO', 0x99),
       ('HCl', 0xAA),
       ('N₂O₄', 0xBB);

INSERT
  OR
  IGNORE
INTO product_type (product_type_name,
                   gas_name,
                   units_name,
                   scale,
                   noble_metal_content,
                   lifetime_months)
VALUES ('010-15', 'O₂', 'об. дол. %', 30, 0, 12),
       ('010-18', 'O₂', 'об. дол. %', 30, 0, 12),
       ('035', 'CO', 'мг/м3', 200, 0.1456, 12),
       ('035.2', 'CO', 'мг/м3', 200, 0.1626, 18),
       ('035-10', 'H₂S', 'мг/м3', 40, 0, 12),
       ('035-100', 'NO₂', 'об. дол. %', 0.014, 0, 12),
       ('035-102', 'H₂S', 'мг/м3', 40, 0, 12),
       ('035-103', 'Cl₂', 'мг/м3', 25, 0, 12),
       ('035-105', 'CO', 'мг/м3', 200, 0.1456, 12),
       ('035-111', 'CO', 'мг/м3', 200, 0.1626, 12),
       ('035-113', 'H₂S', 'мг/м3', 40, 0, 12),
       ('035-114', 'SO₂', 'мг/м3', 20, 0, 12),
       ('035-115', 'Cl₂', 'мг/м3', 25, 0, 12),
       ('035-116', 'Cl₂', 'мг/м3', 50, 0, 12),
       ('035-117', 'NO₂', 'мг/м3', 10, 0, 12),
       ('035-118', 'HCl', 'мг/м3', 30, 0, 12),
       ('035-128', 'SO₂', 'мг/м3', 40, 0, 12),
       ('035-129', 'SO₂', 'мг/м3', 200, 0, 12),
       ('035-130', 'SO₂', 'мг/м3', 3000, 0, 12),
       ('035-131', 'NO₂', 'мг/м3', 100, 0, 12),
       ('035-132', 'NO₂', 'мг/м3', 200, 0, 12),
       ('035-133', 'NO₂', 'мг/м3', 500, 0, 12),
       ('035-134', 'NO₂', 'мг/м3', 3000, 0, 12),
       ('035-21', 'CO', 'мг/м3', 200, 0.1456, 12),
       ('035-40', 'CO', 'мг/м3', 200, 0.1456, 12),
       ('035-52', 'Cl₂', 'мг/м3', 25, 0, 12),
       ('035-54', 'SO₂', 'мг/м3', 20, 0, 12),
       ('035-55', 'NO₂', 'мг/м3', 10, 0, 12),
       ('035-59', 'CO', 'об. дол. %', 0.5, 0.1891, 12),
       ('035-60', 'CO', 'мг/м3', 200, 0.1891, 12),
       ('035-61', 'CO', 'ppm', 2000, 0.1891, 12),
       ('035-62', 'H₂S', 'мг/м3', 40, 0, 12),
       ('035-63', 'Cl₂', 'мг/м3', 25, 0, 12),
       ('035-65', 'SO₂', 'ppm', 200, 0, 12),
       ('035-66', 'SO₂', 'ppm', 3000, 0, 12),
       ('035-69', 'NO₂', 'ppm', 140, 0, 12),
       ('035-69.2', 'NO₂', 'об. дол. %', 0.014, 0, 12),
       ('035-70', 'CO', 'мг/м3', 200, 0.1626, 12),
       ('035-75', 'NO₂', 'мг/м3', 10, 0, 12),
       ('035-80', 'CO', 'мг/м3', 200, 0.1456, 12),
       ('035-81', 'CO', 'мг/м3', 1500, 0.1456, 12),
       ('035-82', 'H₂S', 'мг/м3', 40, 0, 12),
       ('035-83', 'SO₂', 'мг/м3', 20, 0, 12),
       ('035-84', 'Cl₂', 'мг/м3', 25, 0, 12),
       ('035-87', 'NO₂', 'мг/м3', 10, 0, 12),
       ('035-89', 'Cl₂', 'мг/м3', 50, 0, 12),
       ('035-92', 'CO', 'об. дол. %', 0.5, 0.1891, 12),
       ('035-93', 'CO', 'млн-1', 200, 0.1891, 12),
       ('035-94', 'CO', 'млн-1', 2000, 0.1891, 12),
       ('035-95', 'SO₂', 'ppm', 200, 0, 12),
       ('035-96', 'SO₂', 'ppm', 20, 0, 12),
       ('035-99', 'NO₂', 'ppm', 140, 0, 12),
       ('060-10', 'NH₃', 'мг/м3', 150, 0, 12),
       ('060-11', 'NH₃', 'мг/м3', 150, 0, 12),
       ('060-12', 'NH₃', 'мг/м3', 2000, 0, 12),
       ('060-15', 'NH₃', 'мг/м3', 600, 0, 12),
       ('060-16', 'NH₃', 'мг/м3', 600, 0, 12),
       ('060-17', 'NH₃', 'мг/м3', 2000, 0, 12),
       ('060-20', 'NH₃', 'мг/м3', 150, 0, 12),
       ('060-31', 'NH₃', 'мг/м3', 150, 0, 12),
       ('060-32', 'NH₃', 'мг/м3', 2000, 0, 12),
       ('060-33', 'NH₃', 'мг/м3', 600, 0, 12),
       ('060-34', 'NH₃', 'мг/м3', 2000, 0, 12),
       ('100', 'CO', 'мг/м3', 200, 0.0816, 12),
       ('100-01', 'H₂S', 'мг/м3', 40, 0, 12),
       ('100-02', 'H₂S', 'мг/м3', 20, 0, 12),
       ('100-03', 'SO₂', 'мг/м3', 20, 0, 12),
       ('100-04', 'NO₂', 'мг/м3', 10, 0, 12),
       ('100-05', 'CO', 'мг/м3', 50, 0.0816, 12),
       ('100-10', 'CO', 'мг/м3', 200, 0.0816, 12),
       ('100-11', 'H₂S', 'мг/м3', 40, 0, 12),
       ('100-12', 'H₂S', 'мг/м3', 20, 0, 12),
       ('100-13', 'SO₂', 'мг/м3', 20, 0, 12),
       ('100-14', 'NO₂', 'мг/м3', 10, 0, 12),
       ('100-15', 'CO', 'мг/м3', 50, 0.0816, 12),
       ('100-16', 'Cl₂', 'мг/м3', 25, 0, 12),
       ('100-17', 'HCl', 'мг/м3', 25, 0, 12),
       ('130-01', 'CO', 'мг/м3', 200, 0.1626, 12),
       ('130-08', 'CO', 'ppm', 100, 0.1162, 12);


DELETE
FROM party
WHERE NOT EXISTS(SELECT product_id FROM product WHERE party.party_id = product.party_id);
