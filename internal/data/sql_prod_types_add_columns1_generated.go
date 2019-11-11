package data

const SQLProdTypesAddColumns1 = `
ALTER TABLE product_type
    ADD k_sens20 REAL DEFAULT NULL;
ALTER TABLE product_type
    ADD points_method INTEGER NOT NULL CHECK (points_method IN (2, 3)) DEFAULT 2;
ALTER TABLE product_type
    ADD min_fon REAL DEFAULT NULL;
ALTER TABLE product_type
    ADD max_fon REAL DEFAULT NULL;
ALTER TABLE product_type
    ADD max_d_fon REAL DEFAULT NULL;
ALTER TABLE product_type
    ADD min_k_sens20 REAL DEFAULT NULL;
ALTER TABLE product_type
    ADD max_k_sens20 REAL DEFAULT NULL;
ALTER TABLE product_type
    ADD min_k_sens50 REAL DEFAULT NULL;
ALTER TABLE product_type
    ADD max_k_sens50 REAL DEFAULT NULL;
ALTER TABLE product_type
    ADD min_d_temp REAL DEFAULT NULL;
ALTER TABLE product_type
    ADD max_d_temp REAL DEFAULT NULL;
ALTER TABLE product_type
    ADD max_d_not_measured REAL DEFAULT NULL;
`
