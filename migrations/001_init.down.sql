DROP TABLE IF EXISTS audit_events;
DROP TRIGGER IF EXISTS trg_items_updated_at ON items;
DROP FUNCTION IF EXISTS update_updated_at_column();
DROP TABLE IF EXISTS items;
