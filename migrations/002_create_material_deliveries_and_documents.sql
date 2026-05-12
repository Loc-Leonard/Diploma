CREATE TABLE IF NOT EXISTS material_deliveries (
    id              SERIAL PRIMARY KEY,
    object_id       BIGINT NOT NULL,
    work_item_id    BIGINT NULL,
    foreman_id      BIGINT NOT NULL,
    date            TIMESTAMPTZ NOT NULL,
    material        TEXT NOT NULL DEFAULT '',
    qty             DOUBLE PRECISION NOT NULL DEFAULT 0,
    unit            TEXT NOT NULL DEFAULT '',
    document_number TEXT NOT NULL DEFAULT '',
    source          TEXT NOT NULL DEFAULT 'MANUAL',
    cv_confidence   DOUBLE PRECISION NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_material_deliveries_object_id ON material_deliveries(object_id);
CREATE INDEX IF NOT EXISTS idx_material_deliveries_foreman_id ON material_deliveries(foreman_id);
CREATE INDEX IF NOT EXISTS idx_material_deliveries_work_item_id ON material_deliveries(work_item_id);

CREATE TABLE IF NOT EXISTS material_documents (
    id                 SERIAL PRIMARY KEY,
    delivery_id        BIGINT NOT NULL,
    document_type      TEXT NOT NULL DEFAULT 'TTN',
    storage_path       TEXT NOT NULL,
    original_file_name TEXT NOT NULL,
    mime_type          TEXT NOT NULL DEFAULT '',
    cv_status          TEXT NOT NULL DEFAULT 'DONE',
    cv_payload_json    TEXT NOT NULL DEFAULT '',
    created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_material_documents_delivery_id ON material_documents(delivery_id);
